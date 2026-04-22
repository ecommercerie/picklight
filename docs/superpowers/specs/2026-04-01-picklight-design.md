# PickLight — Design Specification

## Overview

PickLight is a Windows desktop application written in Go that polls a PrestaShop JSON endpoint at a configurable interval, reads the `orders_pending` value, and drives a Kuando Busylight via USB HID to display a color based on user-defined thresholds. Optional buzzer sound on threshold changes. Runs as a system tray application with a Wails-based settings window.

Target: Windows 10/11 only. Single binary.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go |
| UI framework | Wails v2 |
| Frontend | Vanilla HTML/CSS/JS |
| USB HID | github.com/karalabe/hid |
| Config | YAML (gopkg.in/yaml.v3) |
| HTTP client | net/http (stdlib) |

## Data Flow

```
Ticker (poll_interval_seconds)
  → HTTP GET endpoint_url
  → Parse JSON → extract stats.orders_pending
  → Classifier: match value against thresholds → RGB color + sound flag
  → USB HID: send 64-byte feature report to Busylight
  → Update UI status via Wails event
```

## Project Structure

```
picklight/
├── main.go                     # Wails bootstrap
├── app.go                      # App struct, Wails bindings, ticker loop
├── tray_windows.go             # Win32 systray (same pattern as PrintDock)
├── tray_stub.go                # Stub for non-Windows dev
├── tray_assets.go              # Embedded tray icon
├── startup_windows.go          # Auto-start registry
├── startup_stub.go             # Stub for non-Windows dev
├── singleinstance_windows.go   # Named mutex single instance
├── singleinstance_stub.go      # Stub
├── internal/
│   ├── config/
│   │   └── config.go           # Load/save config.yaml
│   ├── poller/
│   │   └── poller.go           # HTTP GET + JSON parsing
│   ├── classifier/
│   │   └── classifier.go       # Threshold matching → color + sound
│   ├── busylight/
│   │   └── busylight.go        # USB HID device control
│   └── applog/
│       └── applog.go           # Ring buffer logger (reuse from PrintDock)
├── frontend/
│   ├── index.html              # SPA shell
│   ├── style.css
│   ├── app.js                  # Router
│   └── pages/
│       ├── status.js           # Live status display
│       └── settings.js         # Configuration editor
├── config.yaml
└── wails.json
```

## Data Models

### config.yaml

```yaml
endpoint_url: "https://em.app.localhost/module/ps_stats_exporter/api?token=e7564b8b9091a2f1017df37aba2edc8c"
poll_interval_seconds: 300
json_path: "stats.orders_pending"
tls_skip_verify: true

thresholds:
  - min: 0
    max: 0
    color: "#00FF00"
    sound: false
    label: "Aucune commande"
  - min: 1
    max: 5
    color: "#FFAA00"
    sound: true
    label: "Quelques commandes"
  - min: 6
    max: 999
    color: "#FF0000"
    sound: true
    label: "Beaucoup de commandes"

sound_enabled: true
sound_on_change_only: true
```

### Go structs

```go
type Config struct {
    EndpointURL        string      `yaml:"endpoint_url" json:"endpointUrl"`
    PollIntervalSeconds int        `yaml:"poll_interval_seconds" json:"pollIntervalSeconds"`
    JSONPath           string      `yaml:"json_path" json:"jsonPath"`
    TLSSkipVerify      bool        `yaml:"tls_skip_verify" json:"tlsSkipVerify"`
    Thresholds         []Threshold `yaml:"thresholds" json:"thresholds"`
    SoundEnabled       bool        `yaml:"sound_enabled" json:"soundEnabled"`
    SoundOnChangeOnly  bool        `yaml:"sound_on_change_only" json:"soundOnChangeOnly"`
}

type Threshold struct {
    Min   int    `yaml:"min" json:"min"`
    Max   int    `yaml:"max" json:"max"`
    Color string `yaml:"color" json:"color"`   // "#RRGGBB"
    Sound bool   `yaml:"sound" json:"sound"`
    Label string `yaml:"label" json:"label"`
}
```

## Module Details

### Config

- `Load(path string) (Config, error)` — parse YAML, apply defaults (interval=300, json_path="stats.orders_pending")
- `Save(path string, cfg Config) error` — write YAML
- Default thresholds if none defined: 0=green, 1-5=orange, 6+=red

### Poller

- `Poll(url string, jsonPath string, tlsSkipVerify bool) (int, error)`
- Makes HTTP GET request with configurable TLS verification
- Parses JSON response
- Extracts value at `jsonPath` using dot notation (e.g. "stats.orders_pending")
- Supports nested paths by splitting on `.` and walking the JSON map
- Returns integer value
- Timeout: 30 seconds

### Classifier

- `Classify(value int, thresholds []Threshold) *Threshold`
- Iterates thresholds in order
- Returns first threshold where `min <= value <= max`
- Returns nil if no match (light off)

### Busylight

- `type Busylight struct` — holds HID device handle
- `Detect() (*Busylight, error)` — scans USB HID devices for Vendor ID `0x27BB` (Plenom)
- `SetColor(r, g, b byte) error` — sends 64-byte feature report:
  - Byte 0: `0x11` (keepalive + step command)
  - Byte 1: red
  - Byte 2: green
  - Byte 3: blue
  - Byte 4: `0xFF` (on-time, always on)
  - Byte 5: `0x00` (off-time, no blink)
  - Bytes 6-63: `0x00`
- `SetColorAndSound(r, g, b, tone, volume byte) error` — same but:
  - Byte 6: tone (1-15 for different sounds)
  - Byte 7: volume (0-7)
- `TurnOff() error` — sends all zeros
- `Close() error` — releases HID device
- `IsConnected() bool` — checks if device is still present
- Color parsing helper: `ParseHexColor(hex string) (r, g, b byte, error)`

### AppLog

- Same ring buffer logger as PrintDock (`internal/applog/`)
- 200 entries max

## App (app.go)

### Wails Bindings

```go
// Status
func (a *App) GetStatus() Status
func (a *App) PollNow() error              // manual trigger

// Config
func (a *App) GetConfig() Config
func (a *App) SaveConfig(cfg Config) error

// Busylight
func (a *App) GetBusylightStatus() BusylightStatus
func (a *App) TestColor(color string) error   // preview a color on the light
func (a *App) TestSound(tone int) error       // play a sound on the Busylight
func (a *App) TestOff() error                 // turn off the Busylight

// Logs
func (a *App) GetLogs() []LogEntry
func (a *App) ClearLogs()

// System
func (a *App) IsStartupEnabled() bool
func (a *App) SetStartupEnabled(enabled bool) error
func (a *App) QuitApp()
func (a *App) ShowWindow()
```

### Status struct

```go
type Status struct {
    OrdersPending    int       `json:"ordersPending"`
    ActiveThreshold  *Threshold `json:"activeThreshold"`
    LastPollTime     time.Time `json:"lastPollTime"`
    LastPollError    string    `json:"lastPollError,omitempty"`
    BusylightConnected bool   `json:"busylightConnected"`
    Polling          bool      `json:"polling"`
}
```

### Events (Go → JS)

- `status:updated` — after each poll, sends Status
- `busylight:connected` / `busylight:disconnected`
- `poll:error` — on HTTP or parse error

### Ticker Loop

```
startup:
  1. Load config
  2. Detect Busylight (log warning if not found, retry every 30s)
  3. Start ticker (poll_interval_seconds)

each tick:
  1. HTTP GET endpoint_url
  2. Parse JSON, extract value at json_path
  3. Classify value → threshold
  4. If threshold changed OR !sound_on_change_only:
     - Set color (+ sound if threshold.sound && sound_enabled)
  5. If threshold unchanged:
     - Set color only (refresh keepalive)
  6. Emit "status:updated" event
  7. Log result

on config change:
  - Restart ticker with new interval
  - Re-detect Busylight if needed
```

## Frontend Pages

### Status (`#status`)
- Big colored circle showing current light color
- `orders_pending` value in large text
- Active threshold label
- Last poll time + next poll countdown
- Busylight connected/disconnected indicator
- "Interroger maintenant" button
- Auto-refresh via Wails events

### Settings (`#settings`)
- Endpoint URL input
- Poll interval (seconds)
- JSON path input
- TLS skip verify checkbox
- Sound enabled checkbox
- Sound on change only checkbox
- Thresholds table (editable): min, max, color picker, sound checkbox, label
  - Add/remove threshold rows
- "Tester la couleur" button per threshold row
- Debug section (collapsible): "Tester couleur" (color picker + send), "Tester son" (tone selector + send), "Éteindre" — for verifying USB HID connectivity
- Auto-start Windows checkbox
- Save button
- Logs section at the bottom (collapsible)

## Systray

- Same Win32 native systray as PrintDock
- Icon changes color to match current threshold (generate colored square icons dynamically)
- Tooltip: "PickLight — X commandes en attente"
- Menu: "Ouvrir", separator, "Quitter"
- Window hidden on close, not destroyed

## Error Handling

| Error | Action |
|-------|--------|
| HTTP request fails | Log, keep last known state, emit poll:error |
| JSON parse error | Log, keep last known state |
| Busylight disconnected | Log, set status disconnected, retry detection every 30s |
| Invalid threshold config | Log warning, skip invalid entries |
| Config file missing | Use defaults |

## Build

```bash
# Development
wails dev

# Production Windows binary
wails build -platform windows/amd64
```
