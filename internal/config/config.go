package config

import (
	"fmt"
	"os"

	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/util"

	"gopkg.in/yaml.v3"
)

type Cfg struct {
	Sources []struct {
		Name     string   `yaml:"name"`
		Type     string   `yaml:"type"`
		Location string   `yaml:"location"`
		Format   string   `yaml:"format"`
		Tags     []string `yaml:"tags"`
	} `yaml:"sources"`
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

// FIBER says POST Method Not allowed.

// The reason for this is that the Fiber framework does not allow the POST method by default.
// To enable it, you need to add the following line to the main.go file:
// app.Post("/config/:id", updateConfigHandler)
// This will allow the POST method to be used on the /config/:id route.
// But you also need to implement the updateConfigHandler function in the server.go file.
// This function should handle the POST request and update the configuration.
// Here is an example of how you can implement the updateConfigHandler function:
// func updateConfigHandler(c *fiber.Ctx) error {
// 	// Update the configuration here
// 	id := c.Params("id")
// 	fmt.Println("Updating configuration for ID:", id)
// 	return c.SendString("Configuration updated")
// }
