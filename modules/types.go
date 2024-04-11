package modules

// Module represents a module
type Module struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ConfigPath  string `json:"config_path"`
	Config      *ModuleConfig
}

type ModuleConfig struct {
	Name       string `yaml:"name"`
	Identifier string `yaml:"identifier"`
	Database   struct {
		Path string `yaml:"path"`
	} `yaml:"database,omitempty"`
	DataSources []struct {
		Name     string `yaml:"name"`
		Type     string `yaml:"type"`
		Location string `yaml:"location"`
		Format   string `yaml:"format,omitempty"`
	} `yaml:"data_sources,omitempty"`
}

// This have become grossly overcomplicated. Just use a simple struct instead in the future.
// Ex:
// type MID struct {
// 	StrID      string
// 	Identifier []byte
// }
// var MODULEIDENTIFIERS map[string]MID
//
// MODULEIDENTIFIERS is a map of module names to their identifiers
var MODULEIDENTIFIERS map[string][4]byte

func init() {
	MODULEIDENTIFIERS = make(map[string][4]byte)
}

type ModuleIdentifiers map[string][4]byte
