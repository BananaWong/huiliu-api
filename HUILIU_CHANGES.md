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
- `common/constants.go`: the built-in SystemName default is the neutral `API` (only used until
  the SystemName option is set).

Frontend
- `web/src/huiliu/`: fork-specific code (auth page frame, license notice, source links).
- `web/src/features/about/index.tsx`: the About page always ends with the license notice
  (attribution, upstream link, source of this build, license); the empty state no longer
  shows the upstream repository.
- `web/src/features/auth/auth-layout.tsx`: auth pages use the two-column frame from `web/src/huiliu/`.
- `web/src/routes/_authenticated/route.tsx`: signed-in pages use `ConsoleLayout` from
  `web/src/huiliu/console-shell.tsx`. Account pages (overview, dashboard, keys, usage logs,
  wallet, profile, security) get a top-tab shell without the sidebar; admin pages and nested
  workspaces keep the upstream sidebar layout. Administrators get an "Admin console" tab.
- `web/src/components/layout/components/public-header.tsx`: the public header is a full-width bar
  with a bottom rule (no floating pill on scroll); the active link has a primary-colour underline.
- `web/src/features/pricing/index.tsx`: plain primary-colour wash instead of the multi-colour
  radial background; left-aligned page header with the search box beside it on wide screens.
- `web/src/styles/theme-presets.css`, `web/src/lib/theme-customization.ts`: new `huiliu` (Teal Ink)
  palette, used by default. `web/src/context/theme-customization-provider.tsx`: only the upstream
  `default` palette clears the preset attribute.
- `web/src/lib/theme-storage.ts`, `web/index.html`: theme preference keys use the `site:theme:v1:` prefix.
- `web/src/lib/constants.ts`, `system-brand.tsx`, `footer.tsx`: neutral fallback name before the
  SystemName option loads.
- `web/public/favicon.ico`, `web/public/logo.png`: neutral fallback icons.
- Locale files: strings for the above in all seven languages (`web/scripts/huiliu-add-i18n.mjs`
  appends keys without re-serializing the files). The "Model Square" label reads
  "Models & Pricing" / 模型与价格 / 模型與價格.

## 2026-11-27

Server
- `service/huiliu_health.go`, `setting/operation_setting/huiliu_health_setting.go`: passive channel health job
  on the master node, every five minutes. It counts real request outcomes from the consume and error logs
  per request id (retries on the same channel count once); caller errors (400, 413, 415, 422, 499) do not
  count. A channel with automatic disabling (`auto_ban`) on is disabled after 3 consecutive failures or a
  failure rate of 50% over at least 10 requests in the last 30 minutes, and put back after 30 minutes,
  doubling on each repeat within 6 hours (at most 8 hours). The state lives in the channel's status reason
  (`huiliu-health#<strikes>: …`); the job only re-enables channels it disabled itself. All thresholds and
  the on/off switch are options under `huiliu_health_setting.*`. No probes are sent and nothing is spent.
- `controller/huiliu_health.go`, `router/api-router.go`, `middleware/access_token_routes.go`:
  `GET /api/huiliu/channel-health` (admins: per channel and model, last hour and day) and
  `GET /api/huiliu/model-status` (signed-in users: per model of their group, no channel names or error text).
- `common/init.go`: `ERROR_LOG_ENABLED` defaults to true (the health job reads error logs).
- `main.go`: starts the health job.
- `common/huiliu_upstream_wording.go`, `service/huiliu_upstream_wording.go`, `controller/relay.go`,
  `controller/playground.go`, `model/log.go`, `setting/operation_setting/huiliu_privacy_setting.go`, `i18n/`:
  callers no longer read an upstream's own wording. For a request that reached a channel, an error that
  carries what the upstream returned (any type other than `api_error`) becomes `relay.upstream_unavailable`
  when the upstream could not serve it (401, 402, 403, 408, 429, 5xx); the gateway's own errors (`api_error`,
  such as a rejected request or insufficient quota) keep their wording whatever the status. Every other error
  keeps its text with every URL, domain name and IPv4 address masked (the stock masking covers only the
  channel's own address). The type decides, not the code: a relay upstream answers with codes such as
  `insufficient_user_quota`. The same applies to the error logs users read and export; administrators still
  see the original text. Covered: the main relay and playground error responses; Gemini-native, Responses
  WebSocket and task errors are not. Option `huiliu_privacy_setting.hide_upstream_wording`, on by default.

Frontend
- `web/src/huiliu/health/`: Channel Health page (admin sidebar, `/channel-health`) and Model Status page
  (account tabs, `/model-status`); `web/src/hooks/use-sidebar-data.ts` links them; strings in all seven
  languages; status labels registered in `web/src/i18n/static-keys.ts`.
