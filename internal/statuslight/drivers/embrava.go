package drivers

import (
	"picklight/internal/statuslight"
)

const (
	embravaTrailer1 = 0xFF
	embravaTrailer2 = 0x22
)

// EmbravaBlynclight implements StatusLight for Embrava Blynclight family devices.
// Note: the color order is R, B, G (not R, G, B).
type EmbravaBlynclight struct {
	handle     uintptr
	deviceName string
	path       string
}

func init() {
	factory := func(handle uintptr, path, productName string) statuslight.StatusLight {
		return &EmbravaBlynclight{handle: handle, deviceName: productName, path: path}
	}
	statuslight.Register(factory,
		statuslight.DeviceID{VendorID: 0x2C0D, ProductID: 0x0001},
		statuslight.DeviceID{VendorID: 0x2C0D, ProductID: 0x000C},
		statuslight.DeviceID{VendorID: 0x0E53, ProductID: 0x2516},
	)
}

func (e *EmbravaBlynclight) DeviceName() string    { return e.deviceName }
func (e *EmbravaBlynclight) NeedsKeepAlive() bool  { return false }
func (e *EmbravaBlynclight) KeepAlive() error      { return nil }

func (e *EmbravaBlynclight) Close() error {
	statuslight.CloseHandle(e.handle)
	e.handle = 0
	return nil
}

func (e *EmbravaBlynclight) SetColor(r, g, b byte) error {
	// [R, B, G, flags=0 (on), audio=0, trailer]
	report := []byte{r, b, g, 0x00, 0x00, embravaTrailer1, embravaTrailer2}
	return statuslight.WriteReport(e.handle, report)
}

func (e *EmbravaBlynclight) SetColorBlink(r, g, b byte) error {
	// flags: bit 2=flash(1), bits 3-5=speed(slow=001) → 0x04 | 0x08 = 0x0C
	// Actually: bit 2=flash → 0x04, bits 3-5=speed 001 → 0x08, combined = 0x0C
	// But slow speed value is 1, shifted to bits 3-5: 1<<3 = 0x08
	// flash=1 at bit 2: 0x04
	// total: 0x0C
	report := []byte{r, b, g, 0x0C, 0x00, embravaTrailer1, embravaTrailer2}
	return statuslight.WriteReport(e.handle, report)
}

func (e *EmbravaBlynclight) SetColorAndSound(r, g, b, tone, volume byte, blink bool) error {
	return statuslight.SetColorAndSoundFallback(e, r, g, b, blink)
}

func (e *EmbravaBlynclight) TurnOff() error {
	// flags: bit 0=off → 0x01
	report := []byte{0, 0, 0, 0x01, 0x00, embravaTrailer1, embravaTrailer2}
	return statuslight.WriteReport(e.handle, report)
}
