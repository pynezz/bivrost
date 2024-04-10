package modules

// Module represents a module
type Module struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Config      ModuleConfig `json:"config"`
}

type ModuleConfig struct {
	Name     string `yaml:"name"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database, omitempty"`
	DataSources []struct {
		Name     string `yaml:"name"`
		Type     string `yaml:"type"`
		Location string `yaml:"location"`
		Format   string `yaml:"format, omitempty"`
	} `yaml:"data_sources, omitempty"`
}
