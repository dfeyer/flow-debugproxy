package logger

import (
	"fmt"
	"github.com/mgutz/ansi"
)

func Info(f string, args ...interface{}) {
	fmt.Printf(Colorize(f, "green")+"\n", args...)
}

func Warn(f string, args ...interface{}) {
	fmt.Printf(Colorize(f, "red")+"\n", args...)
}

//helpers
func Colorize(str, style string) string {
	return ansi.Color(str, style)
}
