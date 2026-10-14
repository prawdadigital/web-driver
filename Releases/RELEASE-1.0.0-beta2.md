# web-driver v1.0.0-beta2

A [WebDriver](https://www.w3.org/TR/webdriver/) client for Go that drives both
**Selenium 4** (web browsers) and **Appium 2** (mobile and desktop) over the W3C
protocol. This is the second pre-release of the maintained fork of
[tebeka/selenium](https://github.com/tebeka/selenium).

This is a bug-fix release on top of
[v1.0.0-beta1](RELEASE-1.0.0-beta1.md); no API changes.

## Fixes

A user-reported cookie bug plus five issues found by a multi-agent audit of
every `WebDriver`/`WebElement`/`ShadowRoot` interface method against the W3C
WebDriver spec.

- [#12](https://github.com/prawdadigital/web-driver/issues/12) **`AddCookie`
  dropped cookies without an explicit expiry** — `Cookie.Expiry` was serialized
  as `"expiry": 0` for session cookies, telling the browser the cookie expired
  at the Unix epoch, so it was discarded. `Expiry` is now `omitempty`, producing
  a proper session cookie.
- [#15](https://github.com/prawdadigital/web-driver/issues/15) **`MaximizeWindow`
  inverted its name guard (legacy path)** — a non-empty name was overwritten
  with the current window handle (maximizing the wrong window), and an empty name
  produced the malformed URL `/session/{id}/window//maximize`. The guard is now
  correct: an empty name resolves to the current window, a named window passes
  through.
- [#17](https://github.com/prawdadigital/web-driver/issues/17) **`CurrentURL`
  could panic** — it dereferenced a `*string` with no nil check, panicking on a
  `{"value":null}` response. It now routes through the shared `stringCommand`
  helper, which returns an error instead.
- [#19](https://github.com/prawdadigital/web-driver/issues/19) **`ByName` found
  only `<input>` elements on W3C** — the CSS emulation of the removed `name`
  locator used `input[name=...]`, silently missing `<select>`, `<textarea>`,
  `<form>`, `<iframe>`, and other named elements. It now uses a tag-agnostic
  `*[name=...]` selector, matching every find path (`WebDriver`, `WebElement`,
  `ShadowRoot`).
- [#21](https://github.com/prawdadigital/web-driver/issues/21) **`ReleaseActions`
  left the action queue stale** — it issued the W3C release but never cleared
  `storedActions`, so released actions were silently replayed on the next
  `PerformActions`. The queue is now cleared, mirroring `PerformActions`.
- [#23](https://github.com/prawdadigital/web-driver/issues/23) **`round()` was
  off by one at exactly ±0.5** — `WebElement.Location`/`LocationInView`/`Size`
  rounded `0.5`→`0` and `-0.5`→`0`. It now delegates to `math.Round`
  (round half away from zero).

## Migration

No code changes are required coming from v1.0.0-beta1 — run
`go get github.com/prawdadigital/web-driver`. See
[v1.0.0-beta1](RELEASE-1.0.0-beta1.md) for the full feature set and migration
notes from `tebeka/selenium`, and the [README](../README.md) for usage examples.
