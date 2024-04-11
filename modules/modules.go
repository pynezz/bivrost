/*
 Package modules provides a way to load and manage modules.
 Modules are loaded from the modules directory, and are expected to be a configuration file, and a binary file.
*/

package modules

import (
	"fmt"

	"github.com/pynezz/bivrost/internal/config"
	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/util"
	"gopkg.in/yaml.v3"
)

func (m *Module) Stop() {
	// Load the modules from the modules directory
	// For each module, load the configuration and the binary
	// Start the module with the configuration
}

func StartModule() {
	// Start the module with the provided configuration
}

var Modules map[string][]Module

var Mids ModuleIdentifiers

// LoadModules loads all the modules and their configuration files
func LoadModules(config config.Cfg) error {
	util.PrintDebug("Loading modules")
	if Modules != nil {
		return fmt.Errorf("modules already loaded")
	}
	Modules = make(map[string][]Module)

	var m ModuleIdentifiers = make(map[string][4]byte)
	Mids = m

	modules := func() []Module {
		var m []Module
		for _, module := range config.Sources {
			if module.Type == "module" {
				found := Module{
					Name:        module.Name,
					Description: module.Description,
					ConfigPath:  module.Config,
				}
				m = append(m, found)
			} else {
				util.PrintDebug("Skipping non-module source " + module.Name + " of type " + module.Type + "...")
			}
		}
		return m
	}

	util.PrintDebug("Loading module configurations...")
	for _, module := range modules() {
		conf, err := fsutil.GetFile(fsutil.PathConvert(module.ConfigPath))
		if err != nil {
			return err
		}
		defer conf.Close()

		util.PrintDebug("Decoding module configuration for " + module.Name + "...")
		yamlDecoder := yaml.NewDecoder(conf)
		var mc ModuleConfig
		err = yamlDecoder.Decode(&mc)
		if err != nil {
			util.PrintError("Failed to decode module configuration for " + module.Name)
			return err
		}

		module.Config = &mc
		Modules[module.Name] = append(Modules[module.Name], module)
		util.PrintSuccess("Loaded module " + module.Name)

		id := [4]byte{}
		for i := range id {
			id[i] = mc.Identifier[i]
		}
		Mids[module.Name] = id
		util.PrintSuccess("Loaded identifier for " + module.Name + " as " + fmt.Sprintf("%v", id))
		Mids.StoreModuleIdentifier(module.Name, id)
	}
	return nil
}

// GetModuleIdentifiers returns the module identifiers
func (i *ModuleIdentifiers) GetModuleIdentifier(name string) [4]byte {
	return (*i)[name]
}

// StoreModuleIdentifier stores the module identifier in the MODULEIDENTIFIERS map
func (i *ModuleIdentifiers) StoreModuleIdentifier(name string, identifier [4]byte) {
	(*i)[name] = identifier
}

// GetModuleName returns the name of the module from the identifier
func (i *ModuleIdentifiers) GetModuleName(identifier [4]byte) string {
	for name, id := range *i {
		if id == identifier {
			return name
		}
	}
	return ""
}

// GetModuleNames returns the names of the modules
func (i *ModuleIdentifiers) GetModuleNames() []string {
	var names []string
	for name := range *i {
		names = append(names, name)
	}
	return names
}

// AddModuleIdentifier adds a module identifier to the MODULEIDENTIFIERS map
func SetModuleIdentifier(identifier [4]byte, name string) {
	MODULEIDENTIFIERS[name] = identifier
}

// GetModuleNameFromID returns the name of the module from the identifier
func GetModuleNameFromID(identifier [4]byte) string {
	for name, id := range MODULEIDENTIFIERS {
		if id == identifier {
			return name
		}
	}
	return ""
}
