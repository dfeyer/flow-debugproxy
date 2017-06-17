// Copyright 2015 Dominique Feyer <dfeyer@ttree.ch>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xdebugproxy

import (
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/logger"
	"github.com/dfeyer/flow-debugproxy/pathmapping"

	"fmt"
	"io"
	"net"
)

const h = "%s"

// XDebugProcessorPlugin process message in xDebug protocol
type XDebugProcessorPlugin interface {
	Initialize(c *config.Config, l *logger.Logger, m *pathmapping.PathMapping)
	ApplyMappingToTextProtocol(message []byte) []byte
	ApplyMappingToXML(message []byte) []byte
}

// Proxy represents a pair of connections and their state
type Proxy struct {
	sentBytes      uint64
	receivedBytes  uint64
	Raddr          *net.TCPAddr
	Lconn, rconn   *net.TCPConn
	PathMapper     XDebugProcessorPlugin
	Config         *config.Config
	Logger         *logger.Logger
	postProcessors []XDebugProcessorPlugin
	pipeErrors     chan error
}

// Start the proxy
func (p *Proxy) Start() {
	defer p.Lconn.Close()

	// connect to remote
	rconn, err := net.DialTCP("tcp", nil, p.Raddr)
	if err != nil {
		p.log(h, "Unable to connect to your IDE, please check if your editor listen to incoming connection")
		p.log("Error message: %s", err)
		p.log(h, "Configure your IDE and reload the web page should solve this issue")
		p.log(h, "\nHit Ctrl-C to exit the proxy if don't need it ...")
		p.log(h, "\nYour fellow Umpa Lumpa")
		return
	}

	p.rconn = rconn
	defer p.rconn.Close()

	p.pipeErrors = make(chan error)
	defer close(p.pipeErrors)

	// display both ends
	p.log("Opened %s >>> %s", p.Lconn.RemoteAddr().String(), p.rconn.RemoteAddr().String())
	// bidirectional copy
	go p.pipe(p.Lconn, p.rconn)
	go p.pipe(p.rconn, p.Lconn)

	if err = <-p.pipeErrors; err != io.EOF {
		p.Logger.Warn(h, err)
	}
	<-p.pipeErrors

	p.log("Closed (%d bytes sent, %d bytes recieved)", p.sentBytes, p.receivedBytes)
}

// RegisterPostProcessor add a new message post processor
func (p *Proxy) RegisterPostProcessor(processor XDebugProcessorPlugin) {
	p.postProcessors = append(p.postProcessors, processor)
}

func (p *Proxy) log(s string, args ...interface{}) {
	if p.Config.Verbose {
		p.Logger.Info(s, args...)
	}
}

func (p *Proxy) pipe(src, dst *net.TCPConn) {
	// data direction
	var f, h string
	var processor XDebugProcessorPlugin
	isFromDebugger := src == p.Lconn
	if isFromDebugger {
		f = "\nDebugger >>> IDE\n================"
	} else {
		f = "\nIDE >>> Debugger\n================"
	}
	h = "%s"
	// directional copy (64k buffer)
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			p.pipeErrors <- err
			// make sure the other pipe will stop as well
			dst.Close()
			return
		}
		b := buff[:n]
		p.log(h, f)
		if p.Config.VeryVerbose {
			if isFromDebugger {
				p.log("Raw protocol:\n%s\n", p.Logger.Colorize(fmt.Sprintf(h, p.Logger.FormatXMLProtocol(b)), "blue"))
			} else {
				p.log("Raw protocol:\n%s\n", p.Logger.Colorize(fmt.Sprintf(h, p.Logger.FormatTextProtocol(b)), "blue"))
			}
		}
		// extract command name
		if isFromDebugger {
			b = p.PathMapper.ApplyMappingToXML(b)
			// post processors
			for _, d := range p.postProcessors {
				processor = d
				b = processor.ApplyMappingToXML(b)
			}
		} else {
			b = p.PathMapper.ApplyMappingToTextProtocol(b)
			// post processors
			for _, d := range p.postProcessors {
				processor = d
				b = processor.ApplyMappingToTextProtocol(b)
			}
		}

		// show output
		if p.Config.VeryVerbose {
			if isFromDebugger {
				p.log("Processed protocol:\n%s\n", p.Logger.Colorize(fmt.Sprintf(h, p.Logger.FormatXMLProtocol(b)), "blue"))
			} else {
				p.log("Processed protocol:\n%s\n", p.Logger.Colorize(fmt.Sprintf(h, p.Logger.FormatTextProtocol(b)), "blue"))
			}
		} else {
			p.log(h, "")
		}

		// write out result
		n, err = dst.Write(b)
		if err != nil {
			p.pipeErrors <- err
			// make sure the other pipe will stop as well
			src.Close()
			return
		}
		if isFromDebugger {
			p.sentBytes += uint64(n)
		} else {
			p.receivedBytes += uint64(n)
		}
	}
}
