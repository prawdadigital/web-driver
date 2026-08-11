# web-driver — a WebDriver client for Go (Selenium 4 & Appium 2)

[![Go Reference](https://pkg.go.dev/badge/github.com/prawdadigital/web-driver.svg)](https://pkg.go.dev/github.com/prawdadigital/web-driver)
[![Go Report Card](https://goreportcard.com/badge/github.com/prawdadigital/web-driver)](https://goreportcard.com/report/github.com/prawdadigital/web-driver)

A single [WebDriver](https://www.w3.org/TR/webdriver/) client for Go that drives
both **web browsers** (via Selenium 4 / ChromeDriver / GeckoDriver) and **mobile
devices** (via [Appium 2](https://appium.io/)). It speaks the W3C WebDriver
protocol and is tested against Chrome, Firefox, and Appium.

This is a maintained fork of [tebeka/selenium](https://github.com/tebeka/selenium),
restructured and extended for Selenium 4 and Appium 2.

## Packages

The module is a **pure contract at the root** with focused implementation
packages layered on top:

| Import path | Package | Purpose |
| --- | --- | --- |
| `github.com/prawdadigital/web-driver` | `webdriver` | The contract: `WebDriver`/`WebElement`/`ShadowRoot` interfaces and all shared types (`Capabilities`, `Cookie`, `Rect`, `PrintOptions`, actions, `RelativeBy`, …). No implementation. |
| `.../remote` | `remote` | The W3C HTTP transport: `NewRemote`, `ExecuteCommand`, `DeleteSession`, `SetDebug`. |
| `.../selenium` | `selenium` | Launches/manages local WebDriver servers (Selenium JAR, ChromeDriver, GeckoDriver) and provides browser-friendly `NewRemote`/`SetDebug`/`DeleteSession` wrappers. |
| `.../appium` | `appium` | Appium 2 mobile client: a capability builder and a `Mobile` interface (contexts, app lifecycle, gestures, …). Depends only on `webdriver` + `remote`. |
| `.../chrome`, `.../firefox`, `.../log`, `.../sauce` | | Typed browser options, logging constants, and Sauce Labs support. |

Dependency graph: `remote → webdriver`; `selenium → webdriver + remote`;
`appium → webdriver + remote`.

## Install

```
go get github.com/prawdadigital/web-driver
```

You also need a running WebDriver server for browsers (a Selenium 4 JAR,
ChromeDriver, or GeckoDriver) or an Appium 2 server for mobile. The `selenium`
package can start a local server for you; see below.

## Usage

### Driving a browser with Selenium

```go
package main

import (
	"fmt"

	webdriver "github.com/prawdadigital/web-driver"
	"github.com/prawdadigital/web-driver/chrome"
	"github.com/prawdadigital/web-driver/selenium"
)

func main() {
	const port = 4444

	// Start a Selenium 4 server in the background (or use
	// selenium.NewChromeDriverService / NewGeckoDriverService).
	service, err := selenium.NewSeleniumService("vendor/selenium-server.jar", port)
	if err != nil {
		panic(err)
	}
	defer service.Stop()

	// Connect a WebDriver session to it.
	caps := webdriver.Capabilities{"browserName": "chrome"}
	caps.AddChrome(chrome.Capabilities{Args: []string{"--headless=new"}})

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	if err := wd.Get("https://pkg.go.dev"); err != nil {
		panic(err)
	}
	title, _ := wd.Title()
	fmt.Println(title)
}
```

If a server is already running, skip `NewSeleniumService` and call
`selenium.NewRemote` (or `remote.NewRemote`) against its URL directly.

### Driving a mobile device with Appium 2

```go
package main

import (
	"fmt"
	"time"

	"github.com/prawdadigital/web-driver/appium"
)

func main() {
	// Build capabilities; the "appium:" prefix is applied automatically.
	caps := appium.NewCapabilities().
		PlatformName("Android").
		AutomationName("UiAutomator2").
		DeviceName("Android Emulator").
		App("/path/to/app.apk").
		ToCapabilities()

	// Connect to a running Appium 2 server (root base path, no /wd/hub).
	driver, err := appium.NewRemote(caps, "http://127.0.0.1:4723")
	if err != nil {
		panic(err)
	}
	defer driver.Quit()

	// Mobile commands, alongside every standard WebDriver method.
	contexts, _ := driver.AvailableContexts()
	fmt.Println(contexts)
	_ = driver.Swipe(200, 800, 200, 200, 300*time.Millisecond) // swipe up
}
```

See the [package documentation](https://pkg.go.dev/github.com/prawdadigital/web-driver)
and [example_test.go](example_test.go) for more.

## Features

**Selenium 4 / W3C:** relative ("friendly") locators (`With(...).Above/Below/…`),
print page to PDF, shadow DOM (`GetShadowRoot`), new window/tab, W3C window rect,
and the full input Actions API including key, pointer (mouse/pen/touch), and
wheel/scroll actions.

**Appium 2:** the `appium:`-prefixing capability builder, contexts (native ↔
webview), orientation, geolocation, app lifecycle
(install/activate/terminate/remove/state), keyboard and device commands,
settings, W3C touch gestures (`Tap`, `Swipe`, `Zoom`, `Pinch`, …), and typed
wrappers over Appium `mobile:` gesture commands plus a generic `ExecuteMobile`.

## Downloading dependencies (for testing)

A helper downloads the Selenium 4 server JAR, ChromeDriver, GeckoDriver, and the
Sauce Connect proxy into `vendor/`:

```
cd vendor
go run init.go --alsologtostderr --download_browsers --download_latest
cd ..
```

Re-run periodically to refresh the binaries.

## Testing

```
go test ./...                                # all packages (browser tests skip if binaries are absent)
go test ./appium/                            # Appium client (mock HTTP server; no device needed)
go test . -run=TestSelenium4                 # Selenium 4 integration group (Chrome + Firefox; needs the JAR)
go test . -run=TestSelenium4/Chrome          # just the Chrome subgroup
go test . -skip TestFrameBuffer              # skip the Xvfb-only test (e.g. on macOS)
go test . --arg --help                       # list all test flags (driver/binary paths)
go test . --docker                           # run the suite hermetically inside Docker
```

Integration tests require the relevant binaries (Selenium JAR, browsers,
drivers); if they are not found the corresponding tests are skipped. Local
Selenium/HTMLUnit testing on Linux also needs `xvfb` and a JRE
(`sudo apt-get install xvfb openjdk-11-jre`). Run `gofmt -l .` before committing.

## Known issues

Most issues stem from the underlying browser automation framework rather than
this client. Notably:

- Headless Chrome does not support loading extensions
  ([crbug 706008](https://crbug.com/706008)).
- W3C wheel-scroll actions require a non-zero duration to take effect in headless
  Chrome.
- geckodriver/Firefox may occasionally fail session creation on a cold start
  ("Process unexpectedly closed"); it succeeds on retry.

Please [file an issue](https://github.com/prawdadigital/web-driver/issues/new) if
the client does not behave as expected.

## Contributing

Patches are welcome via pull request. Please add tests for non-trivial changes
and ensure `gofmt` has been run. A pre-commit hook is available:

```
ln -s ../../misc/git/pre-commit .git/hooks/pre-commit
```

## License

MIT — see [LICENSE](LICENSE). This project is a fork of
[tebeka/selenium](https://github.com/tebeka/selenium) by The Selenium Go Client
Authors.
