package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"picklight/internal/applog"
	"picklight/internal/classifier"
	"picklight/internal/config"
	"picklight/internal/poller"
	"picklight/internal/statuslight"
	_ "picklight/internal/statuslight/drivers"
	"picklight/internal/updater"
)

type Status struct {
	OrdersPending    int               `json:"ordersPending"`
	ActiveThreshold  *config.Threshold `json:"activeThreshold"`
	LastPollTime     string            `json:"lastPollTime"`
	LastPollError    string            `json:"lastPollError,omitempty"`
	NextPollIn       int               `json:"nextPollIn"` // seconds until next poll
	DeviceConnected  bool              `json:"deviceConnected"`
	DeviceName       string            `json:"deviceName,omitempty"`
	Polling          bool              `json:"polling"`
}

type App struct {
	ctx context.Context

	configPath string
	cfg        config.Config
	appLog     *applog.Logger

	bl   statuslight.StatusLight
	blMu sync.Mutex

	status       Status
	statusMu     sync.Mutex
	lastPollAt   time.Time

	prevThresholdLabel string

	ticker     *time.Ticker
	tickerStop chan struct{}

	keepaliveTicker *time.Ticker
	keepaliveStop   chan struct{}
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.appLog = applog.New(200)

	dataDir := getDataDir()
	os.MkdirAll(dataDir, 0755)

	a.configPath = filepath.Join(dataDir, "config.yaml")

	cfg, err := config.Load(a.configPath)
	if err != nil {
		a.appLog.Warn("Erreur chargement config: %v", err)
	}
	a.cfg = cfg

	a.detectDevice()

	a.initTray()

	if a.cfg.EndpointURL != "" {
		a.startTicker()
	}

	a.appLog.Info("PickLight démarré")
}

func (a *App) startKeepalive() {
	a.stopKeepalive()
	a.keepaliveTicker = time.NewTicker(10 * time.Second)
	a.keepaliveStop = make(chan struct{})
	go func() {
		for {
			select {
			case <-a.keepaliveTicker.C:
				a.blMu.Lock()
				if a.bl != nil {
					a.bl.KeepAlive()
				}
				a.blMu.Unlock()
			case <-a.keepaliveStop:
				return
			}
		}
	}()
}

func (a *App) stopKeepalive() {
	if a.keepaliveTicker != nil {
		a.keepaliveTicker.Stop()
	}
	if a.keepaliveStop != nil {
		select {
		case <-a.keepaliveStop:
		default:
			close(a.keepaliveStop)
		}
	}
}

func (a *App) shutdown(ctx context.Context) {
	a.stopTicker()
	a.stopKeepalive()
	a.cleanupTray()
	a.blMu.Lock()
	if a.bl != nil {
		a.bl.TurnOff()
		a.bl.Close()
	}
	a.blMu.Unlock()
}

func (a *App) detectDevice() {
	a.blMu.Lock()
	defer a.blMu.Unlock()
	if a.bl != nil {
		a.bl.Close()
		a.bl = nil
	}
	bl, err := statuslight.Detect()
	if err != nil {
		a.appLog.Warn("Status light non détecté: %v", err)
		a.statusMu.Lock()
		a.status.DeviceConnected = false
		a.status.DeviceName = ""
		a.statusMu.Unlock()
		return
	}
	a.bl = bl
	a.statusMu.Lock()
	a.status.DeviceConnected = true
	a.status.DeviceName = bl.DeviceName()
	a.statusMu.Unlock()
	a.appLog.Info("Status light connecté: %s", bl.DeviceName())
}

func (a *App) startTicker() {
	a.stopTicker()
	interval := time.Duration(a.cfg.PollIntervalSeconds) * time.Second
	a.ticker = time.NewTicker(interval)
	a.tickerStop = make(chan struct{})
	a.statusMu.Lock()
	a.status.Polling = true
	a.statusMu.Unlock()

	go func() {
		a.doPoll()
		for {
			select {
			case <-a.ticker.C:
				a.doPoll()
			case <-a.tickerStop:
				return
			}
		}
	}()
}

func (a *App) stopTicker() {
	if a.ticker != nil {
		a.ticker.Stop()
	}
	if a.tickerStop != nil {
		select {
		case <-a.tickerStop:
		default:
			close(a.tickerStop)
		}
	}
	a.statusMu.Lock()
	a.status.Polling = false
	a.statusMu.Unlock()
}

func (a *App) doPoll() {
	// Retry up to 3 times on error
	var value int
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		value, err = poller.Poll(a.cfg.EndpointURL, a.cfg.JSONPath, a.cfg.TLSSkipVerify)
		if err == nil {
			break
		}
		if attempt < 2 {
			a.appLog.Warn("Erreur polling (tentative %d/3): %v", attempt+1, err)
			time.Sleep(2 * time.Second)
		}
	}

	a.statusMu.Lock()
	a.lastPollAt = time.Now()
	a.status.LastPollTime = a.lastPollAt.Format("15:04:05")
	if err != nil {
		a.status.LastPollError = err.Error()
		a.statusMu.Unlock()
		a.appLog.Error("Erreur polling après 3 tentatives: %v", err)
		runtime.EventsEmit(a.ctx, "poll:error", err.Error())
		a.updateTrayTooltip()
		return
	}
	a.status.LastPollError = ""
	a.status.OrdersPending = value
	a.statusMu.Unlock()

	a.appLog.Info("Poll: %d commande(s) en attente", value)

	result := classifier.Classify(value, a.cfg.Thresholds)

	a.statusMu.Lock()
	if result != nil {
		t := result.Threshold
		a.status.ActiveThreshold = &t
	} else {
		a.status.ActiveThreshold = nil
	}
	a.statusMu.Unlock()

	// Auto-reconnect device
	a.blMu.Lock()
	if a.bl == nil {
		if statuslight.IsConnected() {
			bl, blErr := statuslight.Detect()
			if blErr == nil {
				a.bl = bl
				a.statusMu.Lock()
				a.status.DeviceConnected = true
				a.status.DeviceName = bl.DeviceName()
				a.statusMu.Unlock()
				a.appLog.Info("Status light reconnecté: %s", bl.DeviceName())
			}
		}
		if a.bl == nil {
			a.blMu.Unlock()
			a.updateTrayTooltip()
			runtime.EventsEmit(a.ctx, "status:updated")
			return
		}
	}

	// Try to set color, detect disconnection
	var setErr error
	if result == nil {
		setErr = a.bl.TurnOff()
		a.prevThresholdLabel = ""
	} else {
		thresholdChanged := result.Threshold.Label != a.prevThresholdLabel
		playSound := a.cfg.SoundEnabled && result.Threshold.Sound && (!a.cfg.SoundOnChangeOnly || thresholdChanged)

		if playSound {
			a.appLog.Info("Seuil changé → son activé (%s)", result.Threshold.Label)
			setErr = a.bl.SetColorAndSound(result.R, result.G, result.B, 3, 4, result.Threshold.Blink)
		} else if result.Threshold.Blink {
			setErr = a.bl.SetColorBlink(result.R, result.G, result.B)
		} else {
			setErr = a.bl.SetColor(result.R, result.G, result.B)
		}
		if a.prevThresholdLabel == "" && a.bl.NeedsKeepAlive() {
			a.startKeepalive()
		}
		a.prevThresholdLabel = result.Threshold.Label
	}

	// If write failed, device likely disconnected
	if setErr != nil {
		a.appLog.Warn("Status light déconnecté: %v", setErr)
		a.bl.Close()
		a.bl = nil
		a.statusMu.Lock()
		a.status.DeviceConnected = false
		a.status.DeviceName = ""
		a.statusMu.Unlock()
	}
	a.blMu.Unlock()

	a.updateTrayTooltip()
	runtime.EventsEmit(a.ctx, "status:updated")
}

// ─── Wails Bindings ──────────────────────────────────────────────────────────

func (a *App) GetStatus() Status {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	// Compute NextPollIn
	if a.status.Polling && !a.lastPollAt.IsZero() {
		elapsed := time.Since(a.lastPollAt)
		interval := time.Duration(a.cfg.PollIntervalSeconds) * time.Second
		remaining := interval - elapsed
		if remaining < 0 {
			remaining = 0
		}
		a.status.NextPollIn = int(remaining.Seconds())
	} else {
		a.status.NextPollIn = 0
	}
	return a.status
}

func (a *App) PollNow() error {
	if a.cfg.EndpointURL == "" {
		return fmt.Errorf("aucune URL configurée")
	}
	go a.doPoll()
	return nil
}

func (a *App) GetConfig() config.Config {
	return a.cfg
}

func (a *App) SaveConfig(cfg config.Config) error {
	// Validation
	if cfg.EndpointURL != "" && len(cfg.EndpointURL) < 8 {
		return fmt.Errorf("URL invalide")
	}
	if cfg.PollIntervalSeconds < 10 {
		return fmt.Errorf("l'intervalle de polling doit être d'au moins 10 secondes")
	}
	if cfg.JSONPath == "" {
		return fmt.Errorf("le chemin JSON ne peut pas être vide")
	}
	for i, t := range cfg.Thresholds {
		if t.Min > t.Max {
			return fmt.Errorf("seuil #%d: min (%d) > max (%d)", i+1, t.Min, t.Max)
		}
		if t.Color == "" {
			return fmt.Errorf("seuil #%d: couleur requise", i+1)
		}
		if _, _, _, err := classifier.ParseHexColor(t.Color); err != nil {
			return fmt.Errorf("seuil #%d: %v", i+1, err)
		}
	}

	if err := config.Save(a.configPath, cfg); err != nil {
		return fmt.Errorf("erreur sauvegarde: %w", err)
	}
	oldInterval := a.cfg.PollIntervalSeconds
	oldURL := a.cfg.EndpointURL
	a.cfg = cfg

	if cfg.EndpointURL != "" && (cfg.EndpointURL != oldURL || cfg.PollIntervalSeconds != oldInterval) {
		a.startTicker()
	} else if cfg.EndpointURL == "" {
		a.stopTicker()
	}

	a.appLog.Info("Configuration sauvegardée")
	return nil
}

func (a *App) GetDeviceConnected() bool {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	return a.status.DeviceConnected
}

func (a *App) ListUSBDevices() []statuslight.USBDeviceInfo {
	return statuslight.ListAllHIDDevices()
}

func (a *App) DiagnoseDevice() []statuslight.DebugInfo {
	a.appLog.Info("Diagnostic status light lancé...")
	results := statuslight.DiagnoseDevices()
	for _, r := range results {
		a.appLog.Info("Interface: %s | open=%v | output=%d | feature=%d | usage=0x%X page=0x%X",
			r.Path, r.OpenOK, r.OutputReportByteLength, r.FeatureReportByteLength, r.Usage, r.UsagePage)
		a.appLog.Info("  WriteFile: %s", r.WriteFileResult)
		a.appLog.Info("  SetOutputReport: %s", r.SetOutputReportResult)
		a.appLog.Info("  SetFeature: %s", r.SetFeatureResult)
	}
	return results
}

func (a *App) DetectDevice() string {
	a.detectDevice()
	a.statusMu.Lock()
	defer a.statusMu.Unlock()
	if a.status.DeviceConnected {
		return a.status.DeviceName
	}
	return ""
}

func (a *App) TestColor(color string) error {
	r, g, b, err := classifier.ParseHexColor(color)
	if err != nil {
		return err
	}
	a.blMu.Lock()
	defer a.blMu.Unlock()
	if a.bl == nil {
		return fmt.Errorf("status light non connecté")
	}
	a.appLog.Info("Test couleur: %s", color)
	return a.bl.SetColor(r, g, b)
}

func (a *App) TestThreshold(color string, sound bool, blink bool) error {
	r, g, b, err := classifier.ParseHexColor(color)
	if err != nil {
		return err
	}
	a.blMu.Lock()
	defer a.blMu.Unlock()
	if a.bl == nil {
		return fmt.Errorf("status light non connecté")
	}
	if sound {
		a.appLog.Info("Test seuil: %s + son + blink=%v", color, blink)
		return a.bl.SetColorAndSound(r, g, b, 3, 4, blink)
	}
	if blink {
		a.appLog.Info("Test seuil: %s + blink", color)
		return a.bl.SetColorBlink(r, g, b)
	}
	a.appLog.Info("Test seuil: %s", color)
	return a.bl.SetColor(r, g, b)
}

func (a *App) TestSound(tone int) error {
	a.blMu.Lock()
	defer a.blMu.Unlock()
	if a.bl == nil {
		return fmt.Errorf("status light non connecté")
	}
	a.appLog.Info("Test son: tone=%d", tone)
	return a.bl.SetColorAndSound(0, 0, 0xFF, byte(tone), 4, false)
}

func (a *App) TestOff() error {
	a.blMu.Lock()
	defer a.blMu.Unlock()
	if a.bl == nil {
		return fmt.Errorf("status light non connecté")
	}
	a.appLog.Info("Test: éteint")
	return a.bl.TurnOff()
}

func (a *App) GetVersion() string {
	return Version
}

// GetLanguage returns the effective language ("fr" or "en").
// If config is "auto", detects from OS.
func (a *App) GetLanguage() string {
	lang := a.cfg.Language
	if lang == "" || lang == "auto" {
		lang = detectOSLanguage()
	}
	if lang != "fr" && lang != "en" {
		lang = "fr"
	}
	return lang
}

// CheckForUpdate checks GitHub for a newer version of PickLight.
func (a *App) CheckForUpdate() updater.UpdateStatus {
	return updater.Check(Version)
}

// ApplyUpdate downloads and installs the latest version, then restarts the app.
func (a *App) ApplyUpdate() error {
	a.appLog.Info("Checking for update...")
	status := updater.Check(Version)
	if !status.UpdateAvail {
		return fmt.Errorf("already up to date")
	}
	if status.DownloadURL == "" {
		return fmt.Errorf("no download URL found for this platform")
	}
	a.appLog.Info("Downloading update v%s...", status.LatestVersion)
	if err := applyUpdate(status); err != nil {
		a.appLog.Error("Update failed: %v", err)
		return err
	}
	a.appLog.Info("Update downloaded, restarting...")
	go func() {
		time.Sleep(500 * time.Millisecond)
		a.QuitApp()
	}()
	return nil
}

func (a *App) GetLogs() []applog.LogEntry {
	return a.appLog.GetEntries()
}

func (a *App) ClearLogs() {
	a.appLog.Clear()
}

func (a *App) IsStartupEnabled() bool {
	return isStartupEnabled()
}

func (a *App) SetStartupEnabled(enabled bool) error {
	return setStartupEnabled(enabled)
}

func (a *App) ShowWindow() {
	runtime.Show(a.ctx)
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
}

func (a *App) QuitApp() {
	a.stopKeepalive()
	a.blMu.Lock()
	if a.bl != nil {
		a.bl.TurnOff()
		a.bl.Close()
		a.bl = nil
	}
	a.blMu.Unlock()
	a.cleanupTray()
	runtime.Quit(a.ctx)
}

func (a *App) PickDirectory(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
}

// updateTrayTooltip updates the systray tooltip with current status.
func (a *App) updateTrayTooltip() {
	a.statusMu.Lock()
	pending := a.status.OrdersPending
	errMsg := a.status.LastPollError
	a.statusMu.Unlock()

	tip := fmt.Sprintf("PickLight — %d commande(s)", pending)
	if errMsg != "" {
		tip = "PickLight — Erreur polling"
	}
	updateTrayTip(tip)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
