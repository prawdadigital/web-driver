package sauce

import (
	"encoding/json"
	"testing"
)

func TestAddr(t *testing.T) {
	got := Addr("bob", "secret")
	want := "http://bob:secret@ondemand.saucelabs.com/wd/hub"
	if got != want {
		t.Errorf("Addr(bob, secret) = %q, want %q", got, want)
	}
}

func TestCapabilitiesToMap(t *testing.T) {
	rec := true
	c := &Capabilities{
		Browser:             "chrome",
		Version:             "120",
		Platform:            "Windows 10",
		SeleniumVersion:     "4.0.0",
		ChromeDriverVersion: "120.0",
		TestName:            "my-test",
		BuildNumber:         "build-42",
		Tags:                []string{"smoke", "regression"},
		MaximumDuration:     1800,
		Visibility:          Public,
		RecordVideo:         &rec,
	}
	m, err := c.ToMap()
	if err != nil {
		t.Fatalf("ToMap returned error: %v", err)
	}

	// Verify a representative sample of the JSON-tagged (translated) keys.
	strChecks := map[string]string{
		"browser":             "chrome",
		"version":             "120",
		"platform":            "Windows 10",
		"seleniumVersion":     "4.0.0",
		"chromedriverVersion": "120.0",
		"name":                "my-test",
		"build":               "build-42",
		"public":              string(Public),
	}
	for k, want := range strChecks {
		got, ok := m[k]
		if !ok {
			t.Errorf("ToMap missing key %q", k)
			continue
		}
		if gs, _ := got.(string); gs != want {
			t.Errorf("ToMap[%q] = %v, want %q", k, got, want)
		}
	}

	// maxDuration is numeric; JSON unmarshals numbers to float64.
	if v, ok := m["maxDuration"].(float64); !ok || v != 1800 {
		t.Errorf("ToMap[maxDuration] = %v, want 1800", m["maxDuration"])
	}

	// recordVideo pointer should serialize under the "recordVideo" key.
	if v, ok := m["recordVideo"].(bool); !ok || v != true {
		t.Errorf("ToMap[recordVideo] = %v, want true", m["recordVideo"])
	}

	// tags should be a JSON array.
	tags, ok := m["tags"].([]interface{})
	if !ok || len(tags) != 2 {
		t.Fatalf("ToMap[tags] = %v, want a 2-element array", m["tags"])
	}
}

func TestCapabilitiesToMapEmpty(t *testing.T) {
	m, err := (&Capabilities{}).ToMap()
	if err != nil {
		t.Fatalf("ToMap returned error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("empty Capabilities.ToMap() = %v, want empty map", m)
	}
}

func TestCapabilitiesToMapCustomData(t *testing.T) {
	c := &Capabilities{CustomData: json.RawMessage(`{"foo":"bar"}`)}
	m, err := c.ToMap()
	if err != nil {
		t.Fatalf("ToMap returned error: %v", err)
	}
	cd, ok := m["customData"].(map[string]interface{})
	if !ok {
		t.Fatalf("ToMap[customData] = %v (%T), want object", m["customData"], m["customData"])
	}
	if cd["foo"] != "bar" {
		t.Errorf("customData.foo = %v, want bar", cd["foo"])
	}
}

func TestConnectAddr(t *testing.T) {
	c := &Connect{
		UserName:     "bob",
		AccessKey:    "secret",
		SeleniumPort: 4445,
	}
	got := c.Addr()
	want := "http://bob:secret@localhost:4445/wd/hub"
	if got != want {
		t.Errorf("Connect.Addr() = %q, want %q", got, want)
	}
}
