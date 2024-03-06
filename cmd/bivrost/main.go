package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pynezz/bivrost/internal/config"

	"github.com/pynezz/bivrost/internal/util/flags"

	"github.com/pynezz/bivrost/internal/tui"
	"github.com/pynezz/bivrost/internal/util"
)

func main() {
	tui.Header.Color = util.Cyan
	tui.Header.PrintHeader()

	if len(os.Args) < 2 {
		util.PrintWarning("No arguments provided. Use -h for help.")

		flag.Usage()
		return
	}

	flags.ParseFlags()

	cfg, err := config.LoadConfig(*flags.Params.ConfigPath)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Exiting...")
		return
	}

	fmt.Println(cfg)

	// app := api.NewServer(cfg)
	// app.Listen(":3000")
}
