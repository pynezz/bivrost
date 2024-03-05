package util

import (
	"flag"
	"fmt"
	"os"

	"github.com/pynezz/bivrost/pkg/version"
)

var (
	versionFlagS *bool
	versionFlagL *bool
)

func init() {
	versionFlagS = flag.Bool("v", false, "Print version information")
	versionFlagL = flag.Bool("version", false, "Print version information")
}

// Get version information
func GetVersion() {
	if *versionFlagS || *versionFlagL {
		fmt.Println(version.Info())
		os.Exit(0)
	}
}

func ParseFlags() {
	flag.Parse()

	switch {
	case *versionFlagS || *versionFlagL:
		fmt.Println(version.Info())
		os.Exit(0)
	}

}
