package prompt

import (
	"strings"
	"testing"
)

func TestHexToANSI(t *testing.T) {
	result := HexToANSI("#ff0000")
	if !strings.HasPrefix(result, "\033[38;2;") {
		t.Errorf("expected ANSI escape, got %q", result)
	}
}

func TestHexToANSIInvalid(t *testing.T) {
	result := HexToANSI("nothex")
	if result != "\033[0m" {
		t.Errorf("expected reset on invalid hex, got %q", result)
	}
}

func TestHexToANSIEmpty(t *testing.T) {
	result := HexToANSI("")
	if result != "\033[0m" {
		t.Errorf("expected reset on empty, got %q", result)
	}
}

func TestFormatUserHost(t *testing.T) {
	result := Format("[%u@%h]", "alice", "box", "/tmp", "#4287f5", 0)
	if !strings.Contains(result, "alice") {
		t.Errorf("expected format to include username")
	}
	if !strings.Contains(result, "box") {
		t.Errorf("expected format to include hostname")
	}
}

func TestFormatTime(t *testing.T) {
	result := Format("%t", "u", "h", "/", "#4287f5", 0)
	if !strings.Contains(result, ":") {
		t.Errorf("expected %%t (HH:MM) to contain colon, got %q", result)
	}
}

func TestFormatTimeSeconds(t *testing.T) {
	result := Format("%T", "u", "h", "/", "#4287f5", 0)
	parts := strings.Split(result, ":")
	if len(parts) != 3 {
		t.Errorf("expected %%T (HH:MM:SS) to have 2 colons, got %q", result)
	}
}

func TestFormatExitCode(t *testing.T) {
	result := Format("%?", "u", "h", "/", "#4287f5", 0)
	if !strings.Contains(result, "0") {
		t.Errorf("expected %%? with exit=0 to show 0, got %q", result)
	}

	result = Format("[%?]", "u", "h", "/", "#4287f5", 42)
	if !strings.Contains(result, "42") {
		t.Errorf("expected %%? with exit=42 to show 42, got %q", result)
	}
}

func TestFormatRootIndicator(t *testing.T) {
	result := Format("%$", "u", "h", "/", "#4287f5", 0)
	if !strings.Contains(result, "$") {
		t.Errorf("expected %%$ for non-root to show $, got %q", result)
	}
}

func TestFormatDefault(t *testing.T) {
	result := Format("", "u", "h", "/d", "#4287f5", 0)
	if !strings.Contains(result, "/d") {
		t.Errorf("expected default format to include directory")
	}
}
