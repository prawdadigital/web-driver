package chrome

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// unmarshalToMap marshals v to JSON and unmarshals it into a generic map so
// individual emitted keys can be asserted without depending on field order.
func unmarshalToMap(t *testing.T, v interface{}) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal(%#v) returned error: %v", v, err)
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal(%q) returned error: %v", data, err)
	}
	return m
}

func TestCapabilitiesMarshal(t *testing.T) {
	enabled := true
	caps := Capabilities{
		Path:            "/usr/bin/google-chrome",
		Args:            []string{"--headless", "--no-sandbox"},
		ExcludeSwitches: []string{"enable-automation"},
		Prefs: map[string]interface{}{
			"download.default_directory": "/tmp",
		},
		MobileEmulation: &MobileEmulation{
			DeviceName: "Google Nexus 5",
		},
		PerfLoggingPrefs: &PerfLoggingPreferences{
			EnableNetwork: &enabled,
		},
	}

	m := unmarshalToMap(t, caps)

	if got, want := m["binary"], "/usr/bin/google-chrome"; got != want {
		t.Errorf("binary = %v, want %v", got, want)
	}

	args, ok := m["args"].([]interface{})
	if !ok || len(args) != 2 {
		t.Errorf("args = %v, want a two-element slice", m["args"])
	}

	excl, ok := m["excludeSwitches"].([]interface{})
	if !ok || len(excl) != 1 || excl[0] != "enable-automation" {
		t.Errorf("excludeSwitches = %v, want [enable-automation]", m["excludeSwitches"])
	}

	prefs, ok := m["prefs"].(map[string]interface{})
	if !ok || prefs["download.default_directory"] != "/tmp" {
		t.Errorf("prefs = %v, want download.default_directory=/tmp", m["prefs"])
	}

	if _, ok := m["mobileEmulation"]; !ok {
		t.Errorf("mobileEmulation key missing from %v", m)
	}
	if _, ok := m["perfLoggingPrefs"]; !ok {
		t.Errorf("perfLoggingPrefs key missing from %v", m)
	}

	// Fields left at their zero value must be omitted.
	for _, key := range []string{"detach", "localState", "extensions", "w3c"} {
		if _, ok := m[key]; ok {
			t.Errorf("unexpected key %q present in %v", key, m)
		}
	}
}

func TestMobileEmulationMarshal(t *testing.T) {
	touch := true
	me := MobileEmulation{
		DeviceMetrics: &DeviceMetrics{
			Width:      360,
			Height:     640,
			PixelRatio: 3.0,
			Touch:      &touch,
		},
		UserAgent: "custom-agent",
	}

	m := unmarshalToMap(t, me)

	if _, ok := m["deviceName"]; ok {
		t.Errorf("deviceName should be omitted, got %v", m["deviceName"])
	}
	if got, want := m["userAgent"], "custom-agent"; got != want {
		t.Errorf("userAgent = %v, want %v", got, want)
	}

	dm, ok := m["deviceMetrics"].(map[string]interface{})
	if !ok {
		t.Fatalf("deviceMetrics = %v, want a map", m["deviceMetrics"])
	}
	// width, height and pixelRatio are always emitted (no omitempty).
	for _, key := range []string{"width", "height", "pixelRatio", "touch"} {
		if _, ok := dm[key]; !ok {
			t.Errorf("deviceMetrics missing key %q in %v", key, dm)
		}
	}
	if dm["width"].(float64) != 360 {
		t.Errorf("deviceMetrics.width = %v, want 360", dm["width"])
	}
}

func TestDeviceMetricsAlwaysEmitsDimensions(t *testing.T) {
	// Width/Height/PixelRatio have no omitempty, so a zero-value struct still
	// serializes them. Touch is a pointer and must be omitted when unset.
	m := unmarshalToMap(t, DeviceMetrics{})
	for _, key := range []string{"width", "height", "pixelRatio"} {
		if _, ok := m[key]; !ok {
			t.Errorf("DeviceMetrics{} missing key %q in %v", key, m)
		}
	}
	if _, ok := m["touch"]; ok {
		t.Errorf("touch should be omitted for zero-value DeviceMetrics, got %v", m["touch"])
	}
}

func TestPerfLoggingPreferencesMarshal(t *testing.T) {
	enableNet := false
	prefs := PerfLoggingPreferences{
		EnableNetwork:                      &enableNet,
		TraceCategories:                    "browser,devtools",
		BufferUsageReportingIntervalMillis: 1000,
	}

	m := unmarshalToMap(t, prefs)

	if got, ok := m["enableNetwork"].(bool); !ok || got != false {
		t.Errorf("enableNetwork = %v, want false", m["enableNetwork"])
	}
	if got, want := m["traceCategories"], "browser,devtools"; got != want {
		t.Errorf("traceCategories = %v, want %v", got, want)
	}
	if got, want := m["bufferUsageReportingInterval"], float64(1000); got != want {
		t.Errorf("bufferUsageReportingInterval = %v, want %v", got, want)
	}
	// Unset pointer fields are omitted.
	for _, key := range []string{"enablePage", "enableTimeline"} {
		if _, ok := m[key]; ok {
			t.Errorf("unexpected key %q present in %v", key, m)
		}
	}
}

func TestAddExtension(t *testing.T) {
	dir, err := os.MkdirTemp("", "chrome-ext")
	if err != nil {
		t.Fatalf("os.MkdirTemp returned error: %v", err)
	}
	defer os.RemoveAll(dir)

	payload := []byte("fake-crx-contents")
	path := filepath.Join(dir, "extension.crx")
	if err := os.WriteFile(path, payload, 0600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	var caps Capabilities
	if err := caps.AddExtension(path); err != nil {
		t.Fatalf("AddExtension returned error: %v", err)
	}

	if len(caps.Extensions) != 1 {
		t.Fatalf("len(Extensions) = %d, want 1", len(caps.Extensions))
	}

	decoded, err := base64.StdEncoding.DecodeString(caps.Extensions[0])
	if err != nil {
		t.Fatalf("stored extension is not valid base64: %v", err)
	}
	if len(decoded) == 0 {
		t.Errorf("decoded extension is empty")
	}
	if !bytes.Equal(decoded, payload) {
		t.Errorf("decoded extension = %q, want %q", decoded, payload)
	}
}

func TestAddExtensionMissingFile(t *testing.T) {
	var caps Capabilities
	if err := caps.AddExtension(filepath.Join(t.TempDir(), "does-not-exist.crx")); err == nil {
		t.Fatalf("AddExtension(nonexistent) = nil, want error")
	}
	if len(caps.Extensions) != 0 {
		t.Errorf("Extensions should be empty after a failed AddExtension, got %v", caps.Extensions)
	}
}

const crx3Magic = "Cr24"

// crx3HeaderLen is the fixed prefix: 4-byte magic + 4-byte version +
// 4-byte header length.
const crx3HeaderLen = 12

func writeExtensionDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "chrome-unpacked")
	if err != nil {
		t.Fatalf("os.MkdirTemp returned error: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	manifest := `{"name":"test","version":"1.0","manifest_version":2}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0600); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}
	return dir
}

func TestNewExtensionWithKey(t *testing.T) {
	dir := writeExtensionDir(t)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey returned error: %v", err)
	}

	data, err := NewExtensionWithKey(dir, key)
	if err != nil {
		t.Fatalf("NewExtensionWithKey returned error: %v", err)
	}

	if !bytes.HasPrefix(data, []byte(crx3Magic)) {
		t.Errorf("payload does not start with CRX3 magic %q: % x", crx3Magic, data[:4])
	}
	if len(data) <= crx3HeaderLen {
		t.Errorf("payload length = %d, want > %d (header only)", len(data), crx3HeaderLen)
	}
}

func TestNewExtension(t *testing.T) {
	dir := writeExtensionDir(t)

	data, key, err := NewExtension(dir)
	if err != nil {
		t.Fatalf("NewExtension returned error: %v", err)
	}
	if key == nil {
		t.Fatalf("NewExtension returned a nil private key")
	}
	if err := key.Validate(); err != nil {
		t.Errorf("returned private key is invalid: %v", err)
	}
	if !bytes.HasPrefix(data, []byte(crx3Magic)) {
		t.Errorf("payload does not start with CRX3 magic %q: % x", crx3Magic, data[:4])
	}
	if len(data) <= crx3HeaderLen {
		t.Errorf("payload length = %d, want > %d (header only)", len(data), crx3HeaderLen)
	}
}

func TestAddUnpackedExtension(t *testing.T) {
	dir := writeExtensionDir(t)

	var caps Capabilities
	if err := caps.AddUnpackedExtension(dir); err != nil {
		t.Fatalf("AddUnpackedExtension returned error: %v", err)
	}
	if len(caps.Extensions) != 1 {
		t.Fatalf("len(Extensions) = %d, want 1", len(caps.Extensions))
	}

	decoded, err := base64.StdEncoding.DecodeString(caps.Extensions[0])
	if err != nil {
		t.Fatalf("stored extension is not valid base64: %v", err)
	}
	if !bytes.HasPrefix(decoded, []byte(crx3Magic)) {
		t.Errorf("decoded extension does not start with CRX3 magic %q", crx3Magic)
	}
}

func ExampleCapabilities() {
	caps := Capabilities{
		Path: "/usr/bin/google-chrome",
		Args: []string{"--headless"},
	}
	data, _ := json.Marshal(map[string]interface{}{
		CapabilitiesKey: caps,
	})
	fmt.Println(strings.Contains(string(data), CapabilitiesKey))
	// Output: true
}
