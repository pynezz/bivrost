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

// PrintSuccess prints a success message to the console
func PrintSuccess(msg string) {
	fmt.Printf("%s[+]%s %s\n", Green, Reset, msg)
}

// PrintError prints an error message to the console
func PrintError(msg string) {
	fmt.Printf("%s[!]%s %s\n", Red, Reset, msg)
}

// PrintInfo prints an info message to the console
func PrintInfo(msg string) {
	fmt.Printf("%s[i]%s %s\n", Cyan, Reset, msg)
}

// PrintWarning prints a warning message to the console
func PrintWarning(msg string) {
	fmt.Printf("%s[-]%s %s\n", Yellow, Reset, msg)
}

// PrintDebug prints a debug message to the console
func PrintDebug(msg string) {
	fmt.Printf("%s[DEBUG]%s %s\n", Gray, Reset, msg)
}

// PrintBold prints a bold message to the console
func PrintBold(msg string) {
	fmt.Printf("%s%s%s\n", Bold, msg, Reset)
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
