package main

import (
	"github.com/dfeyer/flow-debugproxy/errorhandler"
	"github.com/dfeyer/flow-debugproxy/logger"

	"github.com/codegangsta/cli"
	"os"

	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"net"
)

var connid = uint64(0)
var mapping = map[string]string{}
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
				p.log("Raw protocol:\n%s\n", logger.Colorize(fmt.Sprintf(h, debugTextProtocol(b)), "blue"))
			}
		}
		//extract command name
		if isFromDebugger {
			b = applyMappingToXML(b)
		} else {
			b = applyMappingToTextProtocol(b)
		}
		//show output
		if veryverbose {
			if isFromDebugger {
				p.log("Processed protocol:\n%s\n", logger.Colorize(fmt.Sprintf(h, b), "blue"))
			} else {
				p.log("Processed protocol:\n%s\n", logger.Colorize(fmt.Sprintf(h, debugTextProtocol(b)), "blue"))
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

func debugTextProtocol(protocol []byte) []byte {
	return bytes.Trim(bytes.Replace(protocol, []byte("\x00"), []byte("\n"), -1), "\n")
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

func applyMappingToTextProtocol(protocol []byte) []byte {
	commandParts := strings.Fields(fmt.Sprintf(h, protocol))
	command := commandParts[0]
	if command == "breakpoint_set" {
		file := commandParts[6]
		if verbose {
			logger.Info("Command: %s", logger.Colorize(command, "blue"))
		}
		fileMapping := mapPath(file)
		protocol = bytes.Replace(protocol, []byte(file), []byte("file://"+fileMapping), 1)
	}

	return protocol
}

func applyMappingToXML(xml []byte) []byte {
	r := regexp.MustCompile(`filename=["]?file://(\S+)/Data/Temporary/Development/Cache/Code/Flow_Object_Classes/([^"]*)\.php`)
	var processedMapping = map[string]string{}

	for _, match := range r.FindAllStringSubmatch(string(xml), -1) {
		path := match[1] + "/Data/Temporary/Development/Cache/Code/Flow_Object_Classes/" + match[2] + ".php"
		if _, ok := processedMapping[path]; ok == false {
			if originalPath, exist := mapping[path]; exist {
				if veryverbose {
					logger.Info("Umpa Lumpa can help you, he know the mapping\n%s\n%s\n", logger.Colorize(">>> "+fmt.Sprintf(h, path), "yellow"), logger.Colorize(">>> "+fmt.Sprintf(h, getRealFilename(originalPath)), "green"))
				}
				processedMapping[path] = originalPath
			} else {
				originalPath = readOriginalPathFromCache(path)
				processedMapping[path] = originalPath
			}
		}
	}

	for path, originalPath := range processedMapping {
		path = getRealFilename(path)
		originalPath = getRealFilename(originalPath)
		xml = bytes.Replace(xml, []byte(path), []byte(originalPath), -1)
	}
	s := strings.Split(string(xml), "\x00")
	i, err := strconv.Atoi(s[0])
	errorhandler.Handling(err)
	l := len(s[1])
	if i != l {
		xml = bytes.Replace(xml, []byte(strconv.Itoa(i)), []byte(strconv.Itoa(l)), 1)
	}

	return xml
}

func readOriginalPathFromCache(path string) string {
	dat, err := ioutil.ReadFile(path)
	errorhandler.Handling(err)
	r := regexp.MustCompile(`(?m)^# PathAndFilename: (.*)$`)
	match := r.FindStringSubmatch(string(dat))
	//todo check if the match contain something
	originalPath := match[1]
	if veryverbose {
		logger.Info("Umpa Lumpa need to work harder, need to reverse this one\n>>> %s\n>>> %s\n", logger.Colorize(fmt.Sprintf(h, path), "yellow"), logger.Colorize(fmt.Sprintf(h, originalPath), "green"))
	}
	registerPathMapping(path, originalPath)
	return originalPath
}

func registerPathMapping(path string, originalPath string) string {
	dat, err := ioutil.ReadFile(path)
	errorhandler.Handling(err)
	//check if file contains flow annotation
	if strings.Contains(string(dat), "@Flow\\") {
		if verbose {
			logger.Info("%s", "Our Umpa Lumpa take care of your mapping and they did a great job, they found a proxy for you:")
			logger.Info(">>> %s\n", path, "green")
		}

		if _, exist := mapping[path]; exist == false {
			mapping[path] = originalPath
		}
		return path
	}
	return originalPath
}

func getRealFilename(path string) string {
	return strings.TrimPrefix(path, "file://")
}
