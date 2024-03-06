package config

import (
	"fmt"
	"os"

	"github.com/pynezz/bivrost/internal/fs"
	"github.com/pynezz/bivrost/internal/util"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Sources []struct {
		Name     string   `yaml:"name"`
		Type     string   `yaml:"type"`
		Location string   `yaml:"location"`
		Format   string   `yaml:"format"`
		Tags     []string `yaml:"tags"`
	} `yaml:"sources"`
	Network []struct {
		ReadTimeout  int `yaml:"read_timeout,omitempty"`
		WriteTimeout int `yaml:"write_timeout,omitempty"`
	} `yaml:"network"`
}

// LoadConfig loads the configuration from the given path
func LoadConfig(path string) (*Config, error) {

	// First, let's check if the file exists via the util function fs.GetFile
	file, err := fs.GetFile(path)

	// If the file does not exist, we should return an error
	if err != nil {
		util.PrintErrorf("Failed to load configuration file: %s", path)
		return nil, err
	}

	// Defer the file close, so it's closed after the function returns
	defer file.Close()

	// Buf is a byte slice, which will be used to pass the file contents to the yaml.Unmarshal function
	buf, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(buf, &cfg) // From buf, to &cfg. &cfg is a pointer to the Config struct memory address
	if err != nil {
		return nil, err
	}

	util.PrintSuccess(fmt.Sprintf("Loaded configuration file: %s at %s", file.Name(), path))
	return &cfg, nil
}
