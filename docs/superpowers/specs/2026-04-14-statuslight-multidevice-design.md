# StatusLight — Multi-Device Support Design

## Overview

Refactor the `internal/busylight` package into `internal/statuslight` with a driver-based architecture supporting multiple USB status light vendors. The current Kuando Busylight Omega implementation is correct and becomes one of four drivers.

## Interface

```go
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
```

Devices without keepalive requirements return a no-op from `KeepAlive()` and `false` from `NeedsKeepAlive()`. Devices without a buzzer ignore the sound parameters in `SetColorAndSound` and fall back to `SetColor`/`SetColorBlink`.

## Registry & Detection

```go
type DeviceID struct { VendorID, ProductID uint16 }
type DriverFactory func(handle uintptr, path, productName string) StatusLight

var registry = map[DeviceID]DriverFactory{}
func Register(id DeviceID, factory DriverFactory)
```

`Detect()` enumerates HID devices, looks up `(VendorID, ProductID)` in the registry, and returns the first match. Fallback: if no exact match, try VendorID-only match (covers future ProductIDs from known vendors).

`ListAllHIDDevices()`, `IsConnected()`, and `DiagnoseBusylight()` move to the statuslight package and become vendor-agnostic.

## Package Layout

```
internal/statuslight/
  statuslight.go          — Interface, registry, Detect(), types (USBDeviceInfo, DebugInfo)
  statuslight_stub.go     — Non-Windows stubs
  hid_windows.go          — Shared Windows HID: enumeration, open, WriteReport, SendFeatureReport, getDeviceCaps, helpers
  hid_stub.go             — Non-Windows HID stubs
  drivers/
    kuando.go             — Kuando Busylight Alpha/Omega
    luxafor.go            — Luxafor Flag
    blink1.go             — ThingM Blink(1)
    embrava.go            — Embrava Blynclight
```

## Driver Protocols

### Kuando Busylight (Alpha + Omega)

- **Write method**: WriteFile with 0x00 report ID prepended (65 bytes)
- **Report**: 64 bytes — 7 steps (8 bytes each) + footer with 16-bit checksum
- **Color scaling**: 0-255 input → 0-100 device scale
- **KeepAlive**: Required. Refresh every 10s, device timeout 15s. Opcode 0x8F.
- **SetColor**: Step 0 = Jump (opcode 0x10), repeat=0, R, G, B scaled, dutyCycleOn=0, dutyCycleOff=0, audio=0x00
- **SetColorBlink**: Same as SetColor but dutyCycleOn=4, dutyCycleOff=4
- **SetColorAndSound**: Audio byte = 0x80 | (ringtone << 3) | (volume & 0x07). Sound stopped after 2s via goroutine.
- **Footer**: bytes 56-63 = [0x00, 0x00, 0x00, 0x00, 0x0F, 0xFF, checksum_hi, checksum_lo]
- **Device IDs**:
  - Alpha: (0x04D8, 0xF848), (0x27BB, 0x3BCA), (0x27BB, 0x3BCB), (0x27BB, 0x3BCE)
  - Omega: (0x27BB, 0x3BCD), (0x27BB, 0x3BCF)

### Luxafor Flag

- **Write method**: WriteFile with 0x00 report ID prepended
- **Report**: 5-7 bytes, no checksum
- **Color scaling**: None (0-255 direct)
- **KeepAlive**: Not required
- **SetColor**: [0x01, 0xFF, R, G, B] (Command.Color=1, LEDS.All=0xFF)
- **SetColorBlink**: [0x03, 0xFF, R, G, B, 0x14, 0x00] (Command.Strobe=3, speed=20)
- **TurnOff**: [0x01, 0xFF, 0, 0, 0]
- **SetColorAndSound**: Falls back to SetColor/SetColorBlink (no buzzer hardware)
- **Device IDs**: (0x04D8, 0xF372)

### ThingM Blink(1)

- **Write method**: HidD_SetFeature (feature report, not WriteFile)
- **Report**: 8 bytes big-endian, no checksum
- **Color scaling**: None (0-255 direct)
- **KeepAlive**: Not required
- **SetColor**: [0x01, 0x6E, R, G, B, 0x00, 0x00, 0x00] (Report=1, Action=SetColor='n')
- **SetColorBlink**: Software blink via goroutine alternating color/off every 500ms
- **TurnOff**: [0x01, 0x6E, 0, 0, 0, 0, 0, 0] + cancel blink goroutine
- **SetColorAndSound**: Falls back to SetColor (no buzzer hardware)
- **Device ID**: (0x27B8, 0x01ED)

### Embrava Blynclight

- **Write method**: WriteFile with 0x00 report ID prepended
- **Report**: 9 bytes including report ID: [0x00, R, B, G, flags, audio, 0xFF, 0x22]. Note R-B-G order.
- **Color scaling**: None (0-255 direct)
- **KeepAlive**: Not required
- **Flags byte** (byte 4): bit 0=off, bit 1=dim, bit 2=flash, bits 3-5=speed (1=slow,2=med,4=fast)
- **SetColor**: [0x00, R, B, G, 0x00, 0x00, 0xFF, 0x22]
- **SetColorBlink**: [0x00, R, B, G, 0x04, 0x00, 0xFF, 0x22] (flash=1, speed=slow)
- **TurnOff**: [0x00, 0, 0, 0, 0x01, 0x00, 0xFF, 0x22] (off bit=1)
- **SetColorAndSound**: Falls back to SetColor/SetColorBlink (basic Blynclight has no buzzer)
- **Device IDs**: (0x2C0D, 0x0001), (0x2C0D, 0x000C), (0x0E53, 0x2516)

## Impact on app.go

- Import changes: `picklight/internal/busylight` → `picklight/internal/statuslight`
- Field `bl *busylight.Busylight` → `bl statuslight.StatusLight` (interface)
- `startKeepalive()` only runs if `bl.NeedsKeepAlive()` returns true
- `detectBusylight()` renamed `detectDevice()`
- Wails bindings renamed: `DiagnoseBusylight` → `DiagnoseDevice`, `DetectBusylight` → `DetectDevice`, `GetBusylightConnected` → `GetDeviceConnected`
- Status JSON fields: `busylightConnected` → `deviceConnected`, `busylightDevice` → `deviceName`

## Impact on Frontend

- Status field names update: `busylightConnected` → `deviceConnected`, `busylightDevice` → `deviceName`
- Wails binding calls renamed to match new Go method names
- Settings page "Busylight" section label → "Status Light"

## Testing

- Unit tests for each driver's report building functions (pure functions, no HID access)
- Existing `tray_assets_test.go` unaffected
- Manual testing with physical devices (only Kuando available; other drivers verified against reference implementation byte-for-byte)
