// package main
package main

import "github.com/pynezz/bivrost/cmd"

var buildVersion string

func Execute() {
	cmd.Execute(buildVersion)
}

func main() {
	cmd.Execute(buildVersion)
}
