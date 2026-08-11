package firefox

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestEmptyCapabilitiesMarshalsToEmptyObject(t *testing.T) {
	data, err := json.Marshal(Capabilities{})
	if err != nil {
		t.Fatalf("json.Marshal(Capabilities{}) returned error: %v", err)
	}
	if got, want := string(data), "{}"; got != want {
		t.Errorf("json.Marshal(Capabilities{}) = %q, want %q", got, want)
	}
}

func TestPopulatedCapabilitiesMarshal(t *testing.T) {
	caps := Capabilities{
		Binary:  "/usr/bin/firefox",
		Args:    []string{"--devtools", "--headless"},
		Profile: "cHJvZmlsZQ==",
		Log:     &Log{Level: Debug},
		Prefs: map[string]interface{}{
			"browser.startup.homepage":     "about:blank",
			"dom.disable_beforeunload":     true,
			"network.http.max-connections": 10,
		},
	}

	data, err := json.Marshal(caps)
	if err != nil {
		t.Fatalf("json.Marshal(caps) returned error: %v", err)
	}

	// Round-trip through a generic map to assert individual keys regardless of
	// field ordering.
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if got["binary"] != "/usr/bin/firefox" {
		t.Errorf("binary = %v, want /usr/bin/firefox", got["binary"])
	}
	args, ok := got["args"].([]interface{})
	if !ok || len(args) != 2 || args[0] != "--devtools" || args[1] != "--headless" {
		t.Errorf("args = %v, want [--devtools --headless]", got["args"])
	}
	if got["profile"] != "cHJvZmlsZQ==" {
		t.Errorf("profile = %v, want cHJvZmlsZQ==", got["profile"])
	}
	logObj, ok := got["log"].(map[string]interface{})
	if !ok || logObj["level"] != "debug" {
		t.Errorf("log = %v, want {level: debug}", got["log"])
	}
	prefs, ok := got["prefs"].(map[string]interface{})
	if !ok {
		t.Fatalf("prefs = %v, want a map", got["prefs"])
	}
	if prefs["browser.startup.homepage"] != "about:blank" {
		t.Errorf("prefs homepage = %v, want about:blank", prefs["browser.startup.homepage"])
	}
	if prefs["dom.disable_beforeunload"] != true {
		t.Errorf("prefs beforeunload = %v, want true", prefs["dom.disable_beforeunload"])
	}
	if prefs["network.http.max-connections"] != float64(10) {
		t.Errorf("prefs max-connections = %v, want 10", prefs["network.http.max-connections"])
	}
}

func TestLogMarshal(t *testing.T) {
	data, err := json.Marshal(Log{Level: Trace})
	if err != nil {
		t.Fatalf("json.Marshal(Log{}) returned error: %v", err)
	}
	if got, want := string(data), `{"level":"trace"}`; got != want {
		t.Errorf("json.Marshal(Log{Level: Trace}) = %q, want %q", got, want)
	}
}

func TestLogLevelValues(t *testing.T) {
	levels := map[LogLevel]string{
		Trace:  "trace",
		Debug:  "debug",
		Config: "config",
		Info:   "info",
		Warn:   "warn",
		Error:  "error",
		Fatal:  "fatal",
	}
	for level, want := range levels {
		if string(level) != want {
			t.Errorf("LogLevel = %q, want %q", string(level), want)
		}
	}
}

func TestSetProfile(t *testing.T) {
	dir, err := os.MkdirTemp("", "firefox-profile")
	if err != nil {
		t.Fatalf("os.MkdirTemp returned error: %v", err)
	}
	defer os.RemoveAll(dir)

	userJS := filepath.Join(dir, "user.js")
	if err := os.WriteFile(userJS, []byte(`user_pref("foo", true);`), 0644); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	var c Capabilities
	if err := c.SetProfile(dir); err != nil {
		t.Fatalf("SetProfile returned error: %v", err)
	}

	if c.Profile == "" {
		t.Fatal("SetProfile left Profile empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(c.Profile)
	if err != nil {
		t.Fatalf("Profile is not valid base64: %v", err)
	}
	if len(decoded) == 0 {
		t.Fatal("decoded Profile is empty")
	}
	// Zip files begin with the local file header signature "PK\x03\x04".
	if len(decoded) < 2 || decoded[0] != 'P' || decoded[1] != 'K' {
		t.Errorf("decoded Profile does not start with a zip signature: % x", decoded[:min(4, len(decoded))])
	}
}

func TestSetProfileNonexistentPath(t *testing.T) {
	var c Capabilities
	if err := c.SetProfile(filepath.Join(os.TempDir(), "does-not-exist-firefox-profile")); err == nil {
		t.Error("SetProfile with a nonexistent path returned nil error, want error")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ExampleCapabilities() {
	caps := Capabilities{
		Binary: "/usr/bin/firefox",
		Args:   []string{"--headless"},
		Log:    &Log{Level: Trace},
	}
	data, _ := json.Marshal(caps)
	fmt.Println(string(data))
	// Output: {"binary":"/usr/bin/firefox","args":["--headless"],"log":{"level":"trace"}}
}
