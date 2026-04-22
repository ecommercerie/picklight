package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Threshold struct {
	Min   int    `yaml:"min" json:"min"`
	Max   int    `yaml:"max" json:"max"`
	Color string `yaml:"color" json:"color"`
	Sound bool   `yaml:"sound" json:"sound"`
	Blink bool   `yaml:"blink" json:"blink"`
	Label string `yaml:"label" json:"label"`
}

type Config struct {
	EndpointURL         string      `yaml:"endpoint_url" json:"endpointUrl"`
	PollIntervalSeconds int         `yaml:"poll_interval_seconds" json:"pollIntervalSeconds"`
	JSONPath            string      `yaml:"json_path" json:"jsonPath"`
	TLSSkipVerify       bool        `yaml:"tls_skip_verify" json:"tlsSkipVerify"`
	Thresholds          []Threshold `yaml:"thresholds" json:"thresholds"`
	SoundEnabled        bool        `yaml:"sound_enabled" json:"soundEnabled"`
	SoundOnChangeOnly   bool        `yaml:"sound_on_change_only" json:"soundOnChangeOnly"`
	Language            string      `yaml:"language" json:"language"` // "auto", "fr", "en"
}

func DefaultThresholds() []Threshold {
	return []Threshold{
		{Min: 0, Max: 0, Color: "#00FF00", Sound: false, Label: "Aucune commande"},
		{Min: 1, Max: 5, Color: "#FFAA00", Sound: true, Label: "Quelques commandes"},
		{Min: 6, Max: 999, Color: "#FF0000", Sound: true, Label: "Beaucoup de commandes"},
	}
}

func Load(path string) (Config, error) {
	cfg := Config{
		PollIntervalSeconds: 300,
		JSONPath:            "stats.orders_pending",
		SoundOnChangeOnly:   true,
		Language:            "auto",
		Thresholds:          DefaultThresholds(),
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.PollIntervalSeconds <= 0 {
		cfg.PollIntervalSeconds = 300
	}
	if cfg.JSONPath == "" {
		cfg.JSONPath = "stats.orders_pending"
	}
	if len(cfg.Thresholds) == 0 {
		cfg.Thresholds = DefaultThresholds()
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
