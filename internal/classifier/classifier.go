package classifier

import (
	"fmt"
	"picklight/internal/config"
	"strconv"
	"strings"
)

type Result struct {
	Threshold config.Threshold
	R, G, B   byte
}

func Classify(value int, thresholds []config.Threshold) *Result {
	for _, t := range thresholds {
		if value >= t.Min && value <= t.Max {
			r, g, b, err := ParseHexColor(t.Color)
			if err != nil {
				continue
			}
			return &Result{Threshold: t, R: r, G: g, B: b}
		}
	}
	return nil
}

func ParseHexColor(hex string) (byte, byte, byte, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %q", hex)
	}
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid red: %w", err)
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid green: %w", err)
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid blue: %w", err)
	}
	return byte(r), byte(g), byte(b), nil
}
