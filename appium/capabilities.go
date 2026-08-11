// Package appium provides an Appium 2 client built on top of the webdriver
// WebDriver client. Appium 2 speaks the W3C WebDriver protocol with vendor
// extensions, and requires all non-standard capabilities to carry the "appium:"
// prefix.
//
// The same client drives native mobile apps (Android, iOS), mobile browsers,
// and native desktop apps (Windows via the Windows driver, macOS via the Mac2
// driver): only the capabilities and the Appium driver differ. Use the Platform
// and Automation constants below to select a driver, and ExecuteExtension to
// invoke driver-specific "mobile:"/"windows:"/"macos:" commands.
package appium

import (
	"strings"

	webdriver "github.com/prawdadigital/web-driver"
)

// Platform names for the standard platformName capability.
const (
	PlatformAndroid = "Android"
	PlatformIOS     = "iOS"
	PlatformTVOS    = "tvOS"
	PlatformWindows = "Windows"
	PlatformMac     = "Mac"
)

// Automation driver names for the appium:automationName capability. Each names
// an Appium driver that must be installed on the server (e.g.
// "appium driver install uiautomator2").
const (
	AutomationUiAutomator2 = "UiAutomator2" // Android
	AutomationEspresso     = "Espresso"     // Android
	AutomationXCUITest     = "XCUITest"     // iOS / tvOS
	AutomationMac2         = "Mac2"         // macOS desktop apps
	AutomationWindows      = "Windows"      // Windows desktop apps
	AutomationGecko        = "Gecko"        // Firefox (mobile/desktop)
	AutomationSafari       = "Safari"       // Safari
)

// standardCapabilities are the W3C top-level capability names that must NOT be
// given the "appium:" prefix.
var standardCapabilities = map[string]bool{
	"browserName":             true,
	"browserVersion":          true,
	"platformName":            true,
	"acceptInsecureCerts":     true,
	"pageLoadStrategy":        true,
	"proxy":                   true,
	"setWindowRect":           true,
	"timeouts":                true,
	"unhandledPromptBehavior": true,
}

// Capabilities is a fluent builder for Appium session capabilities. It applies
// the "appium:" prefix required by Appium 2 to any non-standard capability, so
// callers may pass either "deviceName" or "appium:deviceName".
type Capabilities struct {
	caps webdriver.Capabilities
}

// NewCapabilities returns an empty Appium capabilities builder.
func NewCapabilities() *Capabilities {
	return &Capabilities{caps: webdriver.Capabilities{}}
}

// Set assigns a capability, adding the "appium:" prefix unless key is a
// standard W3C capability or already contains a vendor prefix (a ":").
func (c *Capabilities) Set(key string, value interface{}) *Capabilities {
	if !standardCapabilities[key] && !strings.Contains(key, ":") {
		key = "appium:" + key
	}
	c.caps[key] = value
	return c
}

// PlatformName sets the standard "platformName" capability, e.g. "Android" or
// "iOS".
func (c *Capabilities) PlatformName(name string) *Capabilities {
	c.caps["platformName"] = name
	return c
}

// AutomationName sets "appium:automationName", e.g. "UiAutomator2" or
// "XCUITest".
func (c *Capabilities) AutomationName(name string) *Capabilities {
	return c.Set("automationName", name)
}

// DeviceName sets "appium:deviceName", e.g. "Android Emulator".
func (c *Capabilities) DeviceName(name string) *Capabilities {
	return c.Set("deviceName", name)
}

// PlatformVersion sets "appium:platformVersion", e.g. "14.0".
func (c *Capabilities) PlatformVersion(version string) *Capabilities {
	return c.Set("platformVersion", version)
}

// App sets "appium:app", the path or URL to the application under test. For the
// Windows driver this may also be an Application User Model ID or "Root" for the
// desktop.
func (c *Capabilities) App(app string) *Capabilities {
	return c.Set("app", app)
}

// BundleID sets "appium:bundleId", identifying an installed application to
// launch on iOS (XCUITest) or macOS (Mac2), e.g. "com.apple.TextEdit".
func (c *Capabilities) BundleID(id string) *Capabilities {
	return c.Set("bundleId", id)
}

// AppPackage sets "appium:appPackage", the Android application package to launch.
func (c *Capabilities) AppPackage(pkg string) *Capabilities {
	return c.Set("appPackage", pkg)
}

// AppActivity sets "appium:appActivity", the Android activity to launch.
func (c *Capabilities) AppActivity(activity string) *Capabilities {
	return c.Set("appActivity", activity)
}

// BrowserName sets the standard "browserName" capability, for driving a mobile
// browser rather than an app.
func (c *Capabilities) BrowserName(name string) *Capabilities {
	c.caps["browserName"] = name
	return c
}

// ToCapabilities returns the underlying webdriver.Capabilities, suitable for
// passing to NewRemote.
func (c *Capabilities) ToCapabilities() webdriver.Capabilities {
	return c.caps
}
