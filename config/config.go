package config

import (
	"github.com/jinzhu/configor"
)

// Config - Application configuration
type Config struct {
	SQLite struct {
		Path string `yaml:"path" default:"./sqlite.db" env:"SQLITE_PATH"`
	} `yaml:"sqlite"`
}

// LoadConfig - Load configuration file
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.New(&configor.Config{
		Debug:      false,
		Verbose:    false,
		Silent:     true,
		AutoReload: false,
	}).Load(cfg, path)
	return cfg, err
}
