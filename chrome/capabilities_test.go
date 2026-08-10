package chrome

import (
	"encoding/json"
	"testing"
)

func TestEmptyCapabilities(t *testing.T) {
	data, err := json.Marshal(Capabilities{})
	if err != nil {
		t.Fatalf("json.Marshal(Capabilities{}) return error: %v", err)
	}
	// An unset W3C field must be omitted rather than serialized as
	// "w3c": false, which would request the removed legacy protocol and cause
	// Selenium 4 to reject the session handshake.
	got, want := string(data), `{}`
	if got != want {
		t.Fatalf("json.Marshal(Capabilities{}) = %q, want %q", got, want)
	}
}

func TestW3CCapability(t *testing.T) {
	data, err := json.Marshal(Capabilities{W3C: true})
	if err != nil {
		t.Fatalf("json.Marshal(Capabilities{W3C: true}) return error: %v", err)
	}
	got, want := string(data), `{"w3c":true}`
	if got != want {
		t.Fatalf("json.Marshal(Capabilities{W3C: true}) = %q, want %q", got, want)
	}
}
