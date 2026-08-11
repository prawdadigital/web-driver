package webdriver

import (
	"encoding/json"
	"testing"

	"github.com/prawdadigital/web-driver/chrome"
	"github.com/prawdadigital/web-driver/firefox"
	"github.com/prawdadigital/web-driver/log"
)

func TestCapabilitiesAddChrome(t *testing.T) {
	caps := Capabilities{"browserName": "chrome"}
	caps.AddChrome(chrome.Capabilities{Args: []string{"--headless=new"}})
	if _, ok := caps[chrome.CapabilitiesKey]; !ok {
		t.Errorf("AddChrome did not set %q", chrome.CapabilitiesKey)
	}
	if _, ok := caps[chrome.DeprecatedCapabilitiesKey]; !ok {
		t.Errorf("AddChrome did not set the deprecated key %q", chrome.DeprecatedCapabilitiesKey)
	}
}

func TestCapabilitiesAddFirefox(t *testing.T) {
	caps := Capabilities{"browserName": "firefox"}
	caps.AddFirefox(firefox.Capabilities{Binary: "/usr/bin/firefox"})
	if _, ok := caps[firefox.CapabilitiesKey]; !ok {
		t.Errorf("AddFirefox did not set %q", firefox.CapabilitiesKey)
	}
}

func TestCapabilitiesAddProxy(t *testing.T) {
	caps := Capabilities{}
	caps.AddProxy(Proxy{Type: Manual, HTTP: "127.0.0.1:8080"})
	p, ok := caps["proxy"].(Proxy)
	if !ok || p.HTTP != "127.0.0.1:8080" {
		t.Errorf("AddProxy = %v", caps["proxy"])
	}
}

func TestCapabilitiesSetLogLevel(t *testing.T) {
	caps := Capabilities{}
	caps.SetLogLevel(log.Browser, log.Info)
	caps.SetLogLevel(log.Server, log.Warning)
	l, ok := caps[log.CapabilitiesKey].(log.Capabilities)
	if !ok {
		t.Fatalf("SetLogLevel did not store a log.Capabilities, got %T", caps[log.CapabilitiesKey])
	}
	if l[log.Browser] != log.Info || l[log.Server] != log.Warning {
		t.Errorf("log capabilities = %v", l)
	}
}

func TestErrorString(t *testing.T) {
	e := &Error{Err: "no such element", Message: "not found"}
	if got, want := e.Error(), "no such element: not found"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestPrintOptionsOmitEmpty(t *testing.T) {
	// A zero-value PrintOptions must marshal to an empty object so the remote
	// end applies its own defaults.
	data, err := json.Marshal(PrintOptions{})
	if err != nil {
		t.Fatalf("json.Marshal(PrintOptions{}) returned error: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("json.Marshal(PrintOptions{}) = %s, want {}", data)
	}

	data, _ = json.Marshal(PrintOptions{Orientation: LandscapeOrientation, Scale: 0.5})
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	if m["orientation"] != "landscape" || m["scale"].(float64) != 0.5 {
		t.Errorf("PrintOptions marshal = %s", data)
	}
}

func TestRectAndWindowMarshal(t *testing.T) {
	data, _ := json.Marshal(Rect{X: 1, Y: 2, Width: 3, Height: 4})
	var r map[string]interface{}
	json.Unmarshal(data, &r)
	for k, want := range map[string]float64{"x": 1, "y": 2, "width": 3, "height": 4} {
		if r[k].(float64) != want {
			t.Errorf("Rect[%q] = %v, want %v", k, r[k], want)
		}
	}

	var w Window
	if err := json.Unmarshal([]byte(`{"handle":"h1","type":"tab"}`), &w); err != nil {
		t.Fatalf("unmarshal Window: %v", err)
	}
	if w.Handle != "h1" || w.Type != "tab" {
		t.Errorf("Window = %+v", w)
	}
}

func TestCookieMarshal(t *testing.T) {
	data, _ := json.Marshal(Cookie{Name: "sid", Value: "abc", Path: "/", Secure: true})
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	if m["name"] != "sid" || m["value"] != "abc" || m["secure"] != true {
		t.Errorf("Cookie marshal = %s", data)
	}
	// SameSite is omitted when empty.
	if _, ok := m["sameSite"]; ok {
		t.Errorf("empty SameSite should be omitted, got %v", m["sameSite"])
	}
}
