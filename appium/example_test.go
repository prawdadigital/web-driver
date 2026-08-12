package appium_test

import (
	"github.com/prawdadigital/web-driver/appium"
)

// Example drives a native Android application with the UiAutomator2 driver.
func Example() {
	caps := appium.NewCapabilities().
		PlatformName(appium.PlatformAndroid).
		AutomationName(appium.AutomationUiAutomator2).
		DeviceName("Android Emulator").
		App("/path/to/app.apk").
		ToCapabilities()

	driver, err := appium.NewRemote(caps, "http://127.0.0.1:4723")
	if err != nil {
		panic(err)
	}
	defer driver.Quit()

	el, err := driver.FindElement(appium.ByAccessibilityID, "login")
	if err != nil {
		panic(err)
	}
	el.Click()
}

// Example_windowsDesktop drives a native Windows application through the Appium
// Windows driver. Driver-specific commands are issued with ExecuteExtension
// using the "windows:" prefix.
func Example_windowsDesktop() {
	caps := appium.NewCapabilities().
		PlatformName(appium.PlatformWindows).
		AutomationName(appium.AutomationWindows).
		App(`C:\Windows\System32\notepad.exe`).
		ToCapabilities()

	// appium.Driver is an alias for appium.Mobile that reads naturally for a
	// desktop session.
	var driver appium.Driver
	driver, err := appium.NewRemote(caps, "http://127.0.0.1:4723")
	if err != nil {
		panic(err)
	}
	defer driver.Quit()

	_, _ = driver.ExecuteExtension("windows: click", map[string]interface{}{"x": 100, "y": 200})
}

// Example_macDesktop drives a native macOS application through the Appium Mac2
// driver, launching it by bundle identifier.
func Example_macDesktop() {
	caps := appium.NewCapabilities().
		PlatformName(appium.PlatformMac).
		AutomationName(appium.AutomationMac2).
		BundleID("com.apple.TextEdit").
		ToCapabilities()

	driver, err := appium.NewRemote(caps, "http://127.0.0.1:4723")
	if err != nil {
		panic(err)
	}
	defer driver.Quit()

	_, _ = driver.ExecuteExtension("macos: activateApp", map[string]interface{}{"bundleId": "com.apple.TextEdit"})
}
