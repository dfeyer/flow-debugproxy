package logger

import (
	"bytes"
	"fmt"
	"github.com/mgutz/ansi"
)

//Info output a green text line
func Info(f string, args ...interface{}) {
	fmt.Printf(Colorize(f, "green")+"\n", args...)
}

//Warn output a red text line
func Warn(f string, args ...interface{}) {
	fmt.Printf(Colorize(f, "red")+"\n", args...)
}

//Colorize use the Ansi module to colorize output
func Colorize(str, style string) string {
	return ansi.Color(str, style)
}

//FormatTextProtocol replace NULL by a line break for output formatting
func FormatTextProtocol(protocol []byte) []byte {
	return bytes.Trim(bytes.Replace(protocol, []byte("\x00"), []byte("\n"), -1), "\n")
}
