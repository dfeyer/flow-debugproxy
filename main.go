package main

import (
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/errorhandler"
	"github.com/dfeyer/flow-debugproxy/flowstacktraceprocessor"
	"github.com/dfeyer/flow-debugproxy/logger"
	"github.com/dfeyer/flow-debugproxy/pathmapperfactory"
	"github.com/dfeyer/flow-debugproxy/pathmapping"
	"github.com/dfeyer/flow-debugproxy/xdebugproxy"

	"github.com/codegangsta/cli"

	"net"
	"os"
)

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
		cli.StringFlag{
			Name:  "framework",
			Value: "flow",
			Usage: "Framework support, currently on Flow framework is supported",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Verbose",
		},
		cli.BoolFlag{
			Name:  "vv",
			Usage: "Very verbose",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Show debug output",
		},
	}

	app.Action = func(cli *cli.Context) {
		c := &config.Config{
			Context:     cli.String("context"),
			Framework:   cli.String("framework"),
			Verbose:     cli.Bool("verbose") || cli.Bool("vv"),
			VeryVerbose: cli.Bool("vv"),
			Debug:       cli.Bool("debug"),
		}

		log := &logger.Logger{
			Config: c,
		}

		laddr, raddr, listener := setupNetworkConnection(cli.String("xdebug"), cli.String("ide"), log)

		log.Info("Debugger from %v\nIDE      from %v\n", laddr, raddr)

		pathMapping := &pathmapping.PathMapping{}
		pathMapper, err := pathmapperfactory.Create(c, pathMapping, log)
		errorhandler.PanicHandling(err, log)

		// Replace this by a plugin registry, with automatic registration
		stacktraceProcessor := &flowstacktraceprocessor.Processor{
			Config:      c,
			Logger:      log,
			PathMapping: pathMapping,
		}

		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				log.Warn("Failed to accept connection '%s'\n", err)
				continue
			}

			proxy := &xdebugproxy.Proxy{
				Lconn:      conn,
				Raddr:      raddr,
				PathMapper: pathMapper,
				Config:     c,
			}
			proxy.RegisterPostProcessor(stacktraceProcessor)
			go proxy.Start()
		}
	}

	app.Run(os.Args)
}

func setupNetworkConnection(xdebugAddr string, ideAddr string, log *logger.Logger) (*net.TCPAddr, *net.TCPAddr, *net.TCPListener) {
	laddr, err := net.ResolveTCPAddr("tcp", xdebugAddr)
	errorhandler.PanicHandling(err, log)

	raddr, err := net.ResolveTCPAddr("tcp", ideAddr)
	errorhandler.PanicHandling(err, log)

	listener, err := net.ListenTCP("tcp", laddr)
	errorhandler.PanicHandling(err, log)

	return laddr, raddr, listener
}
