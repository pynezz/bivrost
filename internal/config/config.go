package config

import (
	"fmt"

	"github.com/pynezz/bivrost/internal/fs"
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
	file, err := fs.GetFile(path)

	if err != nil {
		util.PrintErrorf("Failed to load configuration file: %s", path)
		return nil, err
	}

	defer file.Close()

	util.PrintSuccess(fmt.Sprintf("Loaded configuration file: %s at %s", file.Name(), path))
	return &Config{}, nil
}
