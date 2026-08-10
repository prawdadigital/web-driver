# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a [WebDriver](https://www.w3.org/TR/webdriver/) client for Go (`github.com/tebeka/selenium`). It drives web browsers for automation and testing by speaking the WebDriver HTTP protocol to a WebDriver server (a standalone Selenium JAR, ChromeDriver, or GeckoDriver). It is a client library — there is no `main` package for the module itself.

## Commands

```bash
go test                        # Run all tests (skips a browser's tests if its binaries aren't found)
go test -test.run=TestChrome   # Run one top-level browser test group
go test -test.run=TestFirefoxGeckoDriver/<subtest>   # Run a single subtest (regex supported)
go test -test.v                # Verbose test-automation output
go test --arg --help           # List all test flags (driver/binary paths, etc.)
go test --docker               # Run the full test suite hermetically inside Docker
gofmt -l .                     # Check formatting (required before committing)
```

Tests require WebDriver binaries. Download them into `vendor/` for testing:

```bash
cd vendor && go run init.go --alsologtostderr --download_browsers --download_latest && cd ..
```

Local testing also needs `xvfb` and a JRE for the Selenium/HTMLUnit paths (`sudo apt-get install xvfb openjdk-11-jre`). Tests default to headless Chrome/Firefox; pass `--start_frame_buffer` to use an Xvfb server instead.

## Architecture

The core abstraction is the `WebDriver` interface (browser session: navigation, cookies, script execution, actions, waits) and the `WebElement` interface (per-element operations). Both are defined in `selenium.go`, alongside all shared value types (`Capabilities`, `Proxy`, `Cookie`, `Status`) and protocol constants (`By*` locators, keyboard keys, pointer/key action types).

Two layers work together:

- **`remote.go`** — the only concrete `WebDriver` implementation (`remoteWD`). `NewRemote(caps, urlPrefix)` connects to an already-running WebDriver server over HTTP and translates every interface method into WebDriver protocol requests. It handles both the legacy JSON Wire protocol and the W3C spec, negotiating which to use per session. This is where protocol-level behavior and browser quirks live.

- **`service.go`** — launches and manages a local WebDriver **server subprocess** so you have something for `NewRemote` to connect to. `NewSeleniumService`, `NewChromeDriverService`, and `NewGeckoDriverService` each return a `*Service` that owns the child process. `ServiceOption` functional options configure it (e.g. `Display`, `StartFrameBuffer`). `FrameBuffer` wraps launching an Xvfb X server for headful browsers without a display.

The typical usage flow: start a `Service` → call `NewRemote` against its URL → drive the returned `WebDriver`.

### Browser-specific packages

`Capabilities` (a `map[string]interface{}`) carries session config. Browser-specific options are added via `AddChrome`/`AddFirefox`/`AddProxy`/`AddLogging`, which inject typed structs from the subpackages under their protocol-defined keys:

- **`chrome/`** — Chrome options (`chrome.Capabilities`), including packing/encoding extensions (CRX3) for upload.
- **`firefox/`** — Firefox options (`firefox.Capabilities`), including profile packaging.
- **`log/`** — log `Type` and `Level` constants used by `Capabilities.SetLogLevel` and `WebDriver.Log`.
- **`sauce/`** — Sauce Labs cloud testing: capability translation and the Sauce Connect proxy tunnel.
- **`appium/`** — Appium 2 mobile support: a `Capabilities` builder that enforces the `appium:` prefix, and a `Mobile` interface (embeds `WebDriver`) with mobile commands (contexts, orientation, geolocation, app lifecycle, keyboard, device, settings). Built on `WebDriver.ExecuteCommand`, the exported raw-session-command extension point in `remote.go`.

### Testing internals

- **`internal/seleniumtest/`** — the shared subtest bodies. The top-level `TestChrome`/`TestFirefox*`/`TestHTMLUnit` functions in `selenium_test.go` each set up a driver and run this common suite against it, so behavior is verified identically across browsers.
- **`internal/zip/`** — zip helpers used when packaging extensions/profiles.
- **`vendor/init.go`** — a standalone `main` binary (not module deps) that downloads browsers, drivers, and JARs for tests.
- **`testing/`** — Dockerfile and scripts for the `--docker` hermetic test path.

## Conventions

- Run `gofmt` on changed files before committing. A pre-commit hook is available: `ln -s ../../misc/git/pre-commit .git/hooks/pre-commit`.
- Selenium 2 is no longer supported; the client targets Selenium 3 (W3C) and standalone GeckoDriver/ChromeDriver.
