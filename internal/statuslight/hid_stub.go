//go:build !windows

package statuslight

import "fmt"

const InvalidHandleValue = ^uintptr(0)

func WriteReport(handle uintptr, report []byte) error {
	return fmt.Errorf("statuslight: not supported on this platform")
}

func WriteReportRaw(handle uintptr, report []byte) error {
	return fmt.Errorf("statuslight: not supported on this platform")
}

func SendFeatureReport(handle uintptr, report []byte) error {
	return fmt.Errorf("statuslight: not supported on this platform")
}

func CloseHandle(handle uintptr) {}

func ListAllHIDDevices() []USBDeviceInfo {
	return nil
}

func DiagnoseDevices() []DebugInfo {
	return nil
}

func IsConnected() bool {
	return false
}

func Detect() (StatusLight, error) {
	return nil, fmt.Errorf("statuslight: not supported on this platform")
}
