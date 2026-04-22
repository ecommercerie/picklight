package statuslight

import "time"

// StatusLight is the interface implemented by all USB status light drivers.
type StatusLight interface {
	SetColor(r, g, b byte) error
	SetColorBlink(r, g, b byte) error
	SetColorAndSound(r, g, b, tone, volume byte, blink bool) error
	TurnOff() error
	KeepAlive() error
	Close() error
	DeviceName() string
	NeedsKeepAlive() bool
}

// USBDeviceInfo represents a discovered USB HID device.
type USBDeviceInfo struct {
	VendorID     string `json:"vendorId"`
	ProductID    string `json:"productId"`
	Product      string `json:"product"`
	Manufacturer string `json:"manufacturer"`
	Path         string `json:"path"`
}

// DebugInfo holds diagnostic info about a device interface.
type DebugInfo struct {
	Path                    string `json:"path"`
	OpenOK                  bool   `json:"openOk"`
	OpenError               string `json:"openError,omitempty"`
	VendorID                string `json:"vendorId,omitempty"`
	ProductID               string `json:"productId,omitempty"`
	Product                 string `json:"product,omitempty"`
	OutputReportByteLength  int    `json:"outputReportByteLength"`
	FeatureReportByteLength int    `json:"featureReportByteLength"`
	InputReportByteLength   int    `json:"inputReportByteLength"`
	Usage                   int    `json:"usage"`
	UsagePage               int    `json:"usagePage"`
	WriteFileResult         string `json:"writeFileResult"`
	SetOutputReportResult   string `json:"setOutputReportResult"`
	SetFeatureResult        string `json:"setFeatureResult"`
}

// DeviceID identifies a USB device by vendor and product IDs.
type DeviceID struct {
	VendorID  uint16
	ProductID uint16
}

// DriverFactory creates a StatusLight from an opened HID device.
type DriverFactory func(handle uintptr, path, productName string) StatusLight

// registration holds a factory and its associated vendor for fallback matching.
type registration struct {
	factory  DriverFactory
	vendorID uint16
}

var registry []registration

// Register adds a driver factory for the given device IDs.
func Register(factory DriverFactory, ids ...DeviceID) {
	for _, id := range ids {
		registry = append(registry, registration{
			factory:  factory,
			vendorID: id.VendorID,
		})
		exactRegistry[id] = factory
	}
}

var exactRegistry = map[DeviceID]DriverFactory{}

// lookupDriver finds a factory for the given vendor/product pair.
// Tries exact match first, then falls back to vendor-only match.
func lookupDriver(vendorID, productID uint16) (DriverFactory, bool) {
	if f, ok := exactRegistry[DeviceID{vendorID, productID}]; ok {
		return f, true
	}
	// Fallback: first registration matching vendor ID
	for _, r := range registry {
		if r.vendorID == vendorID {
			return r.factory, true
		}
	}
	return nil, false
}

// SetColorAndSoundFallback is a helper for drivers without buzzer hardware.
// Simulates a sound alert by flashing white briefly, then restoring the requested color.
func SetColorAndSoundFallback(light StatusLight, r, g, b byte, blink bool) error {
	// Flash white briefly to simulate the sound alert
	if err := light.SetColor(255, 255, 255); err != nil {
		return err
	}
	go func() {
		time.Sleep(500 * time.Millisecond)
		if blink {
			light.SetColorBlink(r, g, b)
		} else {
			light.SetColor(r, g, b)
		}
	}()
	return nil
}
