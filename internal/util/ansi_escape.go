package util

import (
	"fmt"
)

// Ansi colors
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

// Ansi styles
const (
	Bold      = "\033[1m"
	Underline = "\033[4m"
	Inverse   = "\033[7m"
)

// Ansi 256 light colors
const (
	LightRed    = "\033[91m"
	LightGreen  = "\033[92m"
	LightYellow = "\033[93m"
	LightBlue   = "\033[94m"
	LightPurple = "\033[95m"
	LightCyan   = "\033[96m"
)

// Ansi 256 dark colors
const (
	DarkRed    = "\033[31m"
	DarkGreen  = "\033[32m"
	DarkYellow = "\033[33m"
	DarkBlue   = "\033[34m"
	DarkPurple = "\033[35m"
	DarkCyan   = "\033[36m"
)

// Background colors
const (
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
	BgPurple = "\033[45m"
	BgCyan   = "\033[46m"
	BgGray   = "\033[47m"
)

// Cursor movement
const (
	CursorUp    = "\033[A"
	CursorDown  = "\033[B"
	CursorRight = "\033[C"
	CursorLeft  = "\033[D"
)

// Terminal control
const (
	ClearScreen = "\033[2J"
	ClearLine   = "\033[K"
	Backspace   = "\b"
	Delete      = "\033[3~"
	Enter       = "\r"
	Tab         = "\t"

	// Cursor positioning
	Home     = "\033[H"
	Position = "\033[%d;%dH"

	// Save and restore cursor position
	SaveCursor    = "\033[s"
	RestoreCursor = "\033[u"

	// Hide and show cursor
	HideCursor = "\033[?25l"
	ShowCursor = "\033[?25h"

	// Overwrite the current line
	Overwrite = "\033[1A\033[2K"
)

const (
	RoundedTopLeft     = "╭"
	RoundedTopRight    = "╮"
	RoundedBottomLeft  = "╰"
	RoundedBottomRight = "╯"
	RoundedHoriz       = "─"
	RoundedVert        = "│"
)

func SPrintRoundedTop(width int) string {
	rtop := ""
	for i := 0; i < width-2; i++ {
		rtop += RoundedHoriz
	}
	return fmt.Sprintf("%s%s%s", RoundedTopLeft, rtop, RoundedTopRight)
}

func SPrintRoundedBottom(width int) string {
	rbottom := ""
	for i := 0; i < width-2; i++ {
		rbottom += RoundedHoriz
	}
	return fmt.Sprintf("%s%s%s", RoundedBottomLeft, rbottom, RoundedBottomRight)
}

func PrintRoundedTop(width int) {
	fmt.Print(RoundedTopLeft)
	for i := 0; i < width-2; i++ {
		fmt.Print(RoundedHoriz)
	}
	fmt.Print(RoundedTopRight)
}

func PrintRoundedBottom(width int) {
	fmt.Print(RoundedBottomLeft)
	for i := 0; i < width-2; i++ {
		fmt.Print(RoundedHoriz)
	}
	fmt.Print(RoundedBottomRight)
}

func AddPadding(content string, width int) string {
	padding := width - len(content)
	for i := 0; i < padding+1; i++ {
		content += " "
	}
	return content
}

// Get the width of a multiline string
func GetWidth(content string) int {
	width := 0
	tmpW := 0
	for _, c := range content {
		if c == '\n' {
			if width < tmpW {
				width = tmpW
			}
			tmpW = 0
		} else {
			tmpW++
		}
	}
	if width == 0 {
		width = tmpW
	}
	return width
}

// FormatRoundedBox formats a string into a rounded box
// ! Do not spend a lot of time on this, it is not important
func FormatRoundedBox(content string) string {
	tmpW := 0 // Temporary width
	w := 0    // Actual final width
	result := ""

	lines := []string{}

	for i, c := range content {
		if c == '\n' {
			if w < tmpW {
				w = tmpW
			}
			// result += fmt.Sprintf("%s %s %s\n", RoundedVert, content[i-tmpW:i], RoundedVert) // Should be │ content │
			lines = append(lines, fmt.Sprintf("%s %s ", RoundedVert, content[i-tmpW:i])) // Add to the line slice

			tmpW = 0
		} else {
			tmpW++
		}
	}

	if w == 0 {
		w = tmpW // If there are no newlines
	}

	finres := ""

	w += 4 // Add 4 for the corners + padding

	for _, l := range lines {
		finres += AddPadding(l, w) + RoundedVert + "\n"
	}

	result = SPrintRoundedTop(w) + "\n" + finres + SPrintRoundedBottom(w)

	return result
}

// PrintSuccess prints a success message to the console
func PrintSuccess(msg string) {
	fmt.Printf("%s[+]%s %s\n", Green, Reset, msg)
}

// PrintError prints an error message to the console
func PrintError(msg string) {
	fmt.Printf("%s[!]%s %s\n", Red, Reset, msg)
}

// PrintErrorf prints a formatted error message to the console
func PrintErrorf(format string, a ...interface{}) {
	fmt.Printf("%s[!]%s %s\n", Red, Reset, fmt.Sprintf(format, a...))
}

func ColorF(color, format string, a ...interface{}) string {
	return fmt.Sprintf("%s%s%s", color, fmt.Sprintf(format, a...), Reset)
}

func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf("%s[!]%s %s", Red, Reset, fmt.Sprintf(format, a...))
}

func ItalicF(format string, a ...interface{}) string {
	return fmt.Sprintf("%s%s%s", "\033[3m", fmt.Sprintf(format, a...), Reset)
}

// PrintInfo prints an info message to the console
func PrintInfo(msg string) {
	fmt.Printf("%s[i]%s %s\n", Cyan, Reset, msg)
}

// PrintWarning prints a warning message to the console
func PrintWarning(msg string) {
	fmt.Printf("%s⚠️%s %s\n", Yellow, Reset, msg)
}

// PrintDebug prints a debug message to the console
func PrintDebug(msg string) {
	fmt.Printf("%s[DEBUG]%s %s\n", Gray, Reset, msg)
}

// PrintBold prints a bold message to the console
func PrintBold(msg string) {
	fmt.Printf("%s%s%s\n", Bold, msg, Reset)
}

// PrintItalic prints an italic message to the console
func PrintItalic(msg string) {
	fmt.Printf("%s%s%s\n", "\033[3m", msg, Reset)
}

// PrintUnderline prints an underlined message to the console
func PrintUnderline(msg string) {
	fmt.Printf("%s%s%s\n", Underline, msg, Reset)
}

// PrintInverse prints an inverted message to the console
func PrintInverse(msg string) {
	fmt.Printf("%s%s%s\n", Inverse, msg, Reset)
}

// PrintColor prints a colored message to the console
func PrintColor(color, msg string) {
	fmt.Printf("%s%s%s\n", color, msg, Reset)
}

// PrintColorf prints a colored formatted message to the console
func PrintColorf(color, format string, a ...interface{}) {
	fmt.Printf("%s%s%s\n", color, fmt.Sprintf(format, a...), Reset)
}

func PrintColorBold(color, msg string) {
	fmt.Printf("%s%s%s\n", color+Bold, msg, Reset)
}

func PrintColorUnderline(color, msg string) {
	fmt.Printf("%s%s%s\n", color+Underline, msg, Reset)
}

func PrintColorAndBg(color, bg, msg string) {
	fmt.Printf("%s%s%s\n", color+bg, msg, Reset)
}

func PrintColorAndBgBold(color, bg, msg string) {
	fmt.Printf("%s%s%s\n", color+bg+Bold, msg, Reset)
}
