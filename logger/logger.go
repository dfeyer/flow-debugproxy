// Copyright 2015 Dominique Feyer <dfeyer@ttree.ch>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"github.com/dfeyer/flow-debugproxy/config"

	"fmt"
	"github.com/mgutz/ansi"
)

var (
	debugize = ansi.ColorFunc("green+h:black")
	greenize = ansi.ColorFunc("green")
	redize   = ansi.ColorFunc("red")
)

// Logger handle log message
type Logger struct {
	Config *config.Config
}

//Debug output a debug text
func (l *Logger) Debug(f string, args ...interface{}) {
	if l.Config.Debug {
		fmt.Printf(debugize("[DEBUG] "+f)+"\n", args...)
	}
}

//Info output a green text line
func (l *Logger) Info(f string, args ...interface{}) {
	fmt.Printf(greenize(f)+"\n", args...)
}

//Warn output a red text line
func (l *Logger) Warn(f string, args ...interface{}) {
	fmt.Printf(redize(f)+"\n", args...)
}
