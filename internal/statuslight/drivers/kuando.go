package drivers

import (
	"time"

	"picklight/internal/statuslight"
)

const (
	kuandoReportLen       = 64
	kuandoOpcodeJump      = 0x10 // Jump to step 0
	kuandoOpcodeKeepAlive = 0x8F // KeepAlive with 15s timeout
)

// KuandoBusylight implements StatusLight for Kuando Busylight Alpha and Omega devices.
type KuandoBusylight struct {
	handle     uintptr
	deviceName string
	path       string
}

func init() {
	factory := func(handle uintptr, path, productName string) statuslight.StatusLight {
		return &KuandoBusylight{handle: handle, deviceName: productName, path: path}
	}
	// Alpha
	statuslight.Register(factory,
		statuslight.DeviceID{VendorID: 0x04D8, ProductID: 0xF848},
		statuslight.DeviceID{VendorID: 0x27BB, ProductID: 0x3BCA},
		statuslight.DeviceID{VendorID: 0x27BB, ProductID: 0x3BCB},
		statuslight.DeviceID{VendorID: 0x27BB, ProductID: 0x3BCE},
	)
	// Omega
	statuslight.Register(factory,
		statuslight.DeviceID{VendorID: 0x27BB, ProductID: 0x3BCD},
		statuslight.DeviceID{VendorID: 0x27BB, ProductID: 0x3BCF},
	)
}

func (k *KuandoBusylight) DeviceName() string    { return k.deviceName }
func (k *KuandoBusylight) NeedsKeepAlive() bool  { return true }

func (k *KuandoBusylight) Close() error {
	statuslight.CloseHandle(k.handle)
	k.handle = 0
	return nil
}

func (k *KuandoBusylight) SetColor(r, g, b byte) error {
	return statuslight.WriteReport(k.handle, kuandoBuildReport(kuandoOpcodeJump, r, g, b, 0, 0, 0, 0))
}

func (k *KuandoBusylight) SetColorBlink(r, g, b byte) error {
	return statuslight.WriteReport(k.handle, kuandoBuildReport(kuandoOpcodeJump, r, g, b, 4, 4, 0, 0))
}

func (k *KuandoBusylight) SetColorAndSound(r, g, b, tone, volume byte, blink bool) error {
	var onTime, offTime byte
	if blink {
		onTime, offTime = 4, 4
	}
	report := kuandoBuildReport(kuandoOpcodeJump, r, g, b, onTime, offTime, tone, volume)
	if err := statuslight.WriteReport(k.handle, report); err != nil {
		return err
	}
	go func() {
		time.Sleep(2 * time.Second)
		if blink {
			k.SetColorBlink(r, g, b)
		} else {
			k.SetColor(r, g, b)
		}
	}()
	return nil
}

func (k *KuandoBusylight) TurnOff() error {
	return statuslight.WriteReport(k.handle, kuandoBuildReport(kuandoOpcodeJump, 0, 0, 0, 0, 0, 0, 0))
}

func (k *KuandoBusylight) KeepAlive() error {
	return statuslight.WriteReport(k.handle, kuandoBuildKeepAlive())
}

// kuandoScaleColor converts 0-255 RGB to 0-100 device scale.
func kuandoScaleColor(v byte) byte {
	return byte((int(v) * 100) / 255)
}

// kuandoBuildReport constructs a 64-byte Busylight report.
func kuandoBuildReport(opcode byte, r, g, b, dutyCycleOn, dutyCycleOff, ringtone, volume byte) []byte {
	report := make([]byte, kuandoReportLen)

	// Step 0
	report[0] = opcode
	report[1] = 0x00 // repeat
	report[2] = kuandoScaleColor(r)
	report[3] = kuandoScaleColor(g)
	report[4] = kuandoScaleColor(b)
	report[5] = dutyCycleOn
	report[6] = dutyCycleOff
	if ringtone > 0 {
		report[7] = 0x80 | (ringtone << 3) | (volume & 0x07)
	}

	// Steps 1-6: zeros (bytes 8-55)

	// Footer (bytes 56-63)
	report[60] = 0x0F
	report[61] = 0xFF

	// Checksum: sum of bytes 0-61
	var checksum uint16
	for i := 0; i < 62; i++ {
		checksum += uint16(report[i])
	}
	report[62] = byte(checksum >> 8)
	report[63] = byte(checksum & 0xFF)

	return report
}

// kuandoBuildKeepAlive constructs a keepalive report.
func kuandoBuildKeepAlive() []byte {
	report := make([]byte, kuandoReportLen)
	report[0] = kuandoOpcodeKeepAlive

	report[60] = 0x0F
	report[61] = 0xFF

	var checksum uint16
	for i := 0; i < 62; i++ {
		checksum += uint16(report[i])
	}
	report[62] = byte(checksum >> 8)
	report[63] = byte(checksum & 0xFF)

	return report
}
