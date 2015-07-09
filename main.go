package main

import (
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/errorhandler"
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
		context := cli.String("context")
		framework := cli.String("framework")
		debug := cli.Bool("debug")
		veryverbose := cli.Bool("vv")
		verbose := cli.Bool("verbose") || veryverbose

		c := &config.Config{
			Context:     context,
			Framework:   framework,
			Verbose:     verbose,
			VeryVerbose: veryverbose,
			Debug:       debug,
		}

		log := &logger.Logger{
			Config: c,
		}

		localAddr := cli.String("xdebug")
		remoteAddr := cli.String("ide")

		laddr, err := net.ResolveTCPAddr("tcp", localAddr)
		errorhandler.PanicHandling(err, log)
		raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
		errorhandler.PanicHandling(err, log)
		listener, err := net.ListenTCP("tcp", laddr)
		errorhandler.PanicHandling(err, log)

		log.Info("Debugger from %v\nIDE      from %v\n", localAddr, remoteAddr)

		pathMapping := &pathmapping.PathMapping{}
		pathMapper, err := pathmapperfactory.Create(c, pathMapping, log)
		errorhandler.PanicHandling(err, log)

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
			go proxy.Start()
		}
	}

	app.Run(os.Args)
}
