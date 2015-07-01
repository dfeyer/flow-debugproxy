package main

import (
	"github.com/dfeyer/flow-debugproxy/errorhandler"
	"github.com/dfeyer/flow-debugproxy/logger"
	"github.com/dfeyer/flow-debugproxy/pathmapper"

	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"net"
	"os"
)

var connid = uint64(0)
var verbose = false
var veryverbose = false
var h = "%s"

func main() {
	app := cli.NewApp()
	app.Name = "flow-debugproxy"
	app.Usage = "Flow Framework xDebug proxy"
	app.Author = "Dominique Feyer"
	app.Email = "dominique@neos.io"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "xdebug, l",
			Value: "127.0.0.1:9000",
			Usage: "Listen address IP and port number",
		},
		cli.StringFlag{
			Name:  "ide, I",
			Value: "127.0.0.1:9010",
			Usage: "Bind address IP and port number",
		},
		cli.StringFlag{
			Name:  "context, c",
			Value: "Development",
			Usage: "The context to run as",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Verbose",
		},
		cli.BoolFlag{
			Name:  "vv",
			Usage: "Very verbose",
		},
	}

	app.Action = func(cli *cli.Context) {
		localAddr := cli.String("xdebug")
		remoteAddr := cli.String("ide")
		laddr, err := net.ResolveTCPAddr("tcp", localAddr)
		errorhandler.Handling(err)
		raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
		errorhandler.Handling(err)
		listener, err := net.ListenTCP("tcp", laddr)
		errorhandler.Handling(err)

		logger.Info("Debugger from %v\nIDE      from %v\n", localAddr, remoteAddr)

		verbose = cli.Bool("verbose")
		veryverbose = cli.Bool("vv")

		if veryverbose {
			verbose = true
		}

		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				logger.Warn("Failed to accept connection '%s'\n", err)
				continue
			}
			connid++

			proxy := &proxy{
				lconn:  conn,
				laddr:  laddr,
				raddr:  raddr,
				erred:  false,
				errsig: make(chan bool),
			}
			go proxy.start()
		}
	}

	app.Run(os.Args)
}

//A proxy represents a pair of connections and their state
type proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	laddr, raddr  *net.TCPAddr
	lconn, rconn  *net.TCPConn
	erred         bool
	errsig        chan bool
}

func (p *proxy) log(s string, args ...interface{}) {
	if verbose {
		logger.Info(s, args...)
	}
}

func (p *proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		logger.Warn(s, err)
	}
	p.errsig <- true
	p.erred = true
}

func (p *proxy) start() {
	defer p.lconn.Close()
	//connect to remote
	rconn, err := net.DialTCP("tcp", nil, p.raddr)
	if err != nil {
		p.err("Remote connection failed: %s", err)
		return
	}
	p.rconn = rconn
	defer p.rconn.Close()
	//display both ends
	p.log("Opened %s >>> %s", p.lconn.RemoteAddr().String(), p.rconn.RemoteAddr().String())
	//bidirectional copy
	go p.pipe(p.lconn, p.rconn)
	go p.pipe(p.rconn, p.lconn)
	//wait for close...
	<-p.errsig
	p.log("Closed (%d bytes sent, %d bytes recieved)", p.sentBytes, p.receivedBytes)
}

func (p *proxy) pipe(src, dst *net.TCPConn) {
	//data direction
	var f, h string
	isFromDebugger := src == p.lconn
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
