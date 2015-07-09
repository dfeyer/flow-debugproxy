package logger

import (
	"bytes"
	"fmt"
	"github.com/mgutz/ansi"
)

var (
	debugize = ansi.ColorFunc("green+h:black")
	greenize = ansi.ColorFunc("green")
	redize   = ansi.ColorFunc("red")
)

//Debug output a debug text
func Debug(f string, args ...interface{}) {
	fmt.Printf(debugize("[DEBUG] "+f)+"\n", args...)
}

//Info output a green text line
func Info(f string, args ...interface{}) {
	fmt.Printf(greenize(f)+"\n", args...)
}

//Warn output a red text line
func Warn(f string, args ...interface{}) {
	fmt.Printf(redize(f)+"\n", args...)
}

//Colorize use the Ansi module to colorize output
func Colorize(str, style string) string {
	return ansi.Color(str, style)
}

//FormatTextProtocol replace NULL by a line break for output formatting
func FormatTextProtocol(protocol []byte) []byte {
	return bytes.Trim(bytes.Replace(protocol, []byte("\x00"), []byte("\n"), -1), "\n")
}
