package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pynezz/bivrost/cmd/bivrost"
)

func main() {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	isPackage := false

	// If the path contains "bivrost", it is not a package
	// so flags should be parsed by the bivrost package
	// otherwise, the flags should be parsed by the main package
	for _, path := range strings.Split(path, "/") {
		if path == "bivrost" {
			isPackage = false
			fmt.Println("[!] Bivrost is not a package in this context")
		}
	}

	bivrost.Execute(isPackage)
}
