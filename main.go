package main

import (
	"github.com/codegangsta/cli"
	"os"

	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/mgutz/ansi"

	"net"
)

var connid = uint64(0)
var mapping = map[string]string{}
var verbose = false
var veryverbose = false

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
		check(err)
		raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
		check(err)
		listener, err := net.ListenTCP("tcp", laddr)
		check(err)

		fmt.Printf(c("Debugger from %v\n", "green"), localAddr)
		fmt.Printf(c("IDE from %v\n", "green"), remoteAddr)

		verbose = cli.Bool("verbose")
		veryverbose = cli.Bool("vv")

		if veryverbose {
			verbose = true
		}

		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Printf("Failed to accept connection '%s'\n", err)
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
		log(s, args...)
	}
}

func (p *proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		warn(s, err)
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
	var f, h, command string
	isFromDebugger := src == p.lconn
	if isFromDebugger {
		f = "\nDebugger >>> IDE"
	} else {
		f = "\nIDE >>> Debugger"
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
		command = "Not really important for us ..."
		p.log(h, f)
		//extract command name
		if isFromDebugger {
			//todo, found a proper way to parse XML
		} else {
			commandParts := strings.Fields(fmt.Sprintf(h, b))
			command = commandParts[0]
			if command == "breakpoint_set" {
				file := commandParts[6]
				if verbose {
					p.log("Command: %s", c(command, "blue"))
				}
				fileMapping := mapPath(file)
				b = bytes.Replace(b, []byte(file), []byte(fileMapping), 1)
			}
		}
		//show output
		if veryverbose {
			p.log(h, "\n"+c(fmt.Sprintf(h, b), "blue"))
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

func buildClassNameFromPath(path string) []string {
	// todo add support for PSR4
	r := regexp.MustCompile(`(.*?)/Packages/(.*?)/Classes/(.*).php`)
	match := r.FindStringSubmatch(path)
	basePath := match[1]
	r = regexp.MustCompile(`[\./]`)
	className := r.ReplaceAllString(match[3], "_")
	return []string{basePath, className}
}

func mapPath(originalPath string) string {
	if strings.Contains(originalPath, "/Packages/") {
		parts := buildClassNameFromPath(originalPath)
		codeCacheFileName := parts[0] + "/Data/Temporary/Development/Cache/Code/Flow_Object_Classes/" + parts[1] + ".php"
		realCodeCacheFileName := getRealFilename(codeCacheFileName)
		if _, err := os.Stat(realCodeCacheFileName); err == nil {
			return registerPathMapping(realCodeCacheFileName, originalPath)
		}
	}

	return originalPath
}

func registerPathMapping(path string, originalPath string) string {
	dat, err := ioutil.ReadFile(path)
	check(err)
	//check if file contains flow annotation
	if strings.Contains(string(dat), "@Flow\\") {
		if verbose {
			log("%s", c("Our Umpa Lumpa take care of your mapping and they did a great job, they found a proxy for you:", "green"))
			log(">>> %s\n", c(path, "green"))
		}
		mapping[path] = originalPath
		return path
	}
	return originalPath
}

func getRealFilename(path string) string {
	return strings.TrimPrefix(path, "file://")
}

//helper functions

func check(err error) {
	if err != nil {
		warn(err.Error())
		os.Exit(1)
	}
}

func c(str, style string) string {
	return ansi.Color(str, style)
}

func log(f string, args ...interface{}) {
	fmt.Printf(c(f, "green")+"\n", args...)
}

func warn(f string, args ...interface{}) {
	fmt.Printf(c(f, "red")+"\n", args...)
}
