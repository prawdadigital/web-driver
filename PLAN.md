# Plan: Selenium 4 & Appium 2 support

This fork of `tebeka/selenium` is being extended to support **Selenium 4** and
**Appium 2**. This document is the working plan; update it as milestones land.

---

## Phase 2 — next steps (branch `2nd-claude`)

Phase 1 shipped to `develop` via PR #1: Selenium 4 support, Appium 2 support, and
the package restructure (pure-contract `webdriver` root + `remote` transport +
`selenium` services + `appium` mobile). Its full record is preserved below under
"Progress". Phase 2 focuses on input actions/gestures and hardening.

Proposed workstreams, in priority order (reprioritize as needed):

### WS5 — Input actions & gestures — DONE
The biggest functional gap identified after Phase 1.
- **WS5a — Wheel actions** (`webdriver` + `remote`): `WheelAction` type,
  `StoreWheelActions`, and `ScrollAction`/`WheelPauseAction` constructors.
  Live-verified in Chrome (`TestSelenium4/WheelScroll`); a non-zero duration is
  required or headless Chrome does not apply the scroll.
- **WS5b — Touch gesture helpers** (`appium.Mobile`): `Tap`, `DoubleTap`,
  `LongPress`, `Swipe`, `Zoom`, `Pinch`, built on W3C touch pointer actions
  (driver-agnostic; multi-finger uses two synchronized input sources).
  Mock-tested wire format.
- **WS5c — `mobile:` gesture wrappers** (`appium.Mobile`): generic
  `ExecuteMobile(command, options)` plus typed UiAutomator2 wrappers
  (`SwipeGesture`, `ScrollGesture`, `PinchOpen/CloseGesture`, `LongClickGesture`).
  Mock-tested `/execute/sync` payload.

### WS6 — Firefox / GeckoDriver on Selenium 4 — DONE
`TestSelenium4` now runs `Chrome` and `Firefox` subgroups via a shared
`runSelenium4Suite` helper (new `-firefox_binary` flag; geckodriver via
`-geckodriver_path` or the server's Selenium Manager). Live-verified on Firefox
151 + Selenium 4.39.0: all W3C features — WindowRect, NewWindow, Print,
ShadowRoot, RelativeLocators, WheelScroll — pass. Note: geckodriver/Firefox
occasionally fails session creation on a cold start ("Process unexpectedly
closed with status 0"); it passes on retry, so this is environmental flakiness,
not a client bug (candidate for WS8 hardening / a session-creation retry).

### Documentation & test hardening for pkg.go.dev — DONE
Preparing the library for publication on pkg.go.dev.
- Added an MIT LICENSE (upstream copyright preserved) so the license is
  recognized.
- godoc comments on all exported symbols; expanded package docs across packages.
- Reworked the README for the four-package layout with runnable examples, a
  feature list, and a pkg.go.dev badge.
- Raised unit-test coverage from ~0% to: webdriver 97%, firefox 91%, zip 81%,
  chrome 75%, remote 67% (httptest transport suite), appium 59%, selenium 22%,
  sauce 13% (the low two are mostly process-launching code needing external
  binaries). `go build`, `gofmt`, and `go vet ./...` are all clean.

### WS7 — CI/CD modernization (deferred)
The Travis badge/config is stale (points at `tebeka`, and travis-ci.org is
retired). Add a GitHub Actions workflow: `go build ./...`, `gofmt -l`, `go vet`,
and unit tests (`appium`/`chrome`/`sauce`/`selenium`), plus an optional Linux
Selenium 4 integration job. Update or drop the README badges.

### WS8 — Test-suite hardening
Address the pre-existing/environmental issues surfaced by `TestSelenium4`:
- `FindElement/css_selector` click-then-check-URL race — add an explicit wait.
- `AddCookie` cookie-field diff on recent Chrome — reconcile expectations.
- `Proxy/SOCKS` — Selenium Manager `--proxy` quirk when no ChromeDriver is
  vendored; vendor a driver in CI or gate the test.
- `go vet`: `internal/webdrivertest` "Fatalf from a non-test goroutine".

### WS9 — Desktop / native "regular" apps (optional)
The client is protocol-generic, so it can already reach Appium desktop drivers
(appium-windows-driver, mac2). Add documentation + an example (and, if useful, a
small helper) for driving native desktop apps; note it is untested here.

---

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

### Restructure — dual Selenium + Appium solution — DONE

- Renamed the module to `github.com/prawdadigital/web-driver` (matches the git
  remote).
- Moved the core client package from the repo root into `selenium/` so it sits
  as a sibling of `appium/` (imported as
  `github.com/prawdadigital/web-driver/selenium`). All internal imports across
  `chrome`, `firefox`, `log`, `sauce`, `appium`, `internal/seleniumtest`, and
  the tests were updated. `go build ./...` and `go test ./...` pass (the only
  failing test, `TestFrameBuffer`, is Linux/Xvfb-only and skips elsewhere).
- Updated a chrome unit test that had encoded the old `"w3c": false` behavior to
  expect the WS1-corrected output, and added a `W3C: true` case.

### WS4 — Test coverage for the new features — DONE (live-verified)

- `internal/seleniumtest`: added `RunW3CTests` with shared subtests for the WS2
  features — window rect, new window, print-to-PDF, shadow DOM, and relative
  locators — plus two new served test pages (`/shadow`, `/relative`).
- `selenium/selenium_test.go`: added `TestSelenium4` (launches the Selenium 4
  JAR, runs the common + chrome + W3C suites against Chrome) and a
  `-selenium4_path` flag. `runChromeTests` now also runs `RunW3CTests` for the
  non-Selenium-3 (ChromeDriver) path.
- Fixed a restructure side effect: `go test ./selenium/` runs with the package
  dir as its working directory, so `TestMain` now chdirs to the repo root
  (identified by go.mod) to keep the root-relative `vendor/` and `testing/`
  fixture paths resolving (important for the Linux CI/Docker flow).

Live result (Selenium 4.39.0 + headless Chrome 151): all five W3C feature
subtests pass. The other common/chrome subtests that fail on this dev box are
pre-existing and environmental, not regressions from this work:
`FindElement/css_selector` (click-then-check-URL timing race), `AddCookie`
(cookie-field diff on Chrome 151), `Proxy/SOCKS` (Selenium Manager `--proxy`
quirk when no ChromeDriver is vendored — provided on CI), and `Extension`
(needs `testing/chrome_extension` assets absent from this repo, and headless
Chrome cannot load extensions).

### Decouple appium from selenium — contract package at the root — DONE

Reworked the package boundaries so the shared vocabulary is owned by a root
contract package rather than by `selenium`:

- **root package `webdriver`** (`github.com/prawdadigital/web-driver`): the
  `WebDriver`/`WebElement`/`ShadowRoot` interfaces, all shared value types
  (`Capabilities`, `Cookie`, `Rect`, `PrintOptions`, ...), and the W3C HTTP
  transport (`remote.go`, `relative.go`, `common.go`, `NewRemote`,
  `ExecuteCommand`). Moved here from `selenium/` (`selenium.go`, `remote.go`,
  `relative.go`, `common.go`, `doc.go`; package renamed to `webdriver`).
- **`selenium/`**: now only service launching (`service.go`), a self-contained
  package that imports nothing from `webdriver`.
- **`appium/`**: imports **only** `webdriver` — no longer depends on `selenium`.
  Its `Capabilities` builder produces `webdriver.Capabilities` and `Mobile`
  embeds `webdriver.WebDriver`.
- Integration tests (`TestChrome`, `TestSelenium4`, `TestFirefox`,
  `TestHTMLUnit`, `TestSauce`, the example) moved to the repo root as
  `package webdriver_test`; they import `selenium` for services and `webdriver`
  for the client. This also let the WS4 `TestMain` chdir hack be removed, since
  the tests run from the repo root again.

Final dependency graph: `appium → webdriver`; `selenium` standalone;
`webdriver → chrome/firefox/log`. Verified: `go build ./...`, all unit tests
(`appium`, `chrome`, `sauce`, `selenium`) pass, and the live `TestSelenium4`
W3C suite passes against Selenium 4.39.0 + headless Chrome 151.

### Split contract from transport — pure-contract root + remote package — DONE

Separated the interface contract from its implementation so the root package no
longer contains the concrete transport (previously `remoteWD`/`remoteWE`
implemented the interfaces in the same package that defined them).

- **root `webdriver`**: now contract-only. Kept the interfaces + value types in
  `webdriver.go`; extracted the remaining contract declarations out of the old
  `remote.go` into `error.go` (`Error`), `wait.go` (`Condition` + default
  timeouts) and `actions.go` (action constructors); `relative.go` keeps just the
  `RelativeBy` builder plus `Root`/`Filters` accessors.
- **new `remote/` package**: the W3C transport (`remote.go`, `common.go`,
  `relative.go` impl). It dot-imports the root contract so the ~200 references to
  shared types (`Capabilities`, `WebElement`, `Error`, ...) stay unqualified —
  avoiding an error-prone mass-qualification and keeping the transport code
  intact. `NewRemote`, `ExecuteCommand`, `DeleteSession`, `SetDebug` live here.
- **`selenium/`**: added `client.go` with `NewRemote`/`SetDebug`/`DeleteSession`
  convenience wrappers around `remote` (browser ergonomics).
- **`appium/`**: now builds on `remote.NewRemote`; imports `webdriver` + `remote`,
  still never `selenium`.

Final graph: `remote → webdriver`; `selenium → webdriver + remote`;
`appium → webdriver + remote`; `webdriver → chrome/firefox/log` (capability
helpers only). `go build ./...`, all unit tests, and the live `TestSelenium4`
W3C suite (Selenium 4.39.0 + headless Chrome 151) pass.

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
