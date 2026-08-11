package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew_ValidLevel(t *testing.T) {
	var buf bytes.Buffer

	log, err := New("debug", &buf)
	if err != nil {
		t.Fatalf("New(debug) error: %v", err)
	}

	log.Debug().Msg("hello")

	out := buf.String()
	if !strings.Contains(out, `"level":"debug"`) {
		t.Errorf("output %q missing level field", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("output %q missing message", out)
	}
	if !strings.Contains(out, `"time":`) {
		t.Errorf("output %q missing timestamp field", out)
	}
	if !strings.Contains(out, `"caller":`) {
		t.Errorf("output %q missing caller field", out)
	}
}

func TestNew_EmptyLevelDefaultsToInfo(t *testing.T) {
	var buf bytes.Buffer

	log, err := New("", &buf)
	if err != nil {
		t.Fatalf("New(empty) error: %v", err)
	}

	log.Debug().Msg("debug line")
	log.Info().Msg("info line")

	out := buf.String()
	if strings.Contains(out, "debug line") {
		t.Errorf("empty level should default to info: debug output %q was emitted", out)
	}
	if !strings.Contains(out, "info line") {
		t.Errorf("empty level should default to info: info output missing from %q", out)
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	_, err := New("not-a-level", nil)
	if err == nil {
		t.Fatal("New(invalid) expected error, got nil")
	}
	if !strings.Contains(err.Error(), `parse log level "not-a-level"`) {
		t.Errorf("error %q missing level context", err)
	}
}

func TestNew_NilWriterUsesStderr(t *testing.T) {
	// Nil writer must not panic and must produce a usable logger. We cannot
	// capture stderr, so verify the error level is accepted and a log call
	// does not panic.
	log, err := New("error", nil)
	if err != nil {
		t.Fatalf("New(level, nil) error: %v", err)
	}
	log.Error().Msg("error line")
}
