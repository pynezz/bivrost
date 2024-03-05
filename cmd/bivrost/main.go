package main

import (
	"flag"
	"fmt"

	"github.com/pynezz/bivrost/internal/api"
	"github.com/pynezz/bivrost/internal/config"

	"github.com/pynezz/bivrost/internal/util"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	// help := flag.Bool("help", false, "Print this help message")
	// versionFlag := flag.Bool("version", false, "Print version information")

	util.ParseFlags()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Exiting...")
		return
	}

	app := api.NewServer(cfg)
	app.Listen(":3000")
}
