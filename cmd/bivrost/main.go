package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pynezz/bivrost/internal/api"
	"github.com/pynezz/bivrost/internal/config"
	"github.com/pynezz/bivrost/pkg/version"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	help := flag.Bool("help", false, "Print this help message")
	versionFlag := flag.Bool("version", false, "Print version information")

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *versionFlag {
		fmt.Println(version.Info())
		return
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %s\n", err)
		os.Exit(1)
	}

	apiServer := api.NewServer(cfg)
	apiServer.Start()
}
