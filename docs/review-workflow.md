# Manual Review Workflow

Every uploaded version must be **approved by the platform owner** before it can be signed. This is an
anti-piracy control: there is no code path from upload to a signed build that skips review.

## State machine

`application_versions.review_status`:

```
        upload
          │
          ▼
   ┌──────────────┐   approve   ┌──────────┐   (enqueues signing job)
   │ pending_review│ ─────────▶ │ approved │ ─────────▶ signing pipeline
   └──────────────┘             └──────────┘
          │
          │ reject (reason required)
          ▼
     ┌──────────┐
     │ rejected │   (terminal; merchant resubmits as a new version)
     └──────────┘
```

`approved` and `rejected` are terminal.

## Enforcement (two layers)

1. **Service layer** — approve/reject are conditional updates (`WHERE review_status =
   'pending_review'`). A no-op (0 rows) returns `ErrInvalidReviewTransition` → HTTP 409, so a version
   that is already reviewed cannot be re-transitioned.
2. **Database layer** — a `BEFORE UPDATE` trigger (`enforce_version_review_transition`, migration
   017) rejects any change of `review_status` that isn't `pending_review → approved/rejected`. Updates
   that leave `review_status` unchanged (e.g. the signing worker advancing `signing_status`) pass
   through untouched. `domain.ReviewStatus.CanTransitionTo` mirrors this logic and is unit-tested.

## Endpoints (owner-only, `RequirePlatformOwner`)

| Method & path | Action |
|---------------|--------|
| `GET /v1/admin/platform/reviews/pending` | Cross-tenant queue of `pending_review` versions. |
| `POST /v1/admin/platform/reviews/:slug/:versionId/approve` | Approve → enqueue signing job. |
| `POST /v1/admin/platform/reviews/:slug/:versionId/reject` | Reject with required `reason`. |

## Notifications & audit

Every approve/reject writes a row to the tenant's tamper-proof `audit_logs`
(`action = version.approved | version.rejected`, reviewer id + reason in `new_value`). The merchant
sees the outcome and any rejection reason via the version list (`review_status`, `rejection_reason`).
