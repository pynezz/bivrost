package config

import (
	"fmt"
	"os"

	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/util"

	"gopkg.in/yaml.v3"
)

type Sources struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
	Config      string   `yaml:"config"`
	Format      string   `yaml:"format,omitempty"`
	Tags        []string `yaml:"tags"`
}

type Cfg struct {
	Sources []Sources `yaml:"sources"`
	Network struct {
		ReadTimeout  int `yaml:"read_timeout,omitempty"`
		WriteTimeout int `yaml:"write_timeout,omitempty"`
		Port         int `yaml:"port,omitempty"`
	} `yaml:"network"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"users_database"`
}

// LoadConfig loads the configuration from the given path
func LoadConfig(path string) (*Cfg, error) {

	// First, let's check if the file exists via the util function fs.GetFile
	file, err := fsutil.GetFile(path)
	util.PrintDebug(fmt.Sprintf("Loading configuration file: %s...", path))

	// If the file does not exist, we should return an error
	if err != nil {
		util.PrintErrorf("Failed to load configuration file: %s", path)
		return nil, err
	}

	// Defer the file close, so it's closed after the function returns
	defer file.Close()

	// Buf is a byte slice, which will be used to
	// pass the file contents to the yaml.Unmarshal function
	buf, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}

	var cfg Cfg
	// From buf, to &cfg. &cfg is a pointer to the Config struct memory address
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, err
	}

	util.PrintSuccess(fmt.Sprintf(
		"Loaded configuration file: %s at %s", file.Name(), path))
	return &cfg, nil
}

func WriteConfig(cfg *Cfg, path string) error {
	// Open the file for writing
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Marshal the config to YAML
	buf, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Write the buffer to the file
	_, err = file.Write(buf)
	if err != nil {
		return err
	}

	return nil
}
