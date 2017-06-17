// Copyright 2015 Dominique Feyer <dfeyer@ttree.ch>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"encoding/xml"
	"os"
	"regexp"
	"strings"

	"github.com/clbanning/mxj"
	"github.com/dfeyer/flow-debugproxy/config"

	"bytes"
	"fmt"

	"github.com/mgutz/ansi"
)

var (
	debugize          = ansi.ColorFunc("green+h:black")
	greenize          = ansi.ColorFunc("green")
	redize            = ansi.ColorFunc("red")
	regexpFirstNumber = regexp.MustCompile(`^[0-9]*`)
)

type node struct {
	Attr     []xml.Attr
	XMLName  xml.Name
	Children []node `xml:",any"`
	Text     string `xml:",chardata"`
}

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

//Colorize use the Ansi module to colorize output
func (l *Logger) Colorize(str, style string) string {
	return ansi.Color(str, style)
}

func normalizeXMLProtocol(buffer []byte) string {
	b := bytes.Replace(buffer, []byte("\x00"), []byte("\n"), -1)
	buf := make([]rune, len(b))
	for i, b := range b {
		buf[i] = rune(b)
	}
	s := strings.Replace(string(buf), "iso-8859-1", "utf-8", 1)
	s = regexpFirstNumber.ReplaceAllString(s, "")
	s = strings.Replace(s, "<?xml version=\"1.0\" encoding=\"utf-8\"?>", "", 1)
	s = strings.Trim(s, "\n")
	return s
}

//FormatXMLProtocol beautify XML output
func (l *Logger) FormatXMLProtocol(protocol []byte) []byte {
	p := normalizeXMLProtocol(protocol)
	output, err := mxj.BeautifyXml([]byte(p), "", "  ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return output
}

//FormatTextProtocol replace NULL by a line break for output formatting
func (l *Logger) FormatTextProtocol(protocol []byte) []byte {
	return bytes.Trim(bytes.Replace(protocol, []byte("\x00"), []byte("\n"), -1), "\n")
}
