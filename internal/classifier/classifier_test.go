package classifier

import (
	"picklight/internal/config"
	"testing"
)

func TestClassifyMatch(t *testing.T) {
	thresholds := config.DefaultThresholds()
	result := Classify(0, thresholds)
	if result == nil {
		t.Fatal("expected match for 0")
	}
	if result.R != 0x00 || result.G != 0xFF || result.B != 0x00 {
		t.Errorf("expected green, got %02X%02X%02X", result.R, result.G, result.B)
	}
}

func TestClassifyOrange(t *testing.T) {
	thresholds := config.DefaultThresholds()
	result := Classify(3, thresholds)
	if result == nil {
		t.Fatal("expected match for 3")
	}
	if result.Threshold.Label != "Quelques commandes" {
		t.Errorf("expected 'Quelques commandes', got %q", result.Threshold.Label)
	}
}

func TestClassifyRed(t *testing.T) {
	thresholds := config.DefaultThresholds()
	result := Classify(10, thresholds)
	if result == nil {
		t.Fatal("expected match for 10")
	}
	if result.R != 0xFF || result.G != 0x00 || result.B != 0x00 {
		t.Errorf("expected red, got %02X%02X%02X", result.R, result.G, result.B)
	}
}

func TestClassifyNoMatch(t *testing.T) {
	thresholds := []config.Threshold{
		{Min: 1, Max: 5, Color: "#FFFFFF"},
	}
	result := Classify(0, thresholds)
	if result != nil {
		t.Error("expected no match for 0")
	}
}

func TestParseHexColor(t *testing.T) {
	r, g, b, err := ParseHexColor("#FF8000")
	if err != nil {
		t.Fatal(err)
	}
	if r != 255 || g != 128 || b != 0 {
		t.Errorf("expected 255,128,0, got %d,%d,%d", r, g, b)
	}
}

func TestParseHexColorNoHash(t *testing.T) {
	r, g, b, err := ParseHexColor("00FF00")
	if err != nil {
		t.Fatal(err)
	}
	if r != 0 || g != 255 || b != 0 {
		t.Errorf("expected 0,255,0, got %d,%d,%d", r, g, b)
	}
}

func TestParseHexColorInvalid(t *testing.T) {
	_, _, _, err := ParseHexColor("xyz")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}
