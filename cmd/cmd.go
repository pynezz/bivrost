package main

import (
	"github.com/pynezz/bivrost/cmd/bivrost"
	"github.com/pynezz/bivrost/internal/util"
)

func main() {
	util.PrintInfo("Starting Bivrost...")
	bivrost.Execute()
}
