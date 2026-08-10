# Plan: Selenium 4 & Appium 2 support

This fork of `tebeka/selenium` is being extended to support **Selenium 4** and
**Appium 2**. This document is the working plan; update it as milestones land.

## Current state (baseline)

The client already speaks both the legacy JSON Wire protocol and W3C WebDriver,
negotiated in `remoteWD.NewSession` (`remote.go`), which sets `wd.w3cCompatible`.
`newW3CCapabilities` (`remote.go`) passes through any capability whose name
contains a `:` into `alwaysMatch` — this is why vendor-prefixed caps
(`appium:`, `goog:`, `moz:`) already flow through and a basic Appium 2 session
connects today. Neither task is a protocol rewrite.

Selenium-3-specific assumptions that must change:

- `vendor/init.go` downloads Selenium **3.141.59** from the old
  `selenium-release.storage.googleapis.com` bucket.
- `service.go` `NewSeleniumService` launches the JAR via `GridLauncherV3` under
  the `/wd/hub` base path — both removed/changed in Selenium 4.
- `NewSession`'s first attempt sends both `capabilities` and
  `desiredCapabilities`; needs verification against a strict S4 Grid.
- No Appium package exists (no mobile commands, no capability helpers).

## Key design decision: one element interface, two driver interfaces

Under W3C a mobile element and a web element respond to the same element-level
commands, so we keep a **single `WebElement` interface**. Mobile-specific
behavior (contexts, orientation, geolocation, app lifecycle, device keys,
gestures) lives at the **driver** level.

- Element type: **`WebElement`** (one interface), extended with new W3C methods.
- Driver types: the existing **`WebDriver`**, plus a new **`Mobile`** interface
  in a new `appium/` package that *embeds* `WebDriver` and adds mobile commands.
  `appium.NewRemote(...)` returns a `Mobile`.

This matches the modern Appium clients (Java `java-client` v8+, Python, JS),
which dropped a separate mobile element type and return plain `WebElement`.

## Workstreams

### WS1 — Selenium 4: correctness + tooling
1. `vendor/init.go`: download the Selenium 4 server JAR
   (`selenium-server-<ver>.jar`, GitHub releases) instead of the 3.141 bucket
   JAR; keep GeckoDriver/ChromeDriver fetch current. *(Confirm exact URL/naming
   at implementation time — no web access when this plan was written.)*
2. `service.go` `NewSeleniumService`: replace `GridLauncherV3` with
   `java -jar selenium-server-<ver>.jar standalone --port N`; reconcile the
   served base path (`/wd/hub` vs root `/`) so the launcher and `addr` agree.
3. `remote.go` `NewSession`: verify a strict S4 Grid accepts the first attempt;
   if it rejects the extra `desiredCapabilities` key, gate the legacy key off
   once we know we're talking W3C.
4. `service.go` `start`/`Stop`: re-verify the `/status` readiness handshake and
   shutdown behavior against S4.

### WS2 — Selenium 4: new W3C API surface
Add to `WebDriver`/`WebElement` and `remote.go`, each gated on
`wd.w3cCompatible`:
- Relative locators (`above` / `below` / `near` / `toLeftOf` / `toRightOf`)
- `NewWindow(tab bool)` — new tab/window
- Print page to PDF
- Shadow DOM: `GetShadowRoot` + find-from-shadow-root
- W3C window rect: `GetWindowRect` / `SetWindowRect`

### WS3 — Appium 2: full support
New `appium/` package mirroring `chrome/` and `firefox/`:
- `appium.Capabilities` helper that enforces the `appium:` prefix (Appium 2
  rejects unprefixed non-standard caps).
- `Mobile` interface embedding `WebDriver` + mobile commands: contexts,
  orientation, geolocation, app lifecycle (install/activate/terminate/remove),
  device keys, and W3C gesture actions.
- Handle Appium 2's base-path change (`/wd/hub` → `/`); `NewRemote`'s
  `urlPrefix` already supports this — document and default sensibly.
- Optional `NewAppiumService` launcher wrapping the `appium` CLI.

### WS4 — Real-browser testing
- This dev environment is **macOS**; the vendored binaries are Linux-only
  (`chrome-linux`). For live testing, drive locally-installed Chrome + a matching
  ChromeDriver (and/or the Selenium 4 JAR) against real pages, exercising each
  new feature (relative locators, new window, print-to-PDF, shadow DOM).
- Extend `internal/seleniumtest` subtests to cover the new features across
  drivers; add a `TestSelenium4` top-level test.
- Validate Appium against a running Appium 2 server; add a documented mobile
  smoke test.

## Sequencing

WS1 (tooling — lets us launch S4) → WS2 (new API, with live verification) →
WS3 (Appium package) → WS4 (broaden tests + docs). Each milestone ends with
`gofmt`, `go test`, and a real-browser run before moving on.

## Progress

### WS1 — Selenium 4 correctness + tooling — DONE (live-verified against 4.39.0)

Verified end-to-end on macOS: `NewSeleniumService` launches Selenium 4.39.0,
`NewSession` negotiates a W3C session with real headless Chrome 151, and
navigation / `FindElement` / `Text` all work.

- `vendor/init.go`: replaced the pinned 3.141.59 bucket download with
  `addLatestGithubRelease(SeleniumHQ/selenium, ^selenium-server-.*\.jar$)`.
- `service.go` `NewSeleniumService`: `GridLauncherV3` → `org.openqa.selenium.grid.Main
  standalone --port N` (kept classpath launch for HTMLUnit; base path stays
  `/wd/hub`).
- Two bugs found and fixed that only manifest against a strict S4 Grid:
  1. `chrome/capabilities.go`: `W3C bool` now uses `json:"w3c,omitempty"` — an
     unset value emitting `"w3c":false` forced the removed legacy protocol and
     S4 rejected the handshake.
  2. `remote.go` `executeCommand`: W3C error detection decoded `stacktrace` into
     a `string`, but S4 returns a JSON **array** for session-creation errors, so
     the unmarshal failed and a 500 was silently treated as success. Now decoded
     via `json.RawMessage`.

### WS2 — Selenium 4 new W3C API — DONE (live-verified against 4.39.0)

All added behind the W3C session and verified end-to-end with headless Chrome 151.

- Window rect: `GetWindowRect`/`SetWindowRect` via `/session/:id/window/rect`.
- `NewWindow(tab bool)` via `/session/:id/window/new`, returning the new handle
  and context type (does not switch to it).
- `Print(PrintOptions)` via `/session/:id/print`, returning decoded PDF bytes;
  options for orientation/scale/background/margins/page-size/page-ranges.
- Shadow DOM: `WebElement.GetShadowRoot` + `ShadowRoot` interface with
  `FindElement`/`FindElements` via the `/shadow` endpoints.
- Relative locators (`relative.go`): `With(by, value)` + chainable
  `Above`/`Below`/`ToLeftOf`/`ToRightOf`/`Near`/`NearWithin`, plus
  `FindElementRelative`/`FindElementsRelative` on the WebDriver interface.
  Implemented by finding base candidates in Go, then running a compact JS that
  replicates Selenium's own `relative.js` predicates and center-proximity sort
  (verified: directional matches, chained filters, and proximity ordering).

### WS3 — Appium 2 support — DONE (verified via mock server; live device pending)

New `appium/` package plus a small extension point in the core client.

- `selenium.WebDriver.ExecuteCommand(method, path, params)`: exported raw
  session-command executor (reuses the existing W3C error handling) so
  extensions can issue commands outside the core API. Implemented on remoteWD.
- `appium/capabilities.go`: fluent `Capabilities` builder that auto-applies the
  `appium:` prefix to non-standard capabilities (Appium 2 rejects unprefixed
  ones); standard W3C caps and already-prefixed vendor caps are left alone.
- `appium/appium.go`: `Mobile` interface embedding `selenium.WebDriver`, with
  `appium.NewRemote` (defaults documented for Appium 2's root base path) and
  `NewMobile` (wrap an existing session). Commands: contexts
  (available/current/switch), orientation, geolocation, app lifecycle
  (install/isInstalled/activate/terminate/remove/state), keyboard
  (hide/isShown/press+longPress keycode), device (lock/unlock/isLocked/shake/
  time), and settings (get/update).
- `appium/appium_test.go`: httptest mock verifying the exact wire format
  (method/path/body) of session creation and every command group, plus the
  capability-prefix logic. `go test ./appium/` passes.

Note: the user has already confirmed a basic Appium session connects through
this client. A full live smoke test of the new mobile commands needs a running
Appium 2 server + device/emulator, which is not available in this dev
environment; the mock tests lock down the protocol in the meantime.

### Resolved open items

- S4 server JAR: `selenium-server-<ver>.jar` from `SeleniumHQ/selenium` GitHub
  releases (latest is 4.39.0).
- Base path: S4 standalone serves both `/` and `/wd/hub`; kept `/wd/hub`.
- Dual `capabilities` + `desiredCapabilities` payload: **accepted** by the S4
  Grid (no need to gate the legacy key off).

## Pre-existing issues to address before WS4 live tests

These predate this work and currently block `go test` from compiling:

- `example_test.go` references `selenium.StorePointerActions` (not a package-level
  function) — breaks test compilation.
- `remote.go` `Cookie.SameSite` has a malformed struct tag
  (`json:"sameSite",omitempty` → should be `json:"sameSite,omitempty"`); `go vet`
  flags it.

## Open items to confirm during implementation

- Live smoke test of the appium mobile commands against a real Appium 2 server
  + device/emulator (WS4, on the user's side).
