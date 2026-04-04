package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
)

var (
	Success   = color.Green
	Error     = color.Red
	Warning   = color.Yellow
	Info      = color.Cyan
	Header    = color.Bold
	Primary   = color.Magenta
	Secondary = color.Blue
)

func init() {
	color.ForceColor()
}

func PrintSuccess(format string, args ...interface{}) {
	Success.Printf(format, args...)
	fmt.Println()
}

func PrintError(format string, args ...interface{}) {
	Error.Printf("Error: ")
	fmt.Printf(format, args...)
	fmt.Println()
}

func PrintWarning(format string, args ...interface{}) {
	Warning.Printf("Warning: ")
	fmt.Printf(format, args...)
	fmt.Println()
}

func PrintInfo(format string, args ...interface{}) {
	Info.Printf(format, args...)
	fmt.Println()
}

func PrintHeader(format string, args ...interface{}) {
	Header.Printf(format, args...)
	fmt.Println()
}

func IsTerminal() bool {
	return os.Getenv("TERM") != "dumb" || os.Getenv("FORCE_COLOR") != ""
}

func DisableColors() {
	color.Enable = false
}

func PrintBox(title, content string) {
	lines := strings.Split(content, "\n")
	allLines := []string{title}
	allLines = append(allLines, lines...)

	maxLen := 0
	for _, line := range allLines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	if maxLen > 80 {
		maxLen = 80
	}

	border := "+" + strings.Repeat("-", maxLen+2) + "+"

	fmt.Println(border)
	for i, line := range allLines {
		if len(line) > maxLen {
			line = line[:maxLen-3] + "..."
		}
		padding := maxLen - len(line)
		if i == 0 {
			Header.Printf("| %s%s |\n", line, strings.Repeat(" ", padding))
		} else {
			fmt.Printf("| %s%s |\n", line, strings.Repeat(" ", padding))
		}
	}
	fmt.Println(border)
}
