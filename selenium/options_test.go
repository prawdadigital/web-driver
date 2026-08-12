package selenium

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsDisplayExtra(t *testing.T) {
	// Complements TestIsDisplay in service_test.go with the specific inputs
	// called out by the exercise: "1", "1.0" (valid) and "a", "1.2.3" (invalid).
	tests := []struct {
		in   string
		want bool
	}{
		{"1", true},
		{"1.0", true},
		{"a", false},
		{"1.2.3", false},
	}
	for _, tt := range tests {
		if got := isDisplay(tt.in); got != tt.want {
			t.Errorf("isDisplay(%q) = %t, want %t", tt.in, got, tt.want)
		}
	}
}

func TestDisplayOption(t *testing.T) {
	t.Run("sets display and xauthPath", func(t *testing.T) {
		s := &Service{}
		if err := Display("1.0", "/tmp/xauth")(s); err != nil {
			t.Fatalf("Display returned error: %v", err)
		}
		if s.display != "1.0" {
			t.Errorf("s.display = %q, want %q", s.display, "1.0")
		}
		if s.xauthPath != "/tmp/xauth" {
			t.Errorf("s.xauthPath = %q, want %q", s.xauthPath, "/tmp/xauth")
		}
	})

	t.Run("second Display returns error", func(t *testing.T) {
		s := &Service{}
		if err := Display("1", "/tmp/xauth")(s); err != nil {
			t.Fatalf("first Display returned error: %v", err)
		}
		if err := Display("2", "/tmp/other")(s); err == nil {
			t.Error("second Display: expected error, got nil")
		}
	})

	t.Run("invalid display value returns error", func(t *testing.T) {
		s := &Service{}
		if err := Display("not-a-display", "/tmp/xauth")(s); err == nil {
			t.Error("Display with invalid value: expected error, got nil")
		}
	})
}

func TestOutputOption(t *testing.T) {
	s := &Service{}
	var buf bytes.Buffer
	if err := Output(&buf)(s); err != nil {
		t.Fatalf("Output returned error: %v", err)
	}
	if s.output != &buf {
		t.Errorf("s.output = %v, want %v", s.output, &buf)
	}
}

func TestPathOptions(t *testing.T) {
	tests := []struct {
		name   string
		opt    ServiceOption
		getter func(*Service) string
		want   string
	}{
		{"GeckoDriver", GeckoDriver("/bin/geckodriver"), func(s *Service) string { return s.geckoDriverPath }, "/bin/geckodriver"},
		{"ChromeDriver", ChromeDriver("/bin/chromedriver"), func(s *Service) string { return s.chromeDriverPath }, "/bin/chromedriver"},
		{"JavaPath", JavaPath("/bin/java"), func(s *Service) string { return s.javaPath }, "/bin/java"},
		{"HTMLUnit", HTMLUnit("/lib/htmlunit.jar"), func(s *Service) string { return s.htmlUnitPath }, "/lib/htmlunit.jar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{}
			if err := tt.opt(s); err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}
			if got := tt.getter(s); got != tt.want {
				t.Errorf("%s field = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestStartFrameBufferOptionConstruction(t *testing.T) {
	// Verify StartFrameBufferWithOptions and StartFrameBuffer return usable
	// ServiceOption values without actually starting Xvfb. We deliberately do
	// not invoke the returned option (which would launch the frame buffer).
	if opt := StartFrameBufferWithOptions(FrameBufferOptions{ScreenSize: "1024x768x24"}); opt == nil {
		t.Error("StartFrameBufferWithOptions returned nil ServiceOption")
	}
	if opt := StartFrameBuffer(); opt == nil {
		t.Error("StartFrameBuffer returned nil ServiceOption")
	}
	// FrameBufferOptions is a plain struct; confirm the field round-trips.
	o := FrameBufferOptions{ScreenSize: "800x600"}
	if o.ScreenSize != "800x600" {
		t.Errorf("FrameBufferOptions.ScreenSize = %q, want %q", o.ScreenSize, "800x600")
	}
}

func TestSetDebugCallable(t *testing.T) {
	// SetDebug simply toggles a package-global in remote; ensure both values are
	// callable and leave debugging disabled afterward.
	SetDebug(true)
	SetDebug(false)
}

func TestNewRemoteAndDeleteSession(t *testing.T) {
	var mu struct {
		createHits int
		deletePath string
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/session"):
			mu.createHits++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"value":{"sessionId":"s1","capabilities":{}}}`))
		case r.Method == http.MethodDelete:
			mu.deletePath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"value":null}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	drv, err := NewRemote(nil, srv.URL)
	if err != nil {
		t.Fatalf("NewRemote returned error: %v", err)
	}
	if drv == nil {
		t.Fatal("NewRemote returned nil driver")
	}
	if got := drv.SessionID(); got != "s1" {
		t.Errorf("SessionID = %q, want %q", got, "s1")
	}
	if mu.createHits == 0 {
		t.Error("expected POST /session to be called")
	}

	if err := DeleteSession(srv.URL, "s1"); err != nil {
		t.Fatalf("DeleteSession returned error: %v", err)
	}
	if !strings.HasSuffix(mu.deletePath, "/session/s1") {
		t.Errorf("DeleteSession requested path %q, want suffix %q", mu.deletePath, "/session/s1")
	}
}
