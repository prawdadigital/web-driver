// Package appium provides an Appium 2 client built on top of the webdriver
// WebDriver client. Appium 2 speaks the W3C WebDriver protocol with mobile
// extensions, and requires all non-standard capabilities to carry the
// "appium:" prefix.
package appium

import (
	"strings"

	"github.com/prawdadigital/web-driver"
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

// App sets "appium:app", the path or URL to the application under test.
func (c *Capabilities) App(app string) *Capabilities {
	return c.Set("app", app)
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
