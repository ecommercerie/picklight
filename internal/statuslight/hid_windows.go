//go:build windows

package statuslight

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	setupapi    = syscall.NewLazyDLL("setupapi.dll")
	hiddll      = syscall.NewLazyDLL("hid.dll")
	kernel32dll = syscall.NewLazyDLL("kernel32.dll")

	pHidD_GetHidGuid            = hiddll.NewProc("HidD_GetHidGuid")
	pHidD_GetAttributes         = hiddll.NewProc("HidD_GetAttributes")
	pHidD_GetProductString      = hiddll.NewProc("HidD_GetProductString")
	pHidD_GetManufacturerString = hiddll.NewProc("HidD_GetManufacturerString")
	pHidD_SetOutputReport       = hiddll.NewProc("HidD_SetOutputReport")
	pHidD_SetFeature            = hiddll.NewProc("HidD_SetFeature")
	pHidP_GetCaps               = hiddll.NewProc("HidP_GetCaps")
	pHidD_GetPreparsedData      = hiddll.NewProc("HidD_GetPreparsedData")
	pHidD_FreePreparsedData     = hiddll.NewProc("HidD_FreePreparsedData")

	pSetupDiGetClassDevsW             = setupapi.NewProc("SetupDiGetClassDevsW")
	pSetupDiEnumDeviceInterfaces      = setupapi.NewProc("SetupDiEnumDeviceInterfaces")
	pSetupDiGetDeviceInterfaceDetailW = setupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
	pSetupDiDestroyDeviceInfoList     = setupapi.NewProc("SetupDiDestroyDeviceInfoList")

	pCreateFileW = kernel32dll.NewProc("CreateFileW")
	pCloseHandle = kernel32dll.NewProc("CloseHandle")
	pWriteFile   = kernel32dll.NewProc("WriteFile")
)

const (
	digcfPresent         = 0x00000002
	digcfDeviceInterface = 0x00000010
	genericWrite         = 0x40000000
	genericRead          = 0x80000000
	fileShareRead        = 0x00000001
	fileShareWrite       = 0x00000002
	openExisting         = 3
	InvalidHandleValue   = ^uintptr(0)
	hidpStatusSuccess    = 0x00110000
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type spDeviceInterfaceData struct {
	cbSize    uint32
	classGuid guid
	flags     uint32
	reserved  uintptr
}

// HIDAttributes holds USB HID device attributes.
type HIDAttributes struct {
	Size          uint32
	VendorID      uint16
	ProductID     uint16
	VersionNumber uint16
}

type hidpCaps struct {
	Usage                     uint16
	UsagePage                 uint16
	InputReportByteLength     uint16
	OutputReportByteLength    uint16
	FeatureReportByteLength   uint16
	Reserved                  [17]uint16
	NumberLinkCollectionNodes uint16
	NumberInputButtonCaps     uint16
	NumberInputValueCaps      uint16
	NumberInputDataIndices    uint16
	NumberOutputButtonCaps    uint16
	NumberOutputValueCaps     uint16
	NumberOutputDataIndices   uint16
	NumberFeatureButtonCaps   uint16
	NumberFeatureValueCaps    uint16
	NumberFeatureDataIndices  uint16
}

func getHidGuid() guid {
	var g guid
	pHidD_GetHidGuid.Call(uintptr(unsafe.Pointer(&g)))
	return g
}

// WriteReport sends an output report via WriteFile with 0x00 report ID prepended.
func WriteReport(handle uintptr, report []byte) error {
	if handle == 0 || handle == InvalidHandleValue {
		return fmt.Errorf("statuslight: device not open")
	}
	buf := make([]byte, len(report)+1)
	buf[0] = 0x00
	copy(buf[1:], report)

	var written uint32
	ret, _, err := pWriteFile.Call(handle, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), uintptr(unsafe.Pointer(&written)), 0)
	if ret == 0 {
		return fmt.Errorf("statuslight: WriteFile failed: %v", err)
	}
	return nil
}

// WriteReportRaw sends an output report via WriteFile without prepending a report ID.
func WriteReportRaw(handle uintptr, report []byte) error {
	if handle == 0 || handle == InvalidHandleValue {
		return fmt.Errorf("statuslight: device not open")
	}
	var written uint32
	ret, _, err := pWriteFile.Call(handle, uintptr(unsafe.Pointer(&report[0])), uintptr(len(report)), uintptr(unsafe.Pointer(&written)), 0)
	if ret == 0 {
		return fmt.Errorf("statuslight: WriteFile raw failed: %v", err)
	}
	return nil
}

// SendFeatureReport sends a feature report via HidD_SetFeature.
func SendFeatureReport(handle uintptr, report []byte) error {
	if handle == 0 || handle == InvalidHandleValue {
		return fmt.Errorf("statuslight: device not open")
	}
	buf := make([]byte, len(report)+1)
	buf[0] = 0x00
	copy(buf[1:], report)

	ret, _, err := pHidD_SetFeature.Call(handle, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 0 {
		return fmt.Errorf("statuslight: SetFeature failed: %v", err)
	}
	return nil
}

// CloseHandle closes a Windows handle.
func CloseHandle(handle uintptr) {
	if handle != 0 && handle != InvalidHandleValue {
		pCloseHandle.Call(handle)
	}
}

// ListAllHIDDevices enumerates all USB HID devices on the system.
func ListAllHIDDevices() []USBDeviceInfo {
	hidGuid := getHidGuid()
	hDevInfo, _, _ := pSetupDiGetClassDevsW.Call(
		uintptr(unsafe.Pointer(&hidGuid)), 0, 0,
		digcfPresent|digcfDeviceInterface,
	)
	if hDevInfo == InvalidHandleValue {
		return nil
	}
	defer pSetupDiDestroyDeviceInfoList.Call(hDevInfo)

	var result []USBDeviceInfo
	seen := make(map[string]bool)

	for i := uint32(0); ; i++ {
		var did spDeviceInterfaceData
		did.cbSize = uint32(unsafe.Sizeof(did))
		ret, _, _ := pSetupDiEnumDeviceInterfaces.Call(hDevInfo, 0, uintptr(unsafe.Pointer(&hidGuid)), uintptr(i), uintptr(unsafe.Pointer(&did)))
		if ret == 0 {
			break
		}

		path := getDevicePath(hDevInfo, &did)
		if path == "" {
			continue
		}

		handle := openDeviceReadOnly(path)
		if handle == InvalidHandleValue {
			continue
		}

		var attrs HIDAttributes
		attrs.Size = uint32(unsafe.Sizeof(attrs))
		ret, _, _ = pHidD_GetAttributes.Call(handle, uintptr(unsafe.Pointer(&attrs)))
		if ret == 0 {
			pCloseHandle.Call(handle)
			continue
		}

		key := fmt.Sprintf("%04X:%04X", attrs.VendorID, attrs.ProductID)
		if seen[key] {
			pCloseHandle.Call(handle)
			continue
		}
		seen[key] = true

		product := getStringDescriptor(handle, pHidD_GetProductString)
		manufacturer := getStringDescriptor(handle, pHidD_GetManufacturerString)
		pCloseHandle.Call(handle)

		result = append(result, USBDeviceInfo{
			VendorID:     fmt.Sprintf("0x%04X", attrs.VendorID),
			ProductID:    fmt.Sprintf("0x%04X", attrs.ProductID),
			Product:      product,
			Manufacturer: manufacturer,
			Path:         path,
		})
	}
	return result
}

// DiagnoseDevices tries every known-vendor interface with all write methods.
func DiagnoseDevices() []DebugInfo {
	hidGuid := getHidGuid()
	hDevInfo, _, _ := pSetupDiGetClassDevsW.Call(
		uintptr(unsafe.Pointer(&hidGuid)), 0, 0,
		digcfPresent|digcfDeviceInterface,
	)
	if hDevInfo == InvalidHandleValue {
		return nil
	}
	defer pSetupDiDestroyDeviceInfoList.Call(hDevInfo)

	var results []DebugInfo

	for i := uint32(0); ; i++ {
		var did spDeviceInterfaceData
		did.cbSize = uint32(unsafe.Sizeof(did))
		ret, _, _ := pSetupDiEnumDeviceInterfaces.Call(hDevInfo, 0, uintptr(unsafe.Pointer(&hidGuid)), uintptr(i), uintptr(unsafe.Pointer(&did)))
		if ret == 0 {
			break
		}

		path := getDevicePath(hDevInfo, &did)
		if path == "" {
			continue
		}

		roHandle := openDeviceReadOnly(path)
		if roHandle == InvalidHandleValue {
			continue
		}
		var attrs HIDAttributes
		attrs.Size = uint32(unsafe.Sizeof(attrs))
		ret, _, _ = pHidD_GetAttributes.Call(roHandle, uintptr(unsafe.Pointer(&attrs)))
		if ret == 0 {
			pCloseHandle.Call(roHandle)
			continue
		}

		// Only diagnose devices we have a driver for
		if _, ok := lookupDriver(attrs.VendorID, attrs.ProductID); !ok {
			pCloseHandle.Call(roHandle)
			continue
		}

		product := getStringDescriptor(roHandle, pHidD_GetProductString)
		pCloseHandle.Call(roHandle)

		info := DebugInfo{
			Path:      path,
			VendorID:  fmt.Sprintf("0x%04X", attrs.VendorID),
			ProductID: fmt.Sprintf("0x%04X", attrs.ProductID),
			Product:   product,
		}

		handle := openDeviceRW(path)
		if handle == InvalidHandleValue {
			info.OpenOK = false
			info.OpenError = "CreateFile RW failed"
			results = append(results, info)
			continue
		}
		info.OpenOK = true

		caps := getDeviceCaps(handle)
		if caps != nil {
			info.OutputReportByteLength = int(caps.OutputReportByteLength)
			info.FeatureReportByteLength = int(caps.FeatureReportByteLength)
			info.InputReportByteLength = int(caps.InputReportByteLength)
			info.Usage = int(caps.Usage)
			info.UsagePage = int(caps.UsagePage)
		}

		// Try a zero report with each method
		testReport := make([]byte, 64)

		info.WriteFileResult = tryWriteFile(handle, testReport)
		info.SetOutputReportResult = tryWriteFileRaw(handle, testReport)
		info.SetFeatureResult = trySetFeature(handle, testReport)

		pCloseHandle.Call(handle)
		results = append(results, info)
	}

	return results
}

// IsConnected checks if any supported status light is present.
func IsConnected() bool {
	hidGuid := getHidGuid()
	hDevInfo, _, _ := pSetupDiGetClassDevsW.Call(
		uintptr(unsafe.Pointer(&hidGuid)), 0, 0,
		digcfPresent|digcfDeviceInterface,
	)
	if hDevInfo == InvalidHandleValue {
		return false
	}
	defer pSetupDiDestroyDeviceInfoList.Call(hDevInfo)

	for i := uint32(0); ; i++ {
		var did spDeviceInterfaceData
		did.cbSize = uint32(unsafe.Sizeof(did))
		ret, _, _ := pSetupDiEnumDeviceInterfaces.Call(hDevInfo, 0, uintptr(unsafe.Pointer(&hidGuid)), uintptr(i), uintptr(unsafe.Pointer(&did)))
		if ret == 0 {
			break
		}
		path := getDevicePath(hDevInfo, &did)
		if path == "" {
			continue
		}
		handle := openDeviceReadOnly(path)
		if handle == InvalidHandleValue {
			continue
		}
		var attrs HIDAttributes
		attrs.Size = uint32(unsafe.Sizeof(attrs))
		ret, _, _ = pHidD_GetAttributes.Call(handle, uintptr(unsafe.Pointer(&attrs)))
		pCloseHandle.Call(handle)
		if ret == 0 {
			continue
		}
		if _, ok := lookupDriver(attrs.VendorID, attrs.ProductID); ok {
			return true
		}
	}
	return false
}

// Detect scans for and opens the first supported status light device.
func Detect() (StatusLight, error) {
	hidGuid := getHidGuid()
	hDevInfo, _, _ := pSetupDiGetClassDevsW.Call(
		uintptr(unsafe.Pointer(&hidGuid)), 0, 0,
		digcfPresent|digcfDeviceInterface,
	)
	if hDevInfo == InvalidHandleValue {
		return nil, fmt.Errorf("statuslight: SetupDiGetClassDevs failed")
	}
	defer pSetupDiDestroyDeviceInfoList.Call(hDevInfo)

	for i := uint32(0); ; i++ {
		var did spDeviceInterfaceData
		did.cbSize = uint32(unsafe.Sizeof(did))
		ret, _, _ := pSetupDiEnumDeviceInterfaces.Call(hDevInfo, 0, uintptr(unsafe.Pointer(&hidGuid)), uintptr(i), uintptr(unsafe.Pointer(&did)))
		if ret == 0 {
			break
		}

		path := getDevicePath(hDevInfo, &did)
		if path == "" {
			continue
		}

		handle := openDeviceRW(path)
		if handle == InvalidHandleValue {
			continue
		}

		var attrs HIDAttributes
		attrs.Size = uint32(unsafe.Sizeof(attrs))
		ret, _, _ = pHidD_GetAttributes.Call(handle, uintptr(unsafe.Pointer(&attrs)))
		if ret == 0 {
			pCloseHandle.Call(handle)
			continue
		}

		factory, ok := lookupDriver(attrs.VendorID, attrs.ProductID)
		if !ok {
			pCloseHandle.Call(handle)
			continue
		}

		product := getStringDescriptor(handle, pHidD_GetProductString)
		return factory(handle, path, product), nil
	}

	return nil, fmt.Errorf("statuslight: no supported device found")
}

func tryWriteFile(handle uintptr, report []byte) string {
	buf := make([]byte, len(report)+1)
	buf[0] = 0x00
	copy(buf[1:], report)
	var written uint32
	ret, _, err := pWriteFile.Call(handle, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), uintptr(unsafe.Pointer(&written)), 0)
	if ret == 0 {
		return fmt.Sprintf("FAIL: %v (with reportID)", err)
	}
	return fmt.Sprintf("OK (wrote %d bytes, with reportID)", written)
}

func tryWriteFileRaw(handle uintptr, report []byte) string {
	var written uint32
	ret, _, err := pWriteFile.Call(handle, uintptr(unsafe.Pointer(&report[0])), uintptr(len(report)), uintptr(unsafe.Pointer(&written)), 0)
	if ret == 0 {
		return fmt.Sprintf("FAIL: %v (raw)", err)
	}
	return fmt.Sprintf("OK (wrote %d bytes, raw)", written)
}

func trySetFeature(handle uintptr, report []byte) string {
	buf := make([]byte, len(report)+1)
	buf[0] = 0x00
	copy(buf[1:], report)
	ret, _, err := pHidD_SetFeature.Call(handle, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 0 {
		return fmt.Sprintf("FAIL: %v", err)
	}
	return "OK"
}

func getDeviceCaps(handle uintptr) *hidpCaps {
	var preparsedData uintptr
	ret, _, _ := pHidD_GetPreparsedData.Call(handle, uintptr(unsafe.Pointer(&preparsedData)))
	if ret == 0 {
		return nil
	}
	defer pHidD_FreePreparsedData.Call(preparsedData)
	var caps hidpCaps
	ret, _, _ = pHidP_GetCaps.Call(preparsedData, uintptr(unsafe.Pointer(&caps)))
	if ret != hidpStatusSuccess {
		return nil
	}
	return &caps
}

func openDeviceRW(path string) uintptr {
	pathPtr, _ := syscall.UTF16PtrFromString(path)
	handle, _, _ := pCreateFileW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		genericWrite|genericRead,
		fileShareRead|fileShareWrite,
		0, openExisting, 0, 0,
	)
	return handle
}

func openDeviceReadOnly(path string) uintptr {
	pathPtr, _ := syscall.UTF16PtrFromString(path)
	handle, _, _ := pCreateFileW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		genericRead,
		fileShareRead|fileShareWrite,
		0, openExisting, 0, 0,
	)
	return handle
}

func getDevicePath(hDevInfo uintptr, did *spDeviceInterfaceData) string {
	var requiredSize uint32
	pSetupDiGetDeviceInterfaceDetailW.Call(hDevInfo, uintptr(unsafe.Pointer(did)), 0, 0, uintptr(unsafe.Pointer(&requiredSize)), 0)
	if requiredSize == 0 {
		return ""
	}
	buf := make([]byte, requiredSize)
	*(*uint32)(unsafe.Pointer(&buf[0])) = 8
	ret, _, _ := pSetupDiGetDeviceInterfaceDetailW.Call(hDevInfo, uintptr(unsafe.Pointer(did)), uintptr(unsafe.Pointer(&buf[0])), uintptr(requiredSize), 0, 0)
	if ret == 0 {
		return ""
	}
	pathPtr := (*[1 << 16]uint16)(unsafe.Pointer(&buf[4]))
	return syscall.UTF16ToString(pathPtr[:])
}

func getStringDescriptor(handle uintptr, proc *syscall.LazyProc) string {
	buf := make([]uint16, 128)
	ret, _, _ := proc.Call(handle, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)*2))
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
