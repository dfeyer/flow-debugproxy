package xdebugproxy

import (
	"github.com/dfeyer/flow-debugproxy/logger"
	"github.com/dfeyer/flow-debugproxy/pathmapper"

	"fmt"
	"io"
	"net"
)

//todo replace by config module when ready
var verbose = true
var veryverbose = true

//Proxy represents a pair of connections and their state
type Proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	Laddr, Raddr  *net.TCPAddr
	Lconn, rconn  *net.TCPConn
	Erred         bool
	Errsig        chan bool
}

func (p *Proxy) log(s string, args ...interface{}) {
	if verbose {
		logger.Info(s, args...)
	}
}

func (p *Proxy) err(s string, err error) {
	if p.Erred {
		return
	}
	if err != io.EOF {
		logger.Warn(s, err)
	}
	p.Errsig <- true
	p.Erred = true
}

func (p *Proxy) Start() {
	var h = "%s"
	defer p.Lconn.Close()
	//connect to remote
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
	//display both ends
	p.log("Opened %s >>> %s", p.Lconn.RemoteAddr().String(), p.rconn.RemoteAddr().String())
	//bidirectional copy
	go p.pipe(p.Lconn, p.rconn)
	go p.pipe(p.rconn, p.Lconn)
	//wait for close...
	<-p.Errsig
	p.log("Closed (%d bytes sent, %d bytes recieved)", p.sentBytes, p.receivedBytes)
}

func (p *Proxy) pipe(src, dst *net.TCPConn) {
	//data direction
	var f, h string
	isFromDebugger := src == p.Lconn
	if isFromDebugger {
		f = "\nDebugger >>> IDE\n================"
	} else {
		f = "\nIDE >>> Debugger\n================"
	}
	h = "%s"
	//directional copy (64k buffer)
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			p.err("Read failed '%s'\n", err)
			return
		}
		b := buff[:n]
		p.log(h, f)
		if veryverbose {
			if isFromDebugger {
				p.log("Raw protocol:\n%s\n", logger.Colorize(fmt.Sprintf(h, b), "blue"))
			} else {
				p.log("Raw protocol:\n%s\n", logger.Colorize(fmt.Sprintf(h, logger.FormatTextProtocol(b)), "blue"))
			}
		}
		//extract command name
		if isFromDebugger {
			b = pathmapper.ApplyMappingToXML(b)
		} else {
			b = pathmapper.ApplyMappingToTextProtocol(b)
		}
		//show output
		if veryverbose {
			if isFromDebugger {
				p.log("Processed protocol:\n%s\n", logger.Colorize(fmt.Sprintf(h, b), "blue"))
			} else {
				p.log("Processed protocol:\n%s\n", logger.Colorize(fmt.Sprintf(h, logger.FormatTextProtocol(b)), "blue"))
			}
		} else {
			p.log(h, "")
		}
		//write out result
		n, err = dst.Write(b)
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
		if isFromDebugger {
			p.sentBytes += uint64(n)
		} else {
			p.receivedBytes += uint64(n)
		}
	}
}
