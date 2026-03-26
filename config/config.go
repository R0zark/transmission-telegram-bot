package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config struct to hold bot and Transmission configuration
type Transmission struct {
	URL             string `yaml:"url"`
	Port            uint16 `yaml:"port"`
	HTTPS           bool   `yaml:"https"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	DefaultLocation string `yaml:"defaultLocation"`
}

type Solarman struct {
	AppId     string `yaml:"appId"`
	AppSecret string `yaml:"appSecret"`
	Email     string `yaml:"email"`
	Password  string `yaml:"password"`
}

type API struct {
	AuthURL string `yaml:"authURL"`
	ApiURL  string `yaml:"apiURL"`
}

type Telegram struct {
	BotToken string `yaml:"botToken"`
	ChatID   string `yaml:"chatID"`
}

type Device struct {
	DeviceSn string `yaml:"deviceSn"`
}
type Config struct {
	Transmission Transmission `yaml:"transmission"`
	Telegram     Telegram     `yaml:"telegram"`
	Device       Device       `yaml:"device"`
}

// ReadConfig loads configuration from a YAML file
func ReadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Transmission.Port == 0 {
		cfg.Transmission.Port = 9091
	}

	return &cfg, nil
}
