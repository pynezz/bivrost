package model

// This is the pkg/model/config.go file, which contains the configuration model for Bivrost.
// The reason for the existence of this file is to define the configuration model, which is used to load the configuration from the YAML file.
// The placement pkg/model/ is a common pattern for Go projects, which defines the model for the application.

// Config represents the top-level configuration structure for Bivrost.
type Config struct {
	Sources  []Source       `yaml:"sources"`
	Network  NetworkConfig  `yaml:"network"`
	Database DatabaseConfig `yaml:"database"`
}

// Source defines the configuration for a single source in the Bivrost system.
type Source struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Location string   `yaml:"location"`
	Format   string   `yaml:"format"`
	Tags     []string `yaml:"tags"`
}

// NetworkConfig defines network-related configuration settings.
type NetworkConfig struct {
	ReadTimeout  int `yaml:"read_timeout,omitempty"`
	WriteTimeout int `yaml:"write_timeout,omitempty"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}
