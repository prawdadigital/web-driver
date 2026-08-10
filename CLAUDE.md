# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a [WebDriver](https://www.w3.org/TR/webdriver/) client for Go (`github.com/prawdadigital/web-driver`) — a single solution for both **Selenium** (web browsers) and **Appium 2** (mobile). It drives targets by speaking the W3C WebDriver HTTP protocol to a server (a standalone Selenium JAR, ChromeDriver, GeckoDriver, or an Appium 2 server). It is a client library — there is no `main` package for the module itself.

The module is organized as a **shared contract at the root** with two implementation packages built on it:

- **root package `webdriver`** (`github.com/prawdadigital/web-driver`) — the shared contract *and* the W3C HTTP transport: the `WebDriver`/`WebElement`/`ShadowRoot` interfaces, all shared value types (`Capabilities`, `Cookie`, `Rect`, `PrintOptions`, ...), and the concrete remote client including `NewRemote`. Both implementation packages depend only on this.
- **`selenium/`** — launches/manages local WebDriver **server subprocesses** (Selenium JAR, ChromeDriver, GeckoDriver). Independent of `webdriver`.
- **`appium/`** — the Appium 2 **mobile** client: a `Capabilities` builder and a `Mobile` interface (embeds `webdriver.WebDriver`). Depends only on `webdriver`, never on `selenium`.

## Commands

```bash
go test ./...                        # Build and run all packages' tests
go test .                            # Root: transport + integration tests (incl. TestChrome, TestSelenium4)
go test ./selenium/                  # Service-launching tests
go test ./appium/                    # Appium package tests (mock HTTP server; no device needed)
go test . -run=TestSelenium4                     # Selenium 4 integration group (needs the S4 JAR)
go test . -run=TestSelenium4/RelativeLocators    # A single subtest (regex supported)
go test . -skip TestFrameBuffer                  # Skip the Xvfb-only test (e.g. on macOS)
go test . --arg --help                           # List all test flags (driver/binary paths)
go test . --docker                               # Run the suite hermetically inside Docker
gofmt -l .                                       # Check formatting (required before committing)
```

The integration tests (`TestChrome`, `TestSelenium4`, `TestFirefox`, `TestHTMLUnit`, `TestSauce`) live in the root package as `package webdriver_test`; they import `selenium` for service launching and `webdriver` for the client.

Tests require WebDriver binaries. Download them into `vendor/` for testing:

```bash
cd vendor && go run init.go --alsologtostderr --download_browsers --download_latest && cd ..
```

`vendor/init.go` fetches the latest Selenium 4 server JAR (from SeleniumHQ GitHub releases), ChromeDriver, GeckoDriver, and HTMLUnit. Local Selenium/HTMLUnit testing also needs `xvfb` and a JRE (`sudo apt-get install xvfb openjdk-11-jre`); `TestFrameBuffer` requires `Xvfb` and only runs on Linux. Tests default to headless Chrome/Firefox; pass `--start_frame_buffer` to use an Xvfb server instead.

## Architecture

The root `webdriver` package holds the contract and the transport. Its central abstractions are the `WebDriver` interface (browser session: navigation, cookies, script execution, actions, waits) and the `WebElement` interface (per-element operations). Both are defined in `selenium.go`, alongside all shared value types (`Capabilities`, `Proxy`, `Cookie`, `Status`, `Rect`, `PrintOptions`) and protocol constants (`By*` locators, keyboard keys, pointer/key action types).

- **`remote.go`** (package `webdriver`) — the only concrete `WebDriver` implementation (`remoteWD`). `NewRemote(caps, urlPrefix)` connects to an already-running WebDriver server over HTTP and translates every interface method into WebDriver protocol requests. It negotiates the legacy JSON Wire protocol vs. the W3C spec per session (`w3cCompatible`). This is where protocol-level behavior and browser quirks live. `ExecuteCommand(method, path, params)` is the exported extension point that lets other packages (e.g. `appium`) issue arbitrary session commands reusing the W3C error handling.

- **`relative.go`** (package `webdriver`) — relative (spatial) locators: `With(...)` plus `FindElement(s)Relative`, implemented by finding base candidates then running a JS atom that mirrors Selenium's own geometry.

- **`selenium/service.go`** (package `selenium`) — launches and manages a local WebDriver **server subprocess** so you have something for `NewRemote` to connect to. `NewSeleniumService` (Selenium 4: `org.openqa.selenium.grid.Main standalone`), `NewChromeDriverService`, and `NewGeckoDriverService` each return a `*Service` that owns the child process. `ServiceOption` functional options configure it (e.g. `Display`, `StartFrameBuffer`). `FrameBuffer` wraps launching an Xvfb X server for headful browsers without a display. This package is self-contained and does not import `webdriver`.

The typical usage flow: start a `selenium.Service` → call `webdriver.NewRemote` against its URL → drive the returned `WebDriver`. For mobile, `appium.NewRemote` returns a `Mobile` (which embeds `webdriver.WebDriver`).

### Capabilities and option packages

`Capabilities` (a `map[string]interface{}`, in `webdriver`) carries session config. Browser-specific options are added via `AddChrome`/`AddFirefox`/`AddProxy`/`AddLogging`, which inject typed structs from the subpackages under their protocol-defined keys:

- **`chrome/`** — Chrome options (`chrome.Capabilities`), including packing/encoding extensions (CRX3) for upload.
- **`firefox/`** — Firefox options (`firefox.Capabilities`), including profile packaging.
- **`log/`** — log `Type` and `Level` constants used by `Capabilities.SetLogLevel` and `WebDriver.Log`.
- **`sauce/`** — Sauce Labs cloud testing: capability translation and the Sauce Connect proxy tunnel.
- **`appium/`** — Appium 2 mobile support: a `Capabilities` builder that enforces the `appium:` prefix, and a `Mobile` interface (embeds `webdriver.WebDriver`) with mobile commands (contexts, orientation, geolocation, app lifecycle, keyboard, device, settings). Built on `webdriver.WebDriver.ExecuteCommand`, and depends only on `webdriver`. Its tests run against an `httptest` mock, so no device is needed.

### Testing internals

- **`internal/seleniumtest/`** — the shared subtest bodies (import `webdriver` for the client, `selenium` for `ServiceOption`). The top-level `TestChrome`/`TestSelenium4`/`TestFirefox*`/`TestHTMLUnit` functions in the root `selenium_test.go` (`package webdriver_test`) each set up a driver and run this common suite (`RunCommonTests`, plus `RunChromeTests`/`RunW3CTests`) against it.
- **`internal/zip/`** — zip helpers used when packaging extensions/profiles.
- **`vendor/init.go`** — a standalone `main` binary (not module deps) that downloads browsers, drivers, and JARs for tests.
- **`testing/`** — Dockerfile and scripts for the `--docker` hermetic test path.

## Conventions

- Run `gofmt` on changed files before committing. A pre-commit hook is available: `ln -s ../../misc/git/pre-commit .git/hooks/pre-commit`.
- The client targets **Selenium 4** and **Appium 2** over the W3C protocol (plus standalone ChromeDriver/GeckoDriver). Legacy JSON Wire support remains for older servers but Selenium 2 is unsupported.
