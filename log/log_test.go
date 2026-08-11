package log

import (
	"encoding/json"
	"testing"
)

func TestTypeValues(t *testing.T) {
	want := map[Type]string{
		Server:      "server",
		Browser:     "browser",
		Client:      "client",
		Driver:      "driver",
		Performance: "performance",
		Profiler:    "profiler",
	}
	for typ, s := range want {
		if string(typ) != s {
			t.Errorf("Type %q = %q, want %q", s, string(typ), s)
		}
	}
}

func TestLevelValues(t *testing.T) {
	want := map[Level]string{
		Off:     "OFF",
		Severe:  "SEVERE",
		Warning: "WARNING",
		Info:    "INFO",
		Debug:   "DEBUG",
		All:     "ALL",
	}
	for lvl, s := range want {
		if string(lvl) != s {
			t.Errorf("Level %q = %q, want %q", s, string(lvl), s)
		}
	}
}

func TestCapabilitiesKey(t *testing.T) {
	if CapabilitiesKey != "goog:loggingPrefs" {
		t.Errorf("CapabilitiesKey = %q, want %q", CapabilitiesKey, "goog:loggingPrefs")
	}
}

func TestCapabilitiesMarshal(t *testing.T) {
	caps := Capabilities{
		Browser: Info,
		Driver:  All,
	}
	b, err := json.Marshal(caps)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	// Round-trip into a plain map so the assertion is order-independent.
	var got map[string]string
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	want := map[string]string{
		"browser": "INFO",
		"driver":  "ALL",
	}
	if len(got) != len(want) {
		t.Fatalf("marshaled map = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("marshaled[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestMessageFields(t *testing.T) {
	// Ensure Message is JSON-friendly and its fields carry the expected values.
	m := Message{
		Level:   Severe,
		Message: "boom",
	}
	if m.Level != Severe {
		t.Errorf("Level = %q, want %q", m.Level, Severe)
	}
	if m.Message != "boom" {
		t.Errorf("Message = %q, want %q", m.Message, "boom")
	}
	if !m.Timestamp.IsZero() {
		t.Errorf("Timestamp = %v, want zero", m.Timestamp)
	}
}
