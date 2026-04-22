package drivers

import (
	"sync"
	"time"

	"picklight/internal/statuslight"
)

const (
	blink1ReportOne     = 0x01
	blink1ActionSetColor = 0x6E // 'n'
)

// Blink1 implements StatusLight for ThingM Blink(1) devices.
// Uses HID feature reports instead of output reports.
type Blink1 struct {
	handle     uintptr
	deviceName string
	path       string

	// Software blink state
	blinkMu   sync.Mutex
	blinkStop chan struct{}
}

func init() {
	factory := func(handle uintptr, path, productName string) statuslight.StatusLight {
		return &Blink1{handle: handle, deviceName: productName, path: path}
	}
	statuslight.Register(factory,
		statuslight.DeviceID{VendorID: 0x27B8, ProductID: 0x01ED},
	)
}

func (b *Blink1) DeviceName() string    { return b.deviceName }
func (b *Blink1) NeedsKeepAlive() bool  { return false }
func (b *Blink1) KeepAlive() error      { return nil }

func (b *Blink1) Close() error {
	b.stopBlink()
	statuslight.CloseHandle(b.handle)
	b.handle = 0
	return nil
}

func (b *Blink1) SetColor(r, g, bb byte) error {
	b.stopBlink()
	return b.sendColor(r, g, bb)
}

func (b *Blink1) SetColorBlink(r, g, bb byte) error {
	b.stopBlink()

	b.blinkMu.Lock()
	b.blinkStop = make(chan struct{})
	stop := b.blinkStop
	b.blinkMu.Unlock()

	// Software blink: alternate color/off every 500ms
	go func() {
		on := true
		for {
			if on {
				b.sendColor(r, g, bb)
			} else {
				b.sendColor(0, 0, 0)
			}
			on = !on
			select {
			case <-stop:
				return
			case <-time.After(500 * time.Millisecond):
			}
		}
	}()
	return nil
}

func (b *Blink1) SetColorAndSound(r, g, bb, tone, volume byte, blink bool) error {
	return statuslight.SetColorAndSoundFallback(b, r, g, bb, blink)
}

func (b *Blink1) TurnOff() error {
	b.stopBlink()
	return b.sendColor(0, 0, 0)
}

func (b *Blink1) sendColor(r, g, bb byte) error {
	report := []byte{blink1ReportOne, blink1ActionSetColor, r, g, bb, 0x00, 0x00, 0x00}
	return statuslight.SendFeatureReport(b.handle, report)
}

func (b *Blink1) stopBlink() {
	b.blinkMu.Lock()
	defer b.blinkMu.Unlock()
	if b.blinkStop != nil {
		select {
		case <-b.blinkStop:
		default:
			close(b.blinkStop)
		}
		b.blinkStop = nil
	}
}
