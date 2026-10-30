package app

import (
	"encoding/json"
	"os"
)

type AppConfig struct {
	Mode        string `json:"mode,omitempty"`
	Port        int    `json:"port"`
	Key         string `json:"key,omitempty"`
	LogLevel    string `json:"log_level,omitempty"`
	LogDir      string `json:"log_dir,omitempty"`
	AppDeepLink string `json:"app_deep_link,omitempty"`
}

func InitConfig(path string) (AppConfig, error) {
	var conf AppConfig

	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}

	if err := json.Unmarshal(data, &conf); err != nil {
		return AppConfig{}, err
	}

	if conf.Mode == "" {
		conf.Mode = "debug"
	}
	if conf.LogLevel == "" {
		conf.LogLevel = "info"
	}
	if conf.LogDir == "" {
		conf.LogDir = "./logs"
	}

	return conf, nil
}
