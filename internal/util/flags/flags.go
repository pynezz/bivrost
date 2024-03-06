package flags

import (
	"flag"
	"fmt"

	"github.com/pynezz/bivrost/internal/util"
	"github.com/pynezz/bivrost/pkg/version"
)

type Arguments struct {
	ConfigPath *string
}

var (
	versionFlag bool

	helpFlag bool

	configPathL *string
	configPathS *string

	Params Arguments
)

const usage = `Usage: bivrost [options]

Options:
  -v, --version    		Print version information
  -h, --help       		Print this help message
  -c, --config PATH		Path to the configuration file`

func init() {
	flag.Usage = func() {
		fmt.Println(usage)
	}
	// Define flags
	flag.BoolVar(&versionFlag, "version", false, version.Version())
	flag.BoolVar(&versionFlag, "v", false, version.Version())

	flag.BoolVar(&helpFlag, "h", false, "Print this help message")
	flag.BoolVar(&helpFlag, "help", false, "Print this help message")

	configPathL = flag.String("config", "config.yaml", "Path to the configuration file")
	configPathS = flag.String("c", "", "")
	Params.ConfigPath = configPathL // Default value of "config.yaml" (will be overwritten if the flag is set)
}

// // Get version information
// func GetVersion() {
// 	if *versionFlagS || *versionFlagL {
// 		fmt.Println(version.Info())
// 		os.Exit(0)
// 	}
// }

func ParseFlags() *Arguments {
	flag.Parse()

	switch {
	case helpFlag:
		flag.Usage()
	case versionFlag:
		fmt.Println(version.Info())
	case *configPathL != "" || *configPathS != "":
		if *configPathL != "" {
			Params.ConfigPath = configPathL
		} else {
			Params.ConfigPath = configPathS
		}
	default:
		util.PrintWarning("This should not happen. Please report this issue.") // It should be taken care of in main.go
	}

	return &Params
}
