package appium

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// w3cElementKey is the W3C JSON key identifying a web/native element.
const w3cElementKey = "element-6066-11e4-a52e-4f735466cecf"

// newScenarioServer is a mock Appium endpoint for native/desktop flow tests. It
// creates a session reporting the given granted capabilities, returns a canned
// element for finds and canned values for app-lifecycle commands, and records
// every request into reqs.
func newScenarioServer(t *testing.T, granted map[string]interface{}, reqs *[]recordedRequest) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/session" && r.Method == "POST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"value": map[string]interface{}{
					"sessionId":    "sess-1",
					"capabilities": granted,
				},
			})
			return
		}

		body := map[string]interface{}{}
		if raw, _ := ioutil.ReadAll(r.Body); len(raw) > 0 {
			json.Unmarshal(raw, &body)
		}
		*reqs = append(*reqs, recordedRequest{method: r.Method, path: r.URL.Path, body: body})

		var value interface{}
		switch r.URL.Path {
		case "/session/sess-1/element":
			value = map[string]interface{}{w3cElementKey: "el-1"}
		case "/session/sess-1/appium/device/terminate_app":
			value = true
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"value": value})
	}))
}

func lastReq(reqs []recordedRequest) recordedRequest { return reqs[len(reqs)-1] }

func TestAndroidNativeFlow(t *testing.T) {
	var reqs []recordedRequest
	srv := newScenarioServer(t, map[string]interface{}{
		"platformName":          "Android",
		"appium:automationName": "UiAutomator2",
	}, &reqs)
	defer srv.Close()

	caps := NewCapabilities().
		PlatformName(PlatformAndroid).
		AutomationName(AutomationUiAutomator2).
		AppPackage("com.example").
		AppActivity(".MainActivity").
		ToCapabilities()

	driver, err := NewRemote(caps, srv.URL)
	if err != nil {
		t.Fatalf("NewRemote: %v", err)
	}
	defer driver.Quit()

	// Capabilities() reports the granted set (no legacy get-capabilities call).
	got, err := driver.Capabilities()
	if err != nil {
		t.Fatalf("Capabilities: %v", err)
	}
	if got["platformName"] != "Android" || got["appium:automationName"] != "UiAutomator2" {
		t.Errorf("Capabilities = %v", got)
	}

	// Find by accessibility id sends the Appium location strategy verbatim.
	el, err := driver.FindElement(ByAccessibilityID, "login")
	if err != nil {
		t.Fatalf("FindElement: %v", err)
	}
	find := lastReq(reqs)
	if find.body["using"] != "accessibility id" || find.body["value"] != "login" {
		t.Errorf("find request = %v, want using=accessibility id value=login", find.body)
	}
	if err := el.Click(); err != nil {
		t.Fatalf("Click: %v", err)
	}
	if last := lastReq(reqs); last.method != "POST" || last.path != "/session/sess-1/element/el-1/click" {
		t.Errorf("click request = %s %s", last.method, last.path)
	}

	// App lifecycle on the native app.
	if err := driver.ActivateApp("com.example"); err != nil {
		t.Fatalf("ActivateApp: %v", err)
	}
	if last := lastReq(reqs); last.path != "/session/sess-1/appium/device/activate_app" || last.body["appId"] != "com.example" {
		t.Errorf("ActivateApp request = %s %v", last.path, last.body)
	}
	if _, err := driver.TerminateApp("com.example"); err != nil {
		t.Fatalf("TerminateApp: %v", err)
	}
}

func TestWindowsDesktopFlow(t *testing.T) {
	var reqs []recordedRequest
	srv := newScenarioServer(t, map[string]interface{}{
		"platformName":          "Windows",
		"appium:automationName": "Windows",
	}, &reqs)
	defer srv.Close()

	caps := NewCapabilities().
		PlatformName(PlatformWindows).
		AutomationName(AutomationWindows).
		App(`C:\Windows\System32\notepad.exe`).
		ToCapabilities()

	// A desktop session reads naturally as an appium.Driver (alias of Mobile).
	var driver Driver
	driver, err := NewRemote(caps, srv.URL)
	if err != nil {
		t.Fatalf("NewRemote: %v", err)
	}
	defer driver.Quit()

	if got, _ := driver.Capabilities(); got["platformName"] != "Windows" {
		t.Errorf("Capabilities = %v, want platformName=Windows", got)
	}

	// The Windows driver uses the "windows:" extension prefix.
	if _, err := driver.ExecuteExtension("windows: click", map[string]interface{}{"x": 10, "y": 20}); err != nil {
		t.Fatalf("ExecuteExtension(windows:): %v", err)
	}
	last := lastReq(reqs)
	if last.path != "/session/sess-1/execute/sync" || last.body["script"] != "windows: click" {
		t.Errorf("windows exec = %s %v", last.path, last.body)
	}
}

func TestMacDesktopFlow(t *testing.T) {
	var reqs []recordedRequest
	srv := newScenarioServer(t, map[string]interface{}{
		"platformName":          "Mac",
		"appium:automationName": "Mac2",
	}, &reqs)
	defer srv.Close()

	caps := NewCapabilities().
		PlatformName(PlatformMac).
		AutomationName(AutomationMac2).
		BundleID("com.apple.TextEdit").
		ToCapabilities()

	driver, err := NewRemote(caps, srv.URL)
	if err != nil {
		t.Fatalf("NewRemote: %v", err)
	}
	defer driver.Quit()

	// The Mac2 driver uses the "macos:" extension prefix.
	if _, err := driver.ExecuteExtension("macos: launchApp", map[string]interface{}{"bundleId": "com.apple.TextEdit"}); err != nil {
		t.Fatalf("ExecuteExtension(macos:): %v", err)
	}
	if last := lastReq(reqs); last.body["script"] != "macos: launchApp" {
		t.Errorf("macos exec script = %v", last.body["script"])
	}

	// The iOS/Mac class-chain locator passes through on finds.
	if _, err := driver.FindElement(ByIOSClassChain, "**/XCUIElementTypeButton[`label == \"OK\"`]"); err != nil {
		t.Fatalf("FindElement(class chain): %v", err)
	}
	if find := lastReq(reqs); find.body["using"] != "-ios class chain" {
		t.Errorf("class-chain find using = %v, want -ios class chain", find.body["using"])
	}
}
