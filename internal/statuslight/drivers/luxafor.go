package drivers

import (
	"picklight/internal/statuslight"
)

const (
	luxaforCmdColor  = 0x01
	luxaforCmdStrobe = 0x03
	luxaforLEDAll    = 0xFF
)

// LuxaforFlag implements StatusLight for Luxafor Flag devices.
type LuxaforFlag struct {
	handle     uintptr
	deviceName string
	path       string
}

func init() {
	factory := func(handle uintptr, path, productName string) statuslight.StatusLight {
		return &LuxaforFlag{handle: handle, deviceName: productName, path: path}
	}
	statuslight.Register(factory,
		statuslight.DeviceID{VendorID: 0x04D8, ProductID: 0xF372},
	)
}

func (l *LuxaforFlag) DeviceName() string    { return l.deviceName }
func (l *LuxaforFlag) NeedsKeepAlive() bool  { return false }
func (l *LuxaforFlag) KeepAlive() error      { return nil }

func (l *LuxaforFlag) Close() error {
	statuslight.CloseHandle(l.handle)
	l.handle = 0
	return nil
}

func (l *LuxaforFlag) SetColor(r, g, b byte) error {
	report := []byte{luxaforCmdColor, luxaforLEDAll, r, g, b}
	return statuslight.WriteReport(l.handle, report)
}

func (l *LuxaforFlag) SetColorBlink(r, g, b byte) error {
	// Strobe command: [cmd, leds, R, G, B, speed, repeat]
	report := []byte{luxaforCmdStrobe, luxaforLEDAll, r, g, b, 0x14, 0x00}
	return statuslight.WriteReport(l.handle, report)
}

func (l *LuxaforFlag) SetColorAndSound(r, g, b, tone, volume byte, blink bool) error {
	// No buzzer hardware — fall back to color/blink
	return statuslight.SetColorAndSoundFallback(l, r, g, b, blink)
}

func (l *LuxaforFlag) TurnOff() error {
	report := []byte{luxaforCmdColor, luxaforLEDAll, 0, 0, 0}
	return statuslight.WriteReport(l.handle, report)
}
