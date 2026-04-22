package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PollIntervalSeconds != 300 {
		t.Errorf("expected 300, got %d", cfg.PollIntervalSeconds)
	}
	if cfg.JSONPath != "stats.orders_pending" {
		t.Errorf("expected stats.orders_pending, got %s", cfg.JSONPath)
	}
	if len(cfg.Thresholds) != 3 {
		t.Errorf("expected 3 default thresholds, got %d", len(cfg.Thresholds))
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("endpoint_url: http://test\npoll_interval_seconds: 60\njson_path: data.count\n"), 0644)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EndpointURL != "http://test" {
		t.Errorf("expected http://test, got %s", cfg.EndpointURL)
	}
	if cfg.PollIntervalSeconds != 60 {
		t.Errorf("expected 60, got %d", cfg.PollIntervalSeconds)
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := Config{EndpointURL: "http://save-test", PollIntervalSeconds: 120, JSONPath: "x.y", Thresholds: DefaultThresholds()}
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.EndpointURL != "http://save-test" {
		t.Errorf("expected http://save-test, got %s", loaded.EndpointURL)
	}
}
