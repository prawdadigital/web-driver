package appium

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	webdriver "github.com/prawdadigital/web-driver"
)

// recordedRequest captures the method, path and decoded JSON body of a request
// received by the mock Appium server.
type recordedRequest struct {
	method string
	path   string
	body   map[string]interface{}
}

// newMockServer returns an httptest server that behaves like a minimal Appium 2
// endpoint: it creates a W3C session and returns canned "value" responses for
// mobile commands. Received requests are appended to *reqs.
func newMockServer(t *testing.T, reqs *[]recordedRequest) *httptest.Server {
	values := map[string]interface{}{
		"/session/sess-1/contexts":                        []string{"NATIVE_APP", "WEBVIEW_1"},
		"/session/sess-1/context":                         "NATIVE_APP",
		"/session/sess-1/orientation":                     "PORTRAIT",
		"/session/sess-1/location":                        GeoLocation{Latitude: 1.5, Longitude: 2.5, Altitude: 3.5},
		"/session/sess-1/appium/device/app_installed":     true,
		"/session/sess-1/appium/device/app_state":         4,
		"/session/sess-1/appium/device/terminate_app":     true,
		"/session/sess-1/appium/device/remove_app":        true,
		"/session/sess-1/appium/device/is_keyboard_shown": true,
		"/session/sess-1/appium/device/is_locked":         false,
		"/session/sess-1/appium/device/system_time":       "2026-08-10T12:00:00+00:00",
		"/session/sess-1/appium/settings":                 map[string]interface{}{"ignoreUnimportantViews": true},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/session" && r.Method == "POST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"value": map[string]interface{}{
					"sessionId":    "sess-1",
					"capabilities": map[string]interface{}{},
				},
			})
			return
		}

		body := map[string]interface{}{}
		if raw, _ := ioutil.ReadAll(r.Body); len(raw) > 0 {
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Errorf("bad request body for %s %s: %v", r.Method, r.URL.Path, err)
			}
		}
		*reqs = append(*reqs, recordedRequest{method: r.Method, path: r.URL.Path, body: body})

		v, ok := values[r.URL.Path]
		if !ok {
			v = nil
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"value": v})
	}))
}

func newTestMobile(t *testing.T, reqs *[]recordedRequest) (Mobile, func()) {
	srv := newMockServer(t, reqs)
	m, err := NewRemote(webdriver.Capabilities{"platformName": "Android"}, srv.URL)
	if err != nil {
		srv.Close()
		t.Fatalf("NewRemote: %v", err)
	}
	return m, srv.Close
}

func TestContexts(t *testing.T) {
	var reqs []recordedRequest
	m, cleanup := newTestMobile(t, &reqs)
	defer cleanup()

	contexts, err := m.AvailableContexts()
	if err != nil {
		t.Fatalf("AvailableContexts: %v", err)
	}
	if want := []string{"NATIVE_APP", "WEBVIEW_1"}; !reflect.DeepEqual(contexts, want) {
		t.Errorf("AvailableContexts = %v, want %v", contexts, want)
	}

	cur, err := m.CurrentContext()
	if err != nil {
		t.Fatalf("CurrentContext: %v", err)
	}
	if cur != "NATIVE_APP" {
		t.Errorf("CurrentContext = %q, want %q", cur, "NATIVE_APP")
	}

	if err := m.SwitchContext("WEBVIEW_1"); err != nil {
		t.Fatalf("SwitchContext: %v", err)
	}
	last := reqs[len(reqs)-1]
	if last.method != "POST" || last.path != "/session/sess-1/context" {
		t.Errorf("SwitchContext request = %s %s, want POST /session/sess-1/context", last.method, last.path)
	}
	if last.body["name"] != "WEBVIEW_1" {
		t.Errorf("SwitchContext body = %v, want name=WEBVIEW_1", last.body)
	}
}

func TestOrientationAndLocation(t *testing.T) {
	var reqs []recordedRequest
	m, cleanup := newTestMobile(t, &reqs)
	defer cleanup()

	if err := m.SetOrientation(Landscape); err != nil {
		t.Fatalf("SetOrientation: %v", err)
	}
	last := reqs[len(reqs)-1]
	if last.path != "/session/sess-1/orientation" || last.body["orientation"] != "LANDSCAPE" {
		t.Errorf("SetOrientation request = %s %v", last.path, last.body)
	}

	loc, err := m.Location()
	if err != nil {
		t.Fatalf("Location: %v", err)
	}
	if loc.Latitude != 1.5 || loc.Longitude != 2.5 || loc.Altitude != 3.5 {
		t.Errorf("Location = %+v, want {1.5 2.5 3.5}", *loc)
	}

	if err := m.SetLocation(GeoLocation{Latitude: 10, Longitude: 20, Altitude: 30}); err != nil {
		t.Fatalf("SetLocation: %v", err)
	}
	last = reqs[len(reqs)-1]
	nested, ok := last.body["location"].(map[string]interface{})
	if !ok {
		t.Fatalf("SetLocation body missing nested location: %v", last.body)
	}
	if nested["latitude"].(float64) != 10 || nested["longitude"].(float64) != 20 {
		t.Errorf("SetLocation nested = %v", nested)
	}
}

func TestAppLifecycle(t *testing.T) {
	var reqs []recordedRequest
	m, cleanup := newTestMobile(t, &reqs)
	defer cleanup()

	if err := m.InstallApp("/tmp/app.apk"); err != nil {
		t.Fatalf("InstallApp: %v", err)
	}
	last := reqs[len(reqs)-1]
	if last.path != "/session/sess-1/appium/device/install_app" || last.body["appPath"] != "/tmp/app.apk" {
		t.Errorf("InstallApp request = %s %v", last.path, last.body)
	}

	installed, err := m.IsAppInstalled("com.example.app")
	if err != nil {
		t.Fatalf("IsAppInstalled: %v", err)
	}
	if !installed {
		t.Errorf("IsAppInstalled = false, want true")
	}
	last = reqs[len(reqs)-1]
	if last.body["bundleId"] != "com.example.app" {
		t.Errorf("IsAppInstalled body = %v, want bundleId=com.example.app", last.body)
	}

	state, err := m.AppState("com.example.app")
	if err != nil {
		t.Fatalf("AppState: %v", err)
	}
	if state != AppRunningInForeground {
		t.Errorf("AppState = %d, want %d", state, AppRunningInForeground)
	}

	wasRunning, err := m.TerminateApp("com.example.app")
	if err != nil {
		t.Fatalf("TerminateApp: %v", err)
	}
	if !wasRunning {
		t.Errorf("TerminateApp = false, want true")
	}
}

func TestDeviceAndSettings(t *testing.T) {
	var reqs []recordedRequest
	m, cleanup := newTestMobile(t, &reqs)
	defer cleanup()

	shown, err := m.IsKeyboardShown()
	if err != nil {
		t.Fatalf("IsKeyboardShown: %v", err)
	}
	if !shown {
		t.Errorf("IsKeyboardShown = false, want true")
	}

	if err := m.PressKeyCode(66, 0, 0); err != nil {
		t.Fatalf("PressKeyCode: %v", err)
	}
	last := reqs[len(reqs)-1]
	if last.path != "/session/sess-1/appium/device/press_keycode" {
		t.Errorf("PressKeyCode path = %s", last.path)
	}
	if last.body["keycode"].(float64) != 66 {
		t.Errorf("PressKeyCode keycode = %v, want 66", last.body["keycode"])
	}

	if err := m.Lock(5); err != nil {
		t.Fatalf("Lock: %v", err)
	}
	last = reqs[len(reqs)-1]
	if last.body["seconds"].(float64) != 5 {
		t.Errorf("Lock seconds = %v, want 5", last.body["seconds"])
	}

	dt, err := m.DeviceTime()
	if err != nil {
		t.Fatalf("DeviceTime: %v", err)
	}
	if dt != "2026-08-10T12:00:00+00:00" {
		t.Errorf("DeviceTime = %q", dt)
	}

	settings, err := m.Settings()
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if settings["ignoreUnimportantViews"] != true {
		t.Errorf("Settings = %v", settings)
	}

	if err := m.UpdateSettings(map[string]interface{}{"foo": "bar"}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	last = reqs[len(reqs)-1]
	nested, ok := last.body["settings"].(map[string]interface{})
	if !ok || nested["foo"] != "bar" {
		t.Errorf("UpdateSettings body = %v, want settings.foo=bar", last.body)
	}
}

func TestCapabilitiesPrefix(t *testing.T) {
	caps := NewCapabilities().
		PlatformName("Android").
		AutomationName("UiAutomator2").
		DeviceName("Pixel_7").
		App("/tmp/app.apk").
		Set("appPackage", "com.example").
		Set("goog:chromeOptions", map[string]interface{}{"w3c": true}).
		ToCapabilities()

	cases := map[string]interface{}{
		"platformName":          "Android",
		"appium:automationName": "UiAutomator2",
		"appium:deviceName":     "Pixel_7",
		"appium:app":            "/tmp/app.apk",
		"appium:appPackage":     "com.example",
	}
	for k, want := range cases {
		if caps[k] != want {
			t.Errorf("caps[%q] = %v, want %v", k, caps[k], want)
		}
	}
	// A key that already carries a vendor prefix must not be double-prefixed.
	if _, ok := caps["goog:chromeOptions"]; !ok {
		t.Errorf("caps missing goog:chromeOptions; keys=%v", caps)
	}
	if _, ok := caps["appium:goog:chromeOptions"]; ok {
		t.Errorf("caps double-prefixed a vendor capability")
	}
}

func TestGestures(t *testing.T) {
	var reqs []recordedRequest
	m, cleanup := newTestMobile(t, &reqs)
	defer cleanup()

	// lastActionSources returns the input-source list of the most recent POST
	// to the /actions endpoint.
	lastActionSources := func() []interface{} {
		last := reqs[len(reqs)-1]
		if last.method != "POST" || last.path != "/session/sess-1/actions" {
			t.Fatalf("expected POST /session/sess-1/actions, got %s %s", last.method, last.path)
		}
		srcs, ok := last.body["actions"].([]interface{})
		if !ok {
			t.Fatalf("actions payload missing 'actions' array: %v", last.body)
		}
		return srcs
	}
	actionTypes := func(src interface{}) []string {
		var types []string
		for _, a := range src.(map[string]interface{})["actions"].([]interface{}) {
			types = append(types, a.(map[string]interface{})["type"].(string))
		}
		return types
	}

	// Tap: one touch pointer performing move/down/pause/up.
	if err := m.Tap(10, 20); err != nil {
		t.Fatalf("Tap: %v", err)
	}
	srcs := lastActionSources()
	if len(srcs) != 1 {
		t.Fatalf("Tap: got %d input sources, want 1", len(srcs))
	}
	src0 := srcs[0].(map[string]interface{})
	if src0["type"] != "pointer" {
		t.Errorf("Tap source type = %v, want pointer", src0["type"])
	}
	if pt := src0["parameters"].(map[string]interface{})["pointerType"]; pt != "touch" {
		t.Errorf("Tap pointerType = %v, want touch", pt)
	}
	if got, want := actionTypes(srcs[0]), []string{"pointerMove", "pointerDown", "pause", "pointerUp"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Tap action types = %v, want %v", got, want)
	}

	// Swipe: move/down/move/up.
	if err := m.Swipe(0, 0, 100, 200, 200*time.Millisecond); err != nil {
		t.Fatalf("Swipe: %v", err)
	}
	if got, want := actionTypes(lastActionSources()[0]), []string{"pointerMove", "pointerDown", "pointerMove", "pointerUp"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Swipe action types = %v, want %v", got, want)
	}

	// Zoom: two touch pointers performed together.
	if err := m.Zoom(50, 50, 40, 200*time.Millisecond); err != nil {
		t.Fatalf("Zoom: %v", err)
	}
	if srcs := lastActionSources(); len(srcs) != 2 {
		t.Errorf("Zoom: got %d input sources, want 2", len(srcs))
	}
}
