package tui

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/pynezz/bivrost/internal/util"
)

type HeaderStruct struct {
	Color   string
	Version string
	Content string
}

type DataPanel struct {
	Data       <-chan string
	Paragraphs []widgets.Paragraph

	Height int
	Width  int

	PositionX int
	PositionY int

	Border      bool
	BorderColor string
}

type Tui struct {
	Header *HeaderStruct
	Height int
	Width  int

	PositionX int
	PositionY int

	Panels []DataPanel
}

type TuiInterface interface {
	ColorHeader(color string) string
	PrintHeader()
	AddDataSource(data <-chan string)
	AddDataSink()
	Display()
}

func NewDataPanel(data <-chan string, height, width, positionX, positionY int, border bool, borderColor string) *DataPanel {
	return &DataPanel{
		Data:        data,
		Height:      height,
		Width:       width,
		PositionX:   positionX,
		PositionY:   positionY,
		Border:      border,
		BorderColor: borderColor,
	}
}

var (
	// Version is the version of the application
	Version = "0.0.2"

	Header = &HeaderStruct{
		Color:   "1",
		Version: Version,
		Content: AsciiArt(),
	}
)

func AsciiArt() string {
	return `
██████╗ ██╗██╗   ██╗██████╗  ██████╗ ███████╗████████╗
██╔══██╗██║██║   ██║██╔══██╗██╔═══██╗██╔════╝╚══██╔══╝
██████╔╝██║██║   ██║██████╔╝██║   ██║███████╗   ██║
██╔══██╗██║╚██╗ ██╔╝██╔══██╗██║   ██║╚════██║   ██║
██████╔╝██║ ╚████╔╝ ██║  ██║╚██████╔╝███████║   ██║
╚═════╝ ╚═╝  ╚═══╝  ╚═╝  ╚═╝ ╚═════╝ ╚══════╝   ╚═╝
%s`
}

// Should be used in conjunction with the util package for proper formatting
func (h *HeaderStruct) ColorHeader(color string) string {
	return util.ColorF(color, h.Content, h.Version)
}

func (h *HeaderStruct) PrintHeader() {
	if h.Color != "" {
		fmt.Println(h.ColorHeader(h.Color))
	} else {
		fmt.Println(Header.Content, h.Version)
	}
}

func gitCommits() string {
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	version, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to get git commit count: %v", err)
	}
	return fmt.Sprintf("%s\n", version)
}

func NewTui() *Tui {
	h := Header
	h.Version = fmt.Sprintf("v.0.2.%s", gitCommits())
	return &Tui{
		Header: h,
	}
}

func (t *Tui) AddDataSource(data <-chan string, title string, color string) {
	// Initialize the paragraph widget for new data
	p := widgets.NewParagraph()
	p.Title = title
	p.BorderStyle.Fg = ui.ColorCyan
	p.SetRect(t.PositionX, t.PositionY, t.Width, t.Height)
	t.Panels = append(t.Panels, DataPanel{
		Data: data,
		Paragraphs: []widgets.Paragraph{
			*p,
		},
		Height:      t.Height,
		Width:       t.Width,
		PositionX:   t.PositionX,
		PositionY:   t.PositionY,
		Border:      true,
		BorderColor: color,
	})

	// Go routine to update panel content from data channel
	go func() {
		for msg := range data {
			p.Text += msg + "\n" // Append new messages to the existing text
			ui.Render(p)         // Re-render the UI with the updated text
		}
	}()
}

func (t *Tui) AddDataSink() {

}

func (t *Tui) Display() {
	Header.PrintHeader()
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p1 := widgets.NewParagraph()
	p1.Title = "Other Feature Output"
	p1.Text = "Initializing..."
	p1.SetRect(0, 10, 50, 20)
	p1.BorderStyle.Fg = ui.ColorGreen

	ui.Render(p1)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}

func TerminalUI() {
	Header.PrintHeader()
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p1 := widgets.NewParagraph()
	p1.Title = "Watcher Output"
	p1.Text = "Starting..."
	p1.SetRect(0, 0, 50, 10)
	p1.BorderStyle.Fg = ui.ColorYellow

	p2 := widgets.NewParagraph()
	p2.Title = "Other Feature Output"
	p2.Text = "Initializing..."
	p2.SetRect(0, 10, 50, 20)
	p2.BorderStyle.Fg = ui.ColorGreen

	ui.Render(p1, p2)

	// Example to update contents dynamically
	go func() {
		for {
			time.Sleep(time.Second)
			p1.Text = "File watcher update at " + time.Now().Format("15:04:05")
			p2.Text = "Feature update at " + time.Now().Format("15:04:05")
			ui.Render(p1, p2)
		}
	}()

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}
