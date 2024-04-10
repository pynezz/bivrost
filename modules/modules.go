/*
 Package modules provides a way to load and manage modules.
 Modules are loaded from the modules directory, and are expected to be a configuration file, and a binary file.
*/

package modules

import (
	"fmt"

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

func LoadModules() error {
	if Modules != nil {
		return fmt.Errorf("modules already loaded")
	}

	moduleDirs, err := fsutil.GetDirs("./modules")
	if err != nil {
		return err
	}

	util.PrintSuccess("Found modules: " + fmt.Sprint(moduleDirs))

	Modules = make(map[string][]Module)

	for _, dir := range moduleDirs {
		conf, err := fsutil.GetFile(dir + "/config.yaml")
		if err != nil {
			return err
		}

		util.PrintSuccess("Found config file for module " + dir)

		var mc ModuleConfig
		decoder := yaml.NewDecoder(conf)
		err = decoder.Decode(&mc)
		if err != nil {
			return err
		}

		m := Module{ // Load the module configuration
			Name:        dir,
			Description: "Module " + dir,
			Config:      mc,
		}

		Modules[dir] = append(Modules[dir], m)
		util.PrintSuccess("Loaded module " + dir)
	}

	return nil
}
