package appium

import (
	"encoding/json"
	"fmt"

	"github.com/prawdadigital/web-driver/selenium"
)

// Orientation values for Orientation/SetOrientation.
const (
	Portrait  = "PORTRAIT"
	Landscape = "LANDSCAPE"
)

// AppState enumerates the possible running states of an application, as
// returned by AppState.
type AppState int

// The possible application states.
const (
	AppNotInstalled                   AppState = 0
	AppNotRunning                     AppState = 1
	AppRunningInBackground            AppState = 2
	AppRunningInBackgroundOrSuspended AppState = 3
	AppRunningInForeground            AppState = 4
)

// GeoLocation is a geographical location used by Location and SetLocation.
type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

// Mobile is a WebDriver session augmented with the Appium mobile commands. It
// embeds selenium.WebDriver, so all standard WebDriver operations are also
// available.
type Mobile interface {
	selenium.WebDriver

	// AvailableContexts returns the list of contexts (e.g. "NATIVE_APP" and any
	// "WEBVIEW_*") available in the current session.
	AvailableContexts() ([]string, error)
	// CurrentContext returns the name of the currently selected context.
	CurrentContext() (string, error)
	// SwitchContext selects the named context, e.g. to move between the native
	// app and an embedded webview.
	SwitchContext(name string) error

	// Orientation returns the current screen orientation, PORTRAIT or LANDSCAPE.
	Orientation() (string, error)
	// SetOrientation sets the screen orientation to Portrait or Landscape.
	SetOrientation(orientation string) error

	// Location returns the device's current geolocation.
	Location() (*GeoLocation, error)
	// SetLocation sets the device's simulated geolocation.
	SetLocation(location GeoLocation) error

	// InstallApp installs the application at the given path or URL on the device.
	InstallApp(appPath string) error
	// IsAppInstalled reports whether an app with the given bundle/package ID is
	// installed.
	IsAppInstalled(bundleID string) (bool, error)
	// ActivateApp brings the app with the given ID to the foreground, launching
	// it if necessary.
	ActivateApp(appID string) error
	// TerminateApp terminates the app with the given ID and reports whether it
	// was running.
	TerminateApp(appID string) (bool, error)
	// RemoveApp uninstalls the app with the given ID and reports whether it was
	// present.
	RemoveApp(appID string) (bool, error)
	// AppState returns the running state of the app with the given ID.
	AppState(appID string) (AppState, error)

	// HideKeyboard hides the on-screen keyboard if it is shown.
	HideKeyboard() error
	// IsKeyboardShown reports whether the on-screen keyboard is currently shown.
	IsKeyboardShown() (bool, error)
	// PressKeyCode sends an Android key event. metaState and flags may be zero.
	PressKeyCode(keyCode, metaState, flags int) error
	// LongPressKeyCode sends a long-press Android key event.
	LongPressKeyCode(keyCode, metaState, flags int) error

	// Lock locks the screen for the given number of seconds; a value <= 0 locks
	// indefinitely until Unlock is called.
	Lock(seconds int) error
	// Unlock unlocks the screen.
	Unlock() error
	// IsLocked reports whether the screen is currently locked.
	IsLocked() (bool, error)
	// Shake simulates a device shake (iOS simulator / supported drivers).
	Shake() error
	// DeviceTime returns the device's current time as reported by the driver.
	DeviceTime() (string, error)

	// Settings returns the current Appium session settings.
	Settings() (map[string]interface{}, error)
	// UpdateSettings applies the given Appium session settings.
	UpdateSettings(settings map[string]interface{}) error
}

// mobileWD is the concrete Mobile implementation. It wraps a standard
// selenium.WebDriver and issues the mobile commands via ExecuteCommand.
type mobileWD struct {
	selenium.WebDriver
}

// NewRemote starts a new Appium session and returns a Mobile driver. urlPrefix
// is the base URL of the Appium server; for Appium 2 this is typically
// "http://127.0.0.1:4444" (Appium 2 dropped the "/wd/hub" base path that
// Appium 1 used by default).
func NewRemote(capabilities selenium.Capabilities, urlPrefix string) (Mobile, error) {
	wd, err := selenium.NewRemote(capabilities, urlPrefix)
	if err != nil {
		return nil, err
	}
	return &mobileWD{WebDriver: wd}, nil
}

// NewMobile wraps an existing WebDriver session as a Mobile driver, for callers
// that create the session themselves.
func NewMobile(wd selenium.WebDriver) Mobile {
	return &mobileWD{WebDriver: wd}
}

// value unmarshals the "value" field of a command response into out.
func (m *mobileWD) value(method, path string, params interface{}, out interface{}) error {
	response, err := m.ExecuteCommand(method, path, params)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	reply := struct {
		Value json.RawMessage `json:"value"`
	}{}
	if err := json.Unmarshal(response, &reply); err != nil {
		return err
	}
	return json.Unmarshal(reply.Value, out)
}

func (m *mobileWD) AvailableContexts() ([]string, error) {
	var contexts []string
	if err := m.value("GET", "/contexts", nil, &contexts); err != nil {
		return nil, err
	}
	return contexts, nil
}

func (m *mobileWD) CurrentContext() (string, error) {
	var context string
	if err := m.value("GET", "/context", nil, &context); err != nil {
		return "", err
	}
	return context, nil
}

func (m *mobileWD) SwitchContext(name string) error {
	return m.value("POST", "/context", map[string]interface{}{"name": name}, nil)
}

func (m *mobileWD) Orientation() (string, error) {
	var orientation string
	if err := m.value("GET", "/orientation", nil, &orientation); err != nil {
		return "", err
	}
	return orientation, nil
}

func (m *mobileWD) SetOrientation(orientation string) error {
	return m.value("POST", "/orientation", map[string]interface{}{"orientation": orientation}, nil)
}

func (m *mobileWD) Location() (*GeoLocation, error) {
	loc := new(GeoLocation)
	if err := m.value("GET", "/location", nil, loc); err != nil {
		return nil, err
	}
	return loc, nil
}

func (m *mobileWD) SetLocation(location GeoLocation) error {
	return m.value("POST", "/location", map[string]interface{}{"location": location}, nil)
}

func (m *mobileWD) InstallApp(appPath string) error {
	return m.value("POST", "/appium/device/install_app", map[string]interface{}{"appPath": appPath}, nil)
}

func (m *mobileWD) IsAppInstalled(bundleID string) (bool, error) {
	var installed bool
	if err := m.value("POST", "/appium/device/app_installed", map[string]interface{}{"bundleId": bundleID}, &installed); err != nil {
		return false, err
	}
	return installed, nil
}

func (m *mobileWD) ActivateApp(appID string) error {
	return m.value("POST", "/appium/device/activate_app", map[string]interface{}{"appId": appID}, nil)
}

func (m *mobileWD) TerminateApp(appID string) (bool, error) {
	var wasRunning bool
	if err := m.value("POST", "/appium/device/terminate_app", map[string]interface{}{"appId": appID}, &wasRunning); err != nil {
		return false, err
	}
	return wasRunning, nil
}

func (m *mobileWD) RemoveApp(appID string) (bool, error) {
	var removed bool
	if err := m.value("POST", "/appium/device/remove_app", map[string]interface{}{"appId": appID}, &removed); err != nil {
		return false, err
	}
	return removed, nil
}

func (m *mobileWD) AppState(appID string) (AppState, error) {
	var state int
	if err := m.value("POST", "/appium/device/app_state", map[string]interface{}{"appId": appID}, &state); err != nil {
		return 0, err
	}
	return AppState(state), nil
}

func (m *mobileWD) HideKeyboard() error {
	return m.value("POST", "/appium/device/hide_keyboard", map[string]interface{}{}, nil)
}

func (m *mobileWD) IsKeyboardShown() (bool, error) {
	var shown bool
	if err := m.value("GET", "/appium/device/is_keyboard_shown", nil, &shown); err != nil {
		return false, err
	}
	return shown, nil
}

func (m *mobileWD) PressKeyCode(keyCode, metaState, flags int) error {
	return m.value("POST", "/appium/device/press_keycode", keyCodeParams(keyCode, metaState, flags), nil)
}

func (m *mobileWD) LongPressKeyCode(keyCode, metaState, flags int) error {
	return m.value("POST", "/appium/device/long_press_keycode", keyCodeParams(keyCode, metaState, flags), nil)
}

func keyCodeParams(keyCode, metaState, flags int) map[string]interface{} {
	return map[string]interface{}{
		"keycode":   keyCode,
		"metastate": metaState,
		"flags":     flags,
	}
}

func (m *mobileWD) Lock(seconds int) error {
	return m.value("POST", "/appium/device/lock", map[string]interface{}{"seconds": seconds}, nil)
}

func (m *mobileWD) Unlock() error {
	return m.value("POST", "/appium/device/unlock", map[string]interface{}{}, nil)
}

func (m *mobileWD) IsLocked() (bool, error) {
	var locked bool
	if err := m.value("POST", "/appium/device/is_locked", map[string]interface{}{}, &locked); err != nil {
		return false, err
	}
	return locked, nil
}

func (m *mobileWD) Shake() error {
	return m.value("POST", "/appium/device/shake", map[string]interface{}{}, nil)
}

func (m *mobileWD) DeviceTime() (string, error) {
	var t string
	if err := m.value("GET", "/appium/device/system_time", nil, &t); err != nil {
		return "", err
	}
	return t, nil
}

func (m *mobileWD) Settings() (map[string]interface{}, error) {
	settings := map[string]interface{}{}
	if err := m.value("GET", "/appium/settings", nil, &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (m *mobileWD) UpdateSettings(settings map[string]interface{}) error {
	if settings == nil {
		return fmt.Errorf("appium: settings must not be nil")
	}
	return m.value("POST", "/appium/settings", map[string]interface{}{"settings": settings}, nil)
}
