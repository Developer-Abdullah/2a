# Multi-Tenant IPA SaaS Platform
Enterprise-grade iOS application distribution platform.
## Architecture
- Backend: Go 1.22, Gin, PostgreSQL 16, Redis 7
- Frontend: Next.js 14, Tailwind CSS, NextAuth
- Mobile: iOS 16+, SwiftUI
## Local Development
1. Copy `.env.example` to `.env`.
2. Run `docker compose up --build -d` or `make dev`.
3. Run `docker compose --profile tools run --rm --build seed` or `make seed` to create the default tenant and platform admin.
4. Access Core API at `:8080`, Admin at `:3000`.
