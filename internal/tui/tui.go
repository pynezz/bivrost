package tui

import (
	"fmt"

	"github.com/pynezz/bivrost/internal/util"
)

type HeaderStruct struct {
	Color   string
	Version string
	Content string
}

var (
	// Version is the version of the application
	Version = "0.0.1"

	Header = &HeaderStruct{
		Color:   "1",
		Version: Version,
		Content: AsciiArt(),
	}
)

func init() {

}

func AsciiArt() string {
	return `
██████╗ ██╗██╗   ██╗██████╗  ██████╗ ███████╗████████╗
██╔══██╗██║██║   ██║██╔══██╗██╔═══██╗██╔════╝╚══██╔══╝
██████╔╝██║██║   ██║██████╔╝██║   ██║███████╗   ██║
██╔══██╗██║╚██╗ ██╔╝██╔══██╗██║   ██║╚════██║   ██║
██████╔╝██║ ╚████╔╝ ██║  ██║╚██████╔╝███████║   ██║
╚═════╝ ╚═╝  ╚═══╝  ╚═╝  ╚═╝ ╚═════╝ ╚══════╝   ╚═╝
`
}

// Should be used in conjunction with the util package for proper formatting
func (h *HeaderStruct) ColorHeader(color string) string {

	return util.ColorF(color, h.Content)
}

func (h *HeaderStruct) PrintHeader() {
	if h.Color != "" {
		fmt.Println(h.ColorHeader(h.Color))
	} else {
		fmt.Println(Header.Content)
	}
}
