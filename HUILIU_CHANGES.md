# Changes in this fork

This repository is a modified version of [QuantumNous/new-api](https://github.com/QuantumNous/new-api)
(AGPL-3.0, with the additional terms in [NOTICE](./NOTICE)). It is based on tag `v1.0.0-rc.51`.
The original license, NOTICE, copyright headers and Go module paths are kept unchanged.

Frontend design and development by New API contributors. The upstream project is at
https://github.com/QuantumNous/new-api . The modifications listed below were made by
the operator of this fork, starting 2026-11-26. The About page of the running site
links to the exact source commit it was built from.

## 2026-11-26

Build
- `.github/workflows/huiliu-image.yml`: builds `linux/amd64` and publishes `ghcr.io/bananawong/huiliu-api`.
  The image version is `huiliu-<date>-<commit>`.

Server
- `router/web-router.go`: the page title and description come from the SystemName option
  (`__SITE_NAME__` placeholder in `web/index.html`); `/favicon.ico` and `/logo.png` serve the
  Logo option (data URI or URL) and fall back to the bundled files.
- `middleware/cors.go`: the version response header is no longer sent.
- `common/constants.go`, `service/http.go`: the request id header is `X-Request-Id`; the old
  header name from upstream relays is still captured and not passed through.
- `relaykit/types/error.go`, `controller/*`, `middleware/utils.go`, `service/error.go`: the error
  type in relay error bodies is `api_error` (tests updated).
- `main.go`: the analytics placeholders no longer leave marker comments in the page.

Frontend
- `web/src/huiliu/`: fork-specific code (auth page frame, license notice, source links).
- `web/src/features/about/index.tsx`: the About page always ends with the license notice
  (attribution, upstream link, source of this build, license); the empty state no longer
  shows the upstream repository.
- `web/src/features/auth/auth-layout.tsx`: auth pages use the two-column frame from `web/src/huiliu/`.
- `web/src/styles/theme-presets.css`, `web/src/lib/theme-customization.ts`: new `huiliu` (Teal Ink)
  palette, used by default. `web/src/context/theme-customization-provider.tsx`: only the upstream
  `default` palette clears the preset attribute.
- `web/src/lib/theme-storage.ts`, `web/index.html`: theme preference keys use the `site:theme:v1:` prefix.
- `web/src/lib/constants.ts`, `system-brand.tsx`, `footer.tsx`: neutral fallback name before the
  SystemName option loads.
- `web/public/favicon.ico`, `web/public/logo.png`: neutral fallback icons.
- Locale files: strings for the above in all seven languages (`web/scripts/huiliu-add-i18n.mjs`
  appends keys without re-serializing the files).
