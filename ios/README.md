# Double A — iOS app

A native SwiftUI app matching the web store: Double A branding (navy + crimson), Arabic/English with
RTL, light/dark theme, and the core flow — browse packages, **activate a code on the device**, and
**install the app via OTA** (`itms-services`).

Source lives in `ios/DoubleA/`. The old `ios-app/` folder was a non-functional skeleton and is
superseded by this.

## Open it in Xcode (2 minutes — reliable path)

The repo intentionally does **not** ship a hand-written `.xcodeproj` (they corrupt easily). Create the
project shell in Xcode, then drop the source in:

1. **Xcode → File → New → Project → iOS → App.**
   - Product Name: `DoubleA`
   - Interface: **SwiftUI**, Language: **Swift**
   - Save it anywhere (e.g. `ios/`).
2. In the new project, **delete** the two generated files `DoubleAApp.swift` and `ContentView.swift`
   (Move to Trash).
3. **Drag the files from `ios/DoubleA/` into the Xcode project** (all the `.swift` files). In the
   dialog: check **Copy items if needed** and **Create groups**, and add them to the `DoubleA` target.
4. Build & run (⌘R) on a simulator or device.

> No `Info.plist` edits are needed — it's a SwiftUI-lifecycle app, and RTL/dark are handled in code.

## Configure before shipping

In `AppState.swift`:
- `apiBaseURL` → your API origin, e.g. `https://api.2a-plus.com` (no trailing `/v1`).
- `storeSlug` → `store` (matches `STORE_TENANT_SLUG`).
- `whatsappNumber` → your WhatsApp number (digits only).

## Notes

- **The install button needs the production prerequisites** (HTTPS API + a signed IPA uploaded +
  Apple certificate). On a device it opens `itms-services://…manifest.plist` to install the signed
  container app.
- Language + theme toggles are in the **Settings** tab and persist across launches.
- Build/sign in the cloud (no Mac needed) with **Codemagic** or a **GitHub Actions macOS runner** once
  you have an Apple certificate.

## Screens

- **Home** — packages from `/shop/products` with live currency (EGP/KWD).
- **Activate** — check code (`/shop/activation`) → activate on device (`/v1/activation/validate`) →
  install (`/shop/app` → OTA manifest).
- **Settings** — language (ع/EN), theme (light/dark), contact, about.
