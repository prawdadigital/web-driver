# web-driver v1.0.0

A [WebDriver](https://www.w3.org/TR/webdriver/) client for Go that drives both
**Selenium 4** (web browsers) and **Appium 2** (mobile and desktop) over the W3C
protocol. This is the first release of the maintained fork of
[tebeka/selenium](https://github.com/tebeka/selenium), reorganized into a shared
contract with focused implementation packages.

## Highlights

- Full **Selenium 4** support plus its new W3C API (relative locators,
  print-to-PDF, shadow DOM, new window/tab, window rect, wheel actions).
- A new **Appium 2** client covering native mobile *and* native desktop apps,
  with contexts, app lifecycle, gestures, and driver extension commands.
- A clean package layout — `webdriver` (contract) · `remote` (transport) ·
  `selenium` (servers) · `appium` (mobile) — where `appium` never depends on
  `selenium`.
- Verified live on Chrome and Firefox; broad unit coverage; full godoc; MIT
  licensed; `go 1.21`.

## Enhancements

- [#4](https://github.com/prawdadigital/web-driver/issues/4) **Selenium 4
  support** — launch Selenium 4 servers, negotiate W3C sessions against strict
  grids, and use the new W3C commands: relative ("friendly") locators,
  print-to-PDF, shadow DOM, new window/tab, window rect, and wheel/scroll actions.
- [#5](https://github.com/prawdadigital/web-driver/issues/5) **Appium 2 mobile
  support** — a new `appium` package with an `appium:`-prefixing capability
  builder and a `Mobile` driver: contexts (native ↔ webview), orientation,
  geolocation, app lifecycle, keyboard, device, and settings commands.
- [#6](https://github.com/prawdadigital/web-driver/issues/6) **Input actions &
  gestures** — W3C wheel/scroll actions; touch-gesture helpers (`Tap`, `Swipe`,
  `Zoom`, `Pinch`, …); typed `mobile:` gesture wrappers; and a generic
  `ExecuteExtension` for any driver command.
- [#7](https://github.com/prawdadigital/web-driver/issues/7) **Native & desktop
  apps** — platform/driver constants, Appium element-location strategies, and
  `ExecuteExtension` for `mobile:`/`windows:`/`macos:` commands, so the same
  client drives Android/iOS *and* Windows/macOS desktop apps (`Driver` alias).

## Fixes

Found by a live audit of every WebDriver/WebElement/Appium method against Chrome
and Firefox on Selenium 4.

- [#8](https://github.com/prawdadigital/web-driver/issues/8) **`Capabilities()`
  broken on W3C** — it used the legacy `GET /session/:id` command (Selenium 4:
  "unknown command"). Now returns the capabilities the server granted at session
  creation, tracking desired vs. granted capabilities.
- [#9](https://github.com/prawdadigital/web-driver/issues/9) **Legacy
  mouse/submit/move methods fail on W3C** — `Click(button)`, `DoubleClick`,
  `ButtonDown`, `ButtonUp`, `WebElement.MoveTo`, and `WebElement.Submit` used
  removed JSON Wire endpoints; reimplemented over the W3C Actions API (and a
  form-submit script for `Submit`).
- [#11](https://github.com/prawdadigital/web-driver/issues/11) **`Log` test fails
  on Firefox + Selenium 4** — `Log` is a ChromeDriver-only extension; the test
  now skips geckodriver (which implements no log command), and the method is
  documented as such.
- Also: `chrome.Capabilities` no longer emits `"w3c": false`, which had forced
  the removed legacy protocol and broke Chrome sessions on Selenium 4.

## Migration

Coming from `tebeka/selenium`, or updating an early checkout of this fork:

1. **Module path** is `github.com/prawdadigital/web-driver` — run
   `go get github.com/prawdadigital/web-driver`.
2. **Imports changed** with the restructure:
   - The client/contract is the root package `webdriver`
     (`github.com/prawdadigital/web-driver`), not `selenium`.
   - Launch servers from `.../selenium`; drive mobile from `.../appium`.
   - `selenium.NewRemote` still exists as a convenience wrapper.
3. **Go 1.21+** is required.
4. See the [README](../README.md) for the current usage examples.
