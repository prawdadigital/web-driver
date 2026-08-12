package remote

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	webdriver "github.com/prawdadigital/web-driver"
	"github.com/prawdadigital/web-driver/chrome"
)

// recordedRequest captures a request received by the mock W3C server.
type recordedRequest struct {
	method string
	path   string
	body   map[string]interface{}
}

// mockServer is a minimal W3C WebDriver endpoint for exercising the transport.
// It creates a session, records requests, and returns canned "value" responses
// keyed by request path. A handler in errs makes a path return a W3C error.
type mockServer struct {
	srv    *httptest.Server
	values map[string]interface{}
	errs   map[string]interface{} // path -> value payload of an error response
	reqs   []recordedRequest
}

func newMockServer(t *testing.T) *mockServer {
	m := &mockServer{
		values: map[string]interface{}{},
		errs:   map[string]interface{}{},
	}
	m.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/session" && r.Method == "POST" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"value": map[string]interface{}{
					"sessionId": "sess-1",
					"capabilities": map[string]interface{}{
						"browserName":    "chrome",
						"browserVersion": "120.0.1",
					},
				},
			})
			return
		}

		body := map[string]interface{}{}
		if raw, _ := ioutil.ReadAll(r.Body); len(raw) > 0 {
			json.Unmarshal(raw, &body)
		}
		m.reqs = append(m.reqs, recordedRequest{method: r.Method, path: r.URL.Path, body: body})

		if errVal, ok := m.errs[r.URL.Path]; ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"value": errVal})
			return
		}

		v := m.values[r.URL.Path]
		json.NewEncoder(w).Encode(map[string]interface{}{"value": v})
	}))
	return m
}

func (m *mockServer) close()                { m.srv.Close() }
func (m *mockServer) last() recordedRequest { return m.reqs[len(m.reqs)-1] }

// element is the W3C JSON representation of a web element with the given id.
func element(id string) map[string]interface{} {
	return map[string]interface{}{webElementIdentifier: id}
}

func newTestDriver(t *testing.T, m *mockServer) webdriver.WebDriver {
	wd, err := NewRemote(webdriver.Capabilities{"browserName": "chrome"}, m.srv.URL)
	if err != nil {
		t.Fatalf("NewRemote returned error: %v", err)
	}
	return wd
}

func TestNewSessionNegotiatesW3C(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	if wd.SessionID() != "sess-1" {
		t.Errorf("SessionID = %q, want sess-1", wd.SessionID())
	}
	// The concrete driver should record the negotiated W3C mode and version.
	rwd := wd.(*remoteWD)
	if !rwd.w3cCompatible {
		t.Errorf("w3cCompatible = false, want true")
	}
	if got := rwd.browserVersion.String(); got != "120.0.1" {
		t.Errorf("browserVersion = %q, want 120.0.1", got)
	}
}

func TestNavigationAndPageMethods(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/url"] = "http://example.com/"
	m.values["/session/sess-1/title"] = "Example"
	m.values["/session/sess-1/source"] = "<html></html>"

	wd := newTestDriver(t, m)

	if err := wd.Get("http://example.com"); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/url" || last.body["url"] != "http://example.com" {
		t.Errorf("Get request = %s %s %v", last.method, last.path, last.body)
	}

	if u, err := wd.CurrentURL(); err != nil || u != "http://example.com/" {
		t.Errorf("CurrentURL = %q, %v", u, err)
	}
	if title, err := wd.Title(); err != nil || title != "Example" {
		t.Errorf("Title = %q, %v", title, err)
	}
	if src, err := wd.PageSource(); err != nil || src != "<html></html>" {
		t.Errorf("PageSource = %q, %v", src, err)
	}
}

func TestFindElementAndText(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")
	m.values["/session/sess-1/elements"] = []interface{}{element("e1"), element("e2")}
	m.values["/session/sess-1/element/e1/text"] = "hello"
	m.values["/session/sess-1/element/e1/attribute/id"] = "greeting"

	wd := newTestDriver(t, m)

	el, err := wd.FindElement(webdriver.ByCSSSelector, "#greeting")
	if err != nil {
		t.Fatalf("FindElement returned error: %v", err)
	}
	if txt, err := el.Text(); err != nil || txt != "hello" {
		t.Errorf("Text = %q, %v", txt, err)
	}
	if attr, err := el.GetAttribute("id"); err != nil || attr != "greeting" {
		t.Errorf("GetAttribute = %q, %v", attr, err)
	}

	els, err := wd.FindElements(webdriver.ByCSSSelector, "div")
	if err != nil {
		t.Fatalf("FindElements returned error: %v", err)
	}
	if len(els) != 2 {
		t.Errorf("FindElements returned %d elements, want 2", len(els))
	}
}

func TestByIDRewriting(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")

	wd := newTestDriver(t, m)
	if _, err := wd.FindElement(webdriver.ByID, "greeting"); err != nil {
		t.Fatalf("FindElement(ByID) returned error: %v", err)
	}
	// W3C has no "id" strategy, so it must be rewritten to a CSS selector.
	last := m.last()
	if last.body["using"] != webdriver.ByCSSSelector || last.body["value"] != "#greeting" {
		t.Errorf("ByID rewrite: using=%v value=%v, want css/#greeting", last.body["using"], last.body["value"])
	}
}

func TestExecuteScript(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/execute/sync"] = float64(42)

	wd := newTestDriver(t, m)
	res, err := wd.ExecuteScript("return 42;", []interface{}{1, "two"})
	if err != nil {
		t.Fatalf("ExecuteScript returned error: %v", err)
	}
	if res.(float64) != 42 {
		t.Errorf("ExecuteScript = %v, want 42", res)
	}
	last := m.last()
	if last.path != "/session/sess-1/execute/sync" || last.body["script"] != "return 42;" {
		t.Errorf("ExecuteScript request = %s %v", last.path, last.body)
	}
}

func TestActionsPayload(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	wd.StorePointerActions("finger", webdriver.TouchPointer,
		webdriver.PointerDownAction(webdriver.LeftButton),
		webdriver.PointerUpAction(webdriver.LeftButton),
	)
	wd.StoreWheelActions("wheel", webdriver.ScrollAction(0, webdriver.FromViewport, 0, 0, 0, 100))
	if err := wd.PerformActions(); err != nil {
		t.Fatalf("PerformActions returned error: %v", err)
	}
	last := m.last()
	if last.path != "/session/sess-1/actions" {
		t.Fatalf("PerformActions path = %s", last.path)
	}
	srcs := last.body["actions"].([]interface{})
	if len(srcs) != 2 {
		t.Fatalf("got %d input sources, want 2 (pointer + wheel)", len(srcs))
	}
	if srcs[0].(map[string]interface{})["type"] != "pointer" || srcs[1].(map[string]interface{})["type"] != "wheel" {
		t.Errorf("input source types = %v", srcs)
	}
}

func TestExecuteCommand(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/appium/device/shake"] = nil

	wd := newTestDriver(t, m)
	if _, err := wd.ExecuteCommand("POST", "/appium/device/shake", map[string]interface{}{"x": 1}); err != nil {
		t.Fatalf("ExecuteCommand returned error: %v", err)
	}
	last := m.last()
	if last.method != "POST" || last.path != "/session/sess-1/appium/device/shake" || last.body["x"].(float64) != 1 {
		t.Errorf("ExecuteCommand request = %s %s %v", last.method, last.path, last.body)
	}
}

func TestWindowRect(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/window/rect"] = map[string]interface{}{"x": 1, "y": 2, "width": 800, "height": 600}

	wd := newTestDriver(t, m)
	if err := wd.SetWindowRect(webdriver.Rect{X: 1, Y: 2, Width: 800, Height: 600}); err != nil {
		t.Fatalf("SetWindowRect returned error: %v", err)
	}
	r, err := wd.GetWindowRect()
	if err != nil {
		t.Fatalf("GetWindowRect returned error: %v", err)
	}
	if r.Width != 800 || r.Height != 600 {
		t.Errorf("GetWindowRect = %+v, want 800x600", *r)
	}
}

func TestGetCookies(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/cookie"] = []interface{}{
		map[string]interface{}{"name": "sid", "value": "abc", "domain": "example.com"},
	}

	wd := newTestDriver(t, m)
	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatalf("GetCookies returned error: %v", err)
	}
	if len(cookies) != 1 || cookies[0].Name != "sid" || cookies[0].Value != "abc" {
		t.Errorf("GetCookies = %+v", cookies)
	}
}

// TestW3CErrorWithArrayStacktrace is a regression test: Selenium 4 returns the
// "stacktrace" field as a JSON array for some errors, which must still be
// surfaced as a *webdriver.Error rather than silently swallowed.
func TestW3CErrorWithArrayStacktrace(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.errs["/session/sess-1/url"] = map[string]interface{}{
		"error":      "unknown error",
		"message":    "something failed",
		"stacktrace": []interface{}{map[string]interface{}{"line": 1}},
	}

	wd := newTestDriver(t, m)
	err := wd.Get("http://example.com")
	if err == nil {
		t.Fatal("Get returned nil error, want a *webdriver.Error")
	}
	we, ok := err.(*webdriver.Error)
	if !ok {
		t.Fatalf("Get error type = %T, want *webdriver.Error", err)
	}
	if we.Err != "unknown error" || we.Message != "something failed" {
		t.Errorf("error = %+v, want unknown error/something failed", we)
	}
}

func TestDeleteSession(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	if err := DeleteSession(m.srv.URL, wd.SessionID()); err != nil {
		t.Fatalf("DeleteSession returned error: %v", err)
	}
	last := m.last()
	if last.method != "DELETE" || last.path != "/session/sess-1" {
		t.Errorf("DeleteSession request = %s %s", last.method, last.path)
	}
}

func TestStatus(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/status"] = map[string]interface{}{"ready": true, "message": "ok"}

	wd := newTestDriver(t, m)
	st, err := wd.Status()
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if !st.Ready || st.Message != "ok" {
		t.Errorf("Status = %+v, want ready/ok", st)
	}
}

func TestSetDebug(t *testing.T) {
	// SetDebug toggles a package-level flag; ensure it is reversible.
	SetDebug(true)
	if !debugFlag {
		t.Errorf("debugFlag = false after SetDebug(true)")
	}
	SetDebug(false)
	if debugFlag {
		t.Errorf("debugFlag = true after SetDebug(false)")
	}
}

func TestParseVersion(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{"120.0.1", "120.0.1"},
		{"61.0.3116.0", "61.0.3116"}, // 4-part versions are truncated to semver
		{"55.0a1", "55.0.0"},
	} {
		v, err := parseVersion(tc.in)
		if err != nil {
			t.Errorf("parseVersion(%q) returned error: %v", tc.in, err)
			continue
		}
		if v.String() != tc.want {
			t.Errorf("parseVersion(%q) = %q, want %q", tc.in, v.String(), tc.want)
		}
	}
}

func TestWaitTimesOut(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	never := func(webdriver.WebDriver) (bool, error) { return false, nil }
	err := wd.WaitWithTimeoutAndInterval(never, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Fatal("WaitWithTimeoutAndInterval returned nil, want a timeout error")
	}
}

// base64Of encodes some bytes the way Selenium returns screenshots/PDFs.
func base64Of(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func TestWindowHandles(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/window/handles"] = []interface{}{"w1", "w2"}
	m.values["/session/sess-1/window"] = "w1"

	wd := newTestDriver(t, m)

	handles, err := wd.WindowHandles()
	if err != nil {
		t.Fatalf("WindowHandles returned error: %v", err)
	}
	if len(handles) != 2 || handles[0] != "w1" || handles[1] != "w2" {
		t.Errorf("WindowHandles = %v, want [w1 w2]", handles)
	}
	if last := m.last(); last.method != "GET" || last.path != "/session/sess-1/window/handles" {
		t.Errorf("WindowHandles request = %s %s", last.method, last.path)
	}

	h, err := wd.CurrentWindowHandle()
	if err != nil {
		t.Fatalf("CurrentWindowHandle returned error: %v", err)
	}
	if h != "w1" {
		t.Errorf("CurrentWindowHandle = %q, want w1", h)
	}
}

func TestNewWindow(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/window/new"] = map[string]interface{}{"handle": "w9", "type": "tab"}

	wd := newTestDriver(t, m)
	win, err := wd.NewWindow(true)
	if err != nil {
		t.Fatalf("NewWindow returned error: %v", err)
	}
	if win.Handle != "w9" || win.Type != "tab" {
		t.Errorf("NewWindow = %+v, want handle=w9 type=tab", win)
	}
	last := m.last()
	if last.method != "POST" || last.path != "/session/sess-1/window/new" || last.body["type"] != "tab" {
		t.Errorf("NewWindow request = %s %s %v", last.method, last.path, last.body)
	}
}

func TestCloseWindow(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	if err := wd.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	last := m.last()
	if last.method != "DELETE" || last.path != "/session/sess-1/window" {
		t.Errorf("Close request = %s %s", last.method, last.path)
	}
}

func TestSwitchWindow(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	if err := wd.SwitchWindow("w2"); err != nil {
		t.Fatalf("SwitchWindow returned error: %v", err)
	}
	last := m.last()
	// W3C mode uses the "handle" key.
	if last.method != "POST" || last.path != "/session/sess-1/window" || last.body["handle"] != "w2" {
		t.Errorf("SwitchWindow request = %s %s %v", last.method, last.path, last.body)
	}
}

func TestMaximizeWindow(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/window/maximize"] = map[string]interface{}{}

	wd := newTestDriver(t, m)
	if err := wd.MaximizeWindow(""); err != nil {
		t.Fatalf("MaximizeWindow returned error: %v", err)
	}
	last := m.last()
	if last.method != "POST" || last.path != "/session/sess-1/window/maximize" {
		t.Errorf("MaximizeWindow request = %s %s", last.method, last.path)
	}
}

func TestResizeWindow(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/window/rect"] = map[string]interface{}{}

	wd := newTestDriver(t, m)
	if err := wd.ResizeWindow("", 640, 480); err != nil {
		t.Fatalf("ResizeWindow returned error: %v", err)
	}
	last := m.last()
	if last.method != "POST" || last.path != "/session/sess-1/window/rect" {
		t.Errorf("ResizeWindow request = %s %s", last.method, last.path)
	}
	if last.body["width"].(float64) != 640 || last.body["height"].(float64) != 480 {
		t.Errorf("ResizeWindow body = %v, want 640x480", last.body)
	}
}

func TestNavigationForwardBackRefresh(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)

	if err := wd.Forward(); err != nil {
		t.Fatalf("Forward returned error: %v", err)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/forward" {
		t.Errorf("Forward request = %s %s", last.method, last.path)
	}

	if err := wd.Back(); err != nil {
		t.Fatalf("Back returned error: %v", err)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/back" {
		t.Errorf("Back request = %s %s", last.method, last.path)
	}

	if err := wd.Refresh(); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/refresh" {
		t.Errorf("Refresh request = %s %s", last.method, last.path)
	}
}

func TestSwitchFrameByIndex(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	if err := wd.SwitchFrame(2); err != nil {
		t.Fatalf("SwitchFrame returned error: %v", err)
	}
	last := m.last()
	if last.method != "POST" || last.path != "/session/sess-1/frame" || last.body["id"].(float64) != 2 {
		t.Errorf("SwitchFrame request = %s %s %v", last.method, last.path, last.body)
	}
}

func TestSwitchFrameToDefault(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	// An empty string selects the default (top-level) content: id should be nil.
	if err := wd.SwitchFrame(""); err != nil {
		t.Fatalf("SwitchFrame returned error: %v", err)
	}
	last := m.last()
	if last.path != "/session/sess-1/frame" {
		t.Errorf("SwitchFrame path = %s", last.path)
	}
	if v, ok := last.body["id"]; !ok || v != nil {
		t.Errorf("SwitchFrame body id = %v (present=%v), want nil", v, ok)
	}
}

func TestPrint(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	want := []byte("%PDF-1.7 fake")
	m.values["/session/sess-1/print"] = base64Of(want)

	wd := newTestDriver(t, m)
	got, err := wd.Print(webdriver.PrintOptions{})
	if err != nil {
		t.Fatalf("Print returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("Print = %q, want %q", got, want)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/print" {
		t.Errorf("Print request = %s %s", last.method, last.path)
	}
}

func TestTimeouts(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)

	if err := wd.SetImplicitWaitTimeout(3 * time.Second); err != nil {
		t.Fatalf("SetImplicitWaitTimeout returned error: %v", err)
	}
	if last := m.last(); last.path != "/session/sess-1/timeouts" || last.body["implicit"].(float64) != 3000 {
		t.Errorf("SetImplicitWaitTimeout request = %s %v", last.path, last.body)
	}

	if err := wd.SetPageLoadTimeout(4 * time.Second); err != nil {
		t.Fatalf("SetPageLoadTimeout returned error: %v", err)
	}
	if last := m.last(); last.path != "/session/sess-1/timeouts" || last.body["pageLoad"].(float64) != 4000 {
		t.Errorf("SetPageLoadTimeout request = %s %v", last.path, last.body)
	}

	if err := wd.SetAsyncScriptTimeout(5 * time.Second); err != nil {
		t.Fatalf("SetAsyncScriptTimeout returned error: %v", err)
	}
	if last := m.last(); last.path != "/session/sess-1/timeouts" || last.body["script"].(float64) != 5000 {
		t.Errorf("SetAsyncScriptTimeout request = %s %v", last.path, last.body)
	}
}

func TestCookieLifecycle(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)

	if err := wd.AddCookie(&webdriver.Cookie{Name: "sid", Value: "abc"}); err != nil {
		t.Fatalf("AddCookie returned error: %v", err)
	}
	last := m.last()
	if last.method != "POST" || last.path != "/session/sess-1/cookie" {
		t.Errorf("AddCookie request = %s %s", last.method, last.path)
	}
	if c, ok := last.body["cookie"].(map[string]interface{}); !ok || c["name"] != "sid" {
		t.Errorf("AddCookie body = %v", last.body)
	}

	if err := wd.DeleteCookie("sid"); err != nil {
		t.Fatalf("DeleteCookie returned error: %v", err)
	}
	if last := m.last(); last.method != "DELETE" || last.path != "/session/sess-1/cookie/sid" {
		t.Errorf("DeleteCookie request = %s %s", last.method, last.path)
	}

	if err := wd.DeleteAllCookies(); err != nil {
		t.Fatalf("DeleteAllCookies returned error: %v", err)
	}
	if last := m.last(); last.method != "DELETE" || last.path != "/session/sess-1/cookie" {
		t.Errorf("DeleteAllCookies request = %s %s", last.method, last.path)
	}
}

func TestGetCookieChrome(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	// The browser is chrome, so GetCookie filters the full cookie list.
	m.values["/session/sess-1/cookie"] = []interface{}{
		map[string]interface{}{"name": "a", "value": "1"},
		map[string]interface{}{"name": "sid", "value": "abc", "domain": "example.com"},
	}

	wd := newTestDriver(t, m)
	c, err := wd.GetCookie("sid")
	if err != nil {
		t.Fatalf("GetCookie returned error: %v", err)
	}
	if c.Name != "sid" || c.Value != "abc" {
		t.Errorf("GetCookie = %+v, want sid/abc", c)
	}
}

func TestAlerts(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/alert/text"] = "are you sure?"

	wd := newTestDriver(t, m)

	txt, err := wd.AlertText()
	if err != nil {
		t.Fatalf("AlertText returned error: %v", err)
	}
	if txt != "are you sure?" {
		t.Errorf("AlertText = %q", txt)
	}
	if last := m.last(); last.method != "GET" || last.path != "/session/sess-1/alert/text" {
		t.Errorf("AlertText request = %s %s", last.method, last.path)
	}

	if err := wd.SetAlertText("hello"); err != nil {
		t.Fatalf("SetAlertText returned error: %v", err)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/alert/text" || last.body["text"] != "hello" {
		t.Errorf("SetAlertText request = %s %s %v", last.method, last.path, last.body)
	}

	if err := wd.AcceptAlert(); err != nil {
		t.Fatalf("AcceptAlert returned error: %v", err)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/alert/accept" {
		t.Errorf("AcceptAlert request = %s %s", last.method, last.path)
	}

	if err := wd.DismissAlert(); err != nil {
		t.Fatalf("DismissAlert returned error: %v", err)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/alert/dismiss" {
		t.Errorf("DismissAlert request = %s %s", last.method, last.path)
	}
}

func TestActiveElement(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element/active"] = element("e5")
	m.values["/session/sess-1/element/e5/text"] = "focused"

	wd := newTestDriver(t, m)
	el, err := wd.ActiveElement()
	if err != nil {
		t.Fatalf("ActiveElement returned error: %v", err)
	}
	if last := m.last(); last.method != "GET" || last.path != "/session/sess-1/element/active" {
		t.Errorf("ActiveElement request = %s %s", last.method, last.path)
	}
	if txt, err := el.Text(); err != nil || txt != "focused" {
		t.Errorf("Text on active element = %q, %v", txt, err)
	}
}

func TestScreenshot(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	want := []byte{0x89, 0x50, 0x4e, 0x47} // PNG header
	m.values["/session/sess-1/screenshot"] = base64Of(want)

	wd := newTestDriver(t, m)
	got, err := wd.Screenshot()
	if err != nil {
		t.Fatalf("Screenshot returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("Screenshot = %v, want %v", got, want)
	}
}

func TestCapabilitiesAndStatusID(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	caps, err := wd.Capabilities()
	if err != nil {
		t.Fatalf("Capabilities returned error: %v", err)
	}
	// Capabilities are captured from the new-session response (the W3C protocol
	// has no get-capabilities command), so both browserName and the negotiated
	// browserVersion are present.
	if caps["browserName"] != "chrome" || caps["browserVersion"] != "120.0.1" {
		t.Errorf("Capabilities = %v, want browserName=chrome browserVersion=120.0.1", caps)
	}
	// Capabilities() must not issue the legacy GET /session/:id request on W3C.
	for _, r := range m.reqs {
		if r.method == "GET" && r.path == "/session/sess-1" {
			t.Errorf("Capabilities() made a legacy GET /session/sess-1 request")
		}
	}
	if err := wd.SwitchSession("other"); err != nil {
		t.Fatalf("SwitchSession returned error: %v", err)
	}
	if wd.SessionID() != "other" {
		t.Errorf("SessionID after SwitchSession = %q, want other", wd.SessionID())
	}
}

func TestQuit(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	if err := wd.Quit(); err != nil {
		t.Fatalf("Quit returned error: %v", err)
	}
	if last := m.last(); last.method != "DELETE" || last.path != "/session/sess-1" {
		t.Errorf("Quit request = %s %s", last.method, last.path)
	}
	if wd.SessionID() != "" {
		t.Errorf("SessionID after Quit = %q, want empty", wd.SessionID())
	}
}

func TestElementInteractions(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")

	wd := newTestDriver(t, m)
	el, err := wd.FindElement(webdriver.ByCSSSelector, "#x")
	if err != nil {
		t.Fatalf("FindElement returned error: %v", err)
	}

	if err := el.Click(); err != nil {
		t.Fatalf("Click returned error: %v", err)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/element/e1/click" {
		t.Errorf("Click request = %s %s", last.method, last.path)
	}

	if err := el.Clear(); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}
	if last := m.last(); last.path != "/session/sess-1/element/e1/clear" {
		t.Errorf("Clear request = %s", last.path)
	}

	if err := el.SendKeys("hi"); err != nil {
		t.Fatalf("SendKeys returned error: %v", err)
	}
	last := m.last()
	if last.method != "POST" || last.path != "/session/sess-1/element/e1/value" || last.body["text"] != "hi" {
		t.Errorf("SendKeys request = %s %s %v", last.method, last.path, last.body)
	}
	// Submit and MoveTo are W3C-only reimplementations; see
	// TestElementSubmitUsesScript and TestElementMoveToUsesW3CActions.
}

func TestElementQueries(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")
	m.values["/session/sess-1/element/e1/selected"] = true
	m.values["/session/sess-1/element/e1/enabled"] = true
	m.values["/session/sess-1/element/e1/displayed"] = false
	m.values["/session/sess-1/element/e1/name"] = "input"
	m.values["/session/sess-1/element/e1/property/value"] = "hello"
	m.values["/session/sess-1/element/e1/css/color"] = "rgb(0, 0, 0)"

	wd := newTestDriver(t, m)
	el, err := wd.FindElement(webdriver.ByCSSSelector, "#x")
	if err != nil {
		t.Fatalf("FindElement returned error: %v", err)
	}

	if sel, err := el.IsSelected(); err != nil || !sel {
		t.Errorf("IsSelected = %v, %v", sel, err)
	}
	if en, err := el.IsEnabled(); err != nil || !en {
		t.Errorf("IsEnabled = %v, %v", en, err)
	}
	if disp, err := el.IsDisplayed(); err != nil || disp {
		t.Errorf("IsDisplayed = %v, %v", disp, err)
	}
	if tag, err := el.TagName(); err != nil || tag != "input" {
		t.Errorf("TagName = %q, %v", tag, err)
	}
	if prop, err := el.GetProperty("value"); err != nil || prop != "hello" {
		t.Errorf("GetProperty = %q, %v", prop, err)
	}
	if css, err := el.CSSProperty("color"); err != nil || css != "rgb(0, 0, 0)" {
		t.Errorf("CSSProperty = %q, %v", css, err)
	}
}

func TestElementRectSizeLocation(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")
	m.values["/session/sess-1/element/e1/rect"] = map[string]interface{}{"x": 10, "y": 20, "width": 100, "height": 50}

	wd := newTestDriver(t, m)
	el, err := wd.FindElement(webdriver.ByCSSSelector, "#x")
	if err != nil {
		t.Fatalf("FindElement returned error: %v", err)
	}

	sz, err := el.Size()
	if err != nil {
		t.Fatalf("Size returned error: %v", err)
	}
	if sz.Width != 100 || sz.Height != 50 {
		t.Errorf("Size = %+v, want 100x50", *sz)
	}

	loc, err := el.Location()
	if err != nil {
		t.Fatalf("Location returned error: %v", err)
	}
	if loc.X != 10 || loc.Y != 20 {
		t.Errorf("Location = %+v, want (10,20)", *loc)
	}
}

func TestElementScreenshot(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")
	want := []byte{0x89, 0x50, 0x4e, 0x47}
	m.values["/session/sess-1/element/e1/screenshot"] = base64Of(want)

	wd := newTestDriver(t, m)
	el, err := wd.FindElement(webdriver.ByCSSSelector, "#x")
	if err != nil {
		t.Fatalf("FindElement returned error: %v", err)
	}
	got, err := el.Screenshot(true)
	if err != nil {
		t.Fatalf("Screenshot returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("element Screenshot = %v, want %v", got, want)
	}
}

func TestElementFindElement(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")
	m.values["/session/sess-1/element/e1/element"] = element("e2")
	m.values["/session/sess-1/element/e1/elements"] = []interface{}{element("e2"), element("e3")}

	wd := newTestDriver(t, m)
	parent, err := wd.FindElement(webdriver.ByCSSSelector, "#p")
	if err != nil {
		t.Fatalf("FindElement returned error: %v", err)
	}

	child, err := parent.FindElement(webdriver.ByCSSSelector, ".c")
	if err != nil {
		t.Fatalf("nested FindElement returned error: %v", err)
	}
	if last := m.last(); last.path != "/session/sess-1/element/e1/element" {
		t.Errorf("nested FindElement path = %s", last.path)
	}
	_ = child

	children, err := parent.FindElements(webdriver.ByCSSSelector, ".c")
	if err != nil {
		t.Fatalf("nested FindElements returned error: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("nested FindElements returned %d, want 2", len(children))
	}
}

func TestShadowRoot(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")
	m.values["/session/sess-1/element/e1/shadow"] = map[string]interface{}{
		shadowRootIdentifier: "s1",
	}
	m.values["/session/sess-1/shadow/s1/element"] = element("e9")
	m.values["/session/sess-1/shadow/s1/elements"] = []interface{}{element("e9")}

	wd := newTestDriver(t, m)
	el, err := wd.FindElement(webdriver.ByCSSSelector, "#host")
	if err != nil {
		t.Fatalf("FindElement returned error: %v", err)
	}

	sr, err := el.GetShadowRoot()
	if err != nil {
		t.Fatalf("GetShadowRoot returned error: %v", err)
	}
	if last := m.last(); last.method != "GET" || last.path != "/session/sess-1/element/e1/shadow" {
		t.Errorf("GetShadowRoot request = %s %s", last.method, last.path)
	}

	if _, err := sr.FindElement(webdriver.ByCSSSelector, ".inner"); err != nil {
		t.Fatalf("ShadowRoot.FindElement returned error: %v", err)
	}
	if last := m.last(); last.method != "POST" || last.path != "/session/sess-1/shadow/s1/element" {
		t.Errorf("ShadowRoot.FindElement request = %s %s", last.method, last.path)
	}

	els, err := sr.FindElements(webdriver.ByCSSSelector, ".inner")
	if err != nil {
		t.Fatalf("ShadowRoot.FindElements returned error: %v", err)
	}
	if len(els) != 1 {
		t.Errorf("ShadowRoot.FindElements returned %d, want 1", len(els))
	}
}

func TestFindElementRelative(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	// findRelative first fetches candidates via /elements, then runs the JS atom
	// via /execute/sync, which returns the filtered/sorted element list.
	m.values["/session/sess-1/elements"] = []interface{}{element("c1"), element("c2")}
	m.values["/session/sess-1/element"] = element("anchor")
	m.values["/session/sess-1/execute/sync"] = []interface{}{element("c2")}

	wd := newTestDriver(t, m)
	anchor, err := wd.FindElement(webdriver.ByCSSSelector, "#anchor")
	if err != nil {
		t.Fatalf("FindElement returned error: %v", err)
	}

	rel := webdriver.With(webdriver.ByCSSSelector, "div").Above(anchor)
	el, err := wd.FindElementRelative(rel)
	if err != nil {
		t.Fatalf("FindElementRelative returned error: %v", err)
	}
	if el.(*remoteWE).id != "c2" {
		t.Errorf("FindElementRelative returned id %q, want c2", el.(*remoteWE).id)
	}
	if last := m.last(); last.path != "/session/sess-1/execute/sync" {
		t.Errorf("FindElementRelative last request path = %s", last.path)
	}
}

func TestFindElementsRelative(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/elements"] = []interface{}{element("c1"), element("c2")}
	m.values["/session/sess-1/element"] = element("anchor")
	m.values["/session/sess-1/execute/sync"] = []interface{}{element("c1"), element("c2")}

	wd := newTestDriver(t, m)
	anchor, err := wd.FindElement(webdriver.ByCSSSelector, "#anchor")
	if err != nil {
		t.Fatalf("FindElement returned error: %v", err)
	}

	rel := webdriver.With(webdriver.ByCSSSelector, "div").Near(anchor)
	els, err := wd.FindElementsRelative(rel)
	if err != nil {
		t.Fatalf("FindElementsRelative returned error: %v", err)
	}
	if len(els) != 2 {
		t.Errorf("FindElementsRelative returned %d, want 2", len(els))
	}
}

func TestKeyDownUpUsesW3CActions(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)

	// In W3C mode, KeyDown/KeyUp go through the actions endpoint. (The mouse
	// helpers Click/DoubleClick/ButtonDown/ButtonUp are covered by
	// TestMouseMethodsUseW3CActions.)
	if err := wd.KeyDown("a"); err != nil {
		t.Fatalf("KeyDown returned error: %v", err)
	}
	if last := m.last(); last.path != "/session/sess-1/actions" {
		t.Errorf("KeyDown path = %s", last.path)
	}
	if err := wd.KeyUp("a"); err != nil {
		t.Fatalf("KeyUp returned error: %v", err)
	}
	if last := m.last(); last.path != "/session/sess-1/actions" {
		t.Errorf("KeyUp path = %s", last.path)
	}
}

func TestStoreKeyActionsAndRelease(t *testing.T) {
	m := newMockServer(t)
	defer m.close()

	wd := newTestDriver(t, m)
	wd.StoreKeyActions("kbd",
		webdriver.KeyDownAction("a"),
		webdriver.KeyUpAction("a"),
	)
	if err := wd.PerformActions(); err != nil {
		t.Fatalf("PerformActions returned error: %v", err)
	}
	last := m.last()
	if last.path != "/session/sess-1/actions" {
		t.Fatalf("PerformActions path = %s", last.path)
	}
	srcs := last.body["actions"].([]interface{})
	if len(srcs) != 1 || srcs[0].(map[string]interface{})["type"] != "key" {
		t.Errorf("stored key actions = %v", srcs)
	}

	if err := wd.ReleaseActions(); err != nil {
		t.Fatalf("ReleaseActions returned error: %v", err)
	}
	if last := m.last(); last.method != "DELETE" || last.path != "/session/sess-1/actions" {
		t.Errorf("ReleaseActions request = %s %s", last.method, last.path)
	}
}

func TestByNameRewriting(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")

	wd := newTestDriver(t, m)
	if _, err := wd.FindElement(webdriver.ByName, "q"); err != nil {
		t.Fatalf("FindElement(ByName) returned error: %v", err)
	}
	last := m.last()
	if last.body["using"] != webdriver.ByCSSSelector || last.body["value"] != `input[name="q"]` {
		t.Errorf("ByName rewrite: using=%v value=%v", last.body["using"], last.body["value"])
	}
}

func TestExecuteScriptAsync(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/execute/async"] = "done"

	wd := newTestDriver(t, m)
	res, err := wd.ExecuteScriptAsync("cb();", nil)
	if err != nil {
		t.Fatalf("ExecuteScriptAsync returned error: %v", err)
	}
	if res != "done" {
		t.Errorf("ExecuteScriptAsync = %v, want done", res)
	}
	if last := m.last(); last.path != "/session/sess-1/execute/async" {
		t.Errorf("ExecuteScriptAsync path = %s", last.path)
	}
}

func TestErrorPropagatesToTypedMethods(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.errs["/session/sess-1/title"] = map[string]interface{}{
		"error":   "no such window",
		"message": "window closed",
	}

	wd := newTestDriver(t, m)
	_, err := wd.Title()
	if err == nil {
		t.Fatal("Title returned nil, want an error")
	}
	we, ok := err.(*webdriver.Error)
	if !ok {
		t.Fatalf("Title error type = %T, want *webdriver.Error", err)
	}
	if we.Err != "no such window" {
		t.Errorf("error = %+v", we)
	}
}

// pointerActionTypes returns the action-type sequence of the first pointer input
// source in a recorded /actions request body.
func pointerActionTypes(t *testing.T, body map[string]interface{}) (string, []string) {
	srcs, ok := body["actions"].([]interface{})
	if !ok || len(srcs) == 0 {
		t.Fatalf("actions body has no input sources: %v", body)
	}
	src := srcs[0].(map[string]interface{})
	pt, _ := src["parameters"].(map[string]interface{})
	var types []string
	for _, a := range src["actions"].([]interface{}) {
		types = append(types, a.(map[string]interface{})["type"].(string))
	}
	ptype, _ := pt["pointerType"].(string)
	return ptype, types
}

func eqStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestMouseMethodsUseW3CActions verifies the legacy mouse methods now emit W3C
// pointer-action sequences (the /click, /doubleclick, /buttondown, /buttonup
// endpoints were removed in W3C).
func TestMouseMethodsUseW3CActions(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	wd := newTestDriver(t, m)

	cases := []struct {
		name  string
		call  func() error
		types []string
	}{
		{"Click", func() error { return wd.Click(int(webdriver.LeftButton)) }, []string{"pointerDown", "pointerUp"}},
		{"DoubleClick", wd.DoubleClick, []string{"pointerDown", "pointerUp", "pointerDown", "pointerUp"}},
		{"ButtonDown", wd.ButtonDown, []string{"pointerDown"}},
		{"ButtonUp", wd.ButtonUp, []string{"pointerUp"}},
	}
	for _, tc := range cases {
		if err := tc.call(); err != nil {
			t.Fatalf("%s returned error: %v", tc.name, err)
		}
		last := m.last()
		if last.method != "POST" || last.path != "/session/sess-1/actions" {
			t.Errorf("%s request = %s %s, want POST /session/sess-1/actions", tc.name, last.method, last.path)
		}
		ptype, types := pointerActionTypes(t, last.body)
		if ptype != "mouse" {
			t.Errorf("%s pointerType = %q, want mouse", tc.name, ptype)
		}
		if !eqStrings(types, tc.types) {
			t.Errorf("%s action types = %v, want %v", tc.name, types, tc.types)
		}
	}
}

func TestElementMoveToUsesW3CActions(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")
	wd := newTestDriver(t, m)

	el, err := wd.FindElement(webdriver.ByCSSSelector, "#x")
	if err != nil {
		t.Fatalf("FindElement: %v", err)
	}
	if err := el.MoveTo(5, 7); err != nil {
		t.Fatalf("MoveTo returned error: %v", err)
	}
	last := m.last()
	if last.path != "/session/sess-1/actions" {
		t.Fatalf("MoveTo path = %s, want /actions", last.path)
	}
	_, types := pointerActionTypes(t, last.body)
	if !eqStrings(types, []string{"pointerMove"}) {
		t.Errorf("MoveTo action types = %v, want [pointerMove]", types)
	}
	move := last.body["actions"].([]interface{})[0].(map[string]interface{})["actions"].([]interface{})[0].(map[string]interface{})
	if move["x"].(float64) != 5 || move["y"].(float64) != 7 {
		t.Errorf("MoveTo offsets = %v/%v, want 5/7", move["x"], move["y"])
	}
	// The origin must be the element reference, not "viewport".
	origin, ok := move["origin"].(map[string]interface{})
	if !ok || origin[webElementIdentifier] != "e1" {
		t.Errorf("MoveTo origin = %v, want element e1", move["origin"])
	}
}

func TestElementSubmitUsesScript(t *testing.T) {
	m := newMockServer(t)
	defer m.close()
	m.values["/session/sess-1/element"] = element("e1")
	wd := newTestDriver(t, m)

	el, err := wd.FindElement(webdriver.ByCSSSelector, "#x")
	if err != nil {
		t.Fatalf("FindElement: %v", err)
	}
	if err := el.Submit(); err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	// W3C removed element/submit, so it must run a script instead.
	last := m.last()
	if last.path != "/session/sess-1/execute/sync" {
		t.Fatalf("Submit path = %s, want /execute/sync", last.path)
	}
	script, _ := last.body["script"].(string)
	if !strings.Contains(script, "requestSubmit") || !strings.Contains(script, "closest('form')") {
		t.Errorf("Submit script = %q, want it to submit the closest form", script)
	}
	args := last.body["args"].([]interface{})
	if len(args) != 1 || args[0].(map[string]interface{})[webElementIdentifier] != "e1" {
		t.Errorf("Submit args = %v, want [element e1]", args)
	}
}

// TestChromeCapabilitiesReachAlwaysMatch verifies that AddChrome's options flow
// into the W3C alwaysMatch payload under goog:chromeOptions (with custom prefs
// preserved), and that the deprecated unprefixed "chromeOptions" key — which
// strict W3C servers reject — is dropped.
func TestChromeCapabilitiesReachAlwaysMatch(t *testing.T) {
	caps := webdriver.Capabilities{"browserName": "chrome"}
	caps.AddChrome(chrome.Capabilities{
		Args:  []string{"--headless=new"},
		Prefs: map[string]interface{}{"intl.accept_languages": "de-DE"},
		MobileEmulation: &chrome.MobileEmulation{
			DeviceMetrics: &chrome.DeviceMetrics{Width: 360, Height: 640, PixelRatio: 2},
		},
	})

	am, ok := newW3CCapabilities(caps)["alwaysMatch"].(webdriver.Capabilities)
	if !ok {
		t.Fatal("newW3CCapabilities returned no alwaysMatch")
	}
	opts, ok := am["goog:chromeOptions"].(chrome.Capabilities)
	if !ok {
		t.Fatalf("goog:chromeOptions missing/incorrect in alwaysMatch: %#v", am["goog:chromeOptions"])
	}
	if opts.Prefs["intl.accept_languages"] != "de-DE" {
		t.Errorf("custom pref not preserved: %v", opts.Prefs)
	}
	if opts.MobileEmulation == nil || opts.MobileEmulation.DeviceMetrics.Width != 360 {
		t.Errorf("mobileEmulation not preserved: %v", opts.MobileEmulation)
	}
	// The deprecated unprefixed key must not appear in the W3C payload.
	if _, present := am["chromeOptions"]; present {
		t.Error("deprecated chromeOptions leaked into W3C alwaysMatch")
	}
}
