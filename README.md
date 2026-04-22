# PickLight

Voyant USB de suivi des commandes e-commerce. PickLight interroge periodiquement un endpoint JSON (ex: PrestaShop), lit le nombre de commandes en attente de preparation, et pilote un voyant USB pour signaler visuellement l'etat : vert quand tout est traite, orange quand quelques commandes attendent, rouge clignotant quand la file deborde. L'equipe logistique sait en un coup d'oeil s'il y a du travail a lancer, sans ouvrir le back-office.

## Features

- **Multi-device support** — Kuando Busylight (Alpha/Omega), Luxafor Flag, ThingM Blink(1), Embrava Blynclight
- **JSON polling** — monitors any HTTP endpoint returning a numeric value via configurable JSON path
- **Threshold-based colors** — define min/max ranges with colors, blink patterns, and optional sound
- **Auto-detection** — finds and connects to any supported USB status light automatically
- **Auto-reconnect** — recovers when the device is unplugged and reconnected
- **KeepAlive management** — automatic keepalive for devices that require it (Kuando)
- **System tray** — runs quietly in the background with status tooltip
- **Single instance** — launching a second instance brings the existing window to focus
- **Startup** — can be configured to start with Windows
- **Self-installing** — installs to Program Files with desktop shortcut on first run

## Architecture

```
Ticker (poll_interval_seconds)
  --> HTTP GET endpoint_url
  --> Parse JSON --> extract value at json_path
  --> Classifier: match value against thresholds --> RGB color + sound flag
  --> USB HID: send report to status light via driver
  --> Update UI status via Wails event
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go |
| UI | [Wails v2](https://wails.io) + Vanilla HTML/CSS/JS |
| USB HID | Windows SetupAPI + HID.dll (direct syscall, no CGo) |
| Config | YAML |
| HTTP client | net/http (stdlib) |

## Supported Devices

| Device | Vendor ID | Protocol |
|--------|-----------|----------|
| Kuando Busylight Alpha | 0x04D8 / 0x27BB | 64-byte report, keepalive required |
| Kuando Busylight Omega | 0x27BB | 64-byte report, keepalive required |
| Luxafor Flag | 0x04D8 | 5-7 byte simple command |
| ThingM Blink(1) | 0x27B8 | 8-byte feature report |
| Embrava Blynclight | 0x2C0D / 0x0E53 | 9-byte report (R-B-G order) |

Device protocols were implemented using [JnyJny/busylight](https://github.com/JnyJny/busylight) (Apache 2.0) as a reference.

## Requirements

- Windows 10 or 11
- A supported USB status light device

## Installation

Download the latest `picklight.exe` from the [Releases](https://github.com/ecommercerie/picklight/releases) page and run it. On first launch, it will offer to install itself to `C:\Program Files\PickLight` with a desktop shortcut.

## Configuration

On first launch, open **Configuration** and set:

1. **Endpoint URL** — the JSON API to poll (e.g. your PrestaShop stats exporter)
2. **JSON path** — dot-separated path to the numeric value (e.g. `stats.orders_pending`)
3. **Poll interval** — how often to check (in seconds, minimum 10)
4. **Thresholds** — define color ranges:
   - `0-0` = green (no orders)
   - `1-5` = orange (a few orders)
   - `6+` = red (many orders)

Each threshold can have a color, blink pattern, sound alert, and label.

## Building from source

Prerequisites:
- [Go 1.21+](https://go.dev/dl/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Development mode (hot reload)
wails dev

# Production build
wails build -platform windows/amd64

# Production build with version
wails build -platform windows/amd64 -ldflags "-X main.Version=1.0.0"
```

The output binary is at `build/bin/picklight.exe`.

## Project Structure

```
picklight/
  main.go                        # Wails bootstrap, single instance, self-install
  app.go                         # App struct, all Wails bindings, ticker loop
  version.go                     # Version variable (injected via ldflags)
  tray_windows.go                # System tray (Win32 API)
  install_windows.go             # Self-installer (Program Files + shortcuts)
  internal/
    statuslight/
      statuslight.go             # StatusLight interface, registry, detection
      hid_windows.go             # Shared Windows HID enumeration & I/O
      drivers/
        kuando.go                # Kuando Busylight Alpha/Omega
        luxafor.go               # Luxafor Flag
        blink1.go                # ThingM Blink(1)
        embrava.go               # Embrava Blynclight
    config/                      # YAML config load/save
    poller/                      # HTTP GET + JSON path extraction
    classifier/                  # Threshold matching -> color + sound
    applog/                      # In-memory ring buffer for UI logs
  frontend/
    index.html                   # SPA shell with hash router
    pages/
      status.js                  # Live status display
      settings.js                # Configuration editor + debug tools
      about.js                   # About page with credits
```

## Co-authored

This project was co-authored with [Claude Code](https://claude.ai/claude-code) by Anthropic.

## License

MIT
