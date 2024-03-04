package config

import (
	"github.com/pynezz/bivrost/internal/util"
)

// Config is the main struct for the configuration
type Config struct {
	Sources []struct {
		Name     string   `json:"name"`
		Type     string   `json:"type"`
		Location string   `json:"location"`
		Format   string   `json:"format"`
		Tags     []string `json:"tags"`
	} `json:"sources"`
}

// LoadConfig loads the configuration from the given path
func LoadConfig(path string) (*Config, error) {
	util.PrintSuccess("Loaded configuration from " + path)
	return &Config{}, nil
}
