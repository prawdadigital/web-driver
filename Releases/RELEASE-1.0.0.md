# web-driver v1.0.0

The first release of `github.com/prawdadigital/web-driver` — a fork of
[tebeka/selenium](https://github.com/tebeka/selenium) reworked into a single Go
client for both **Selenium 4** (browsers) and **Appium 2** (mobile and desktop),
speaking the W3C WebDriver protocol.

## Architecture

Reorganized into a pure-contract root with focused implementation packages:

- `webdriver` (root) — interfaces and shared types, no implementation
- `remote` — the W3C HTTP transport
- `selenium` — local WebDriver server launching + convenience wrappers
- `appium` — the Appium 2 client (depends on `webdriver` + `remote`, never `selenium`)

Commits: `6621f15`, `523f052`, `d9a7b6d`, `ece9687`, `e3b79d6` (go 1.21 + layout).

## Selenium 4

- Server launch, W3C session negotiation for strict grids, and driver tooling — `3af19f2`
- Relative locators, print-to-PDF, shadow DOM, new window/tab, window rect — `3fca7eb`, `2b476b9`
- Wheel/scroll input actions — `c0a8a31`
- Verified live on Chrome and Firefox — `b1cf28b`, `1eb518f`

## Appium 2

- `appium` package: capability builder, `Mobile`/`Driver` session, contexts,
  orientation, geolocation, app lifecycle, keyboard, device, settings — `cb008ef`
- Gestures: W3C touch helpers (Tap/Swipe/Zoom/Pinch) and typed `mobile:` wrappers — `68b0c01`, `144f718`
- Native + desktop apps: platform/driver constants, Appium locators,
  `ExecuteExtension` for `mobile:`/`windows:`/`macos:`, and the `Driver` alias — `dcee4c4`, `99b958d`

## Quality & release readiness

- Unit tests across all packages (webdriver 97%, remote 67%, firefox 91%, …) — `5a8e630`, `07b1677`
- Full godoc, reworked README, MIT LICENSE, clean `go vet` — `ba19f46`, `16cade3`
- Test-suite hardening (session-creation retry, navigation race) — `9412461`

See the git history for the complete set of changes.
