package main

import (
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/errorhandler"
	"github.com/dfeyer/flow-debugproxy/flowpathmapper"
	"github.com/dfeyer/flow-debugproxy/logger"
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
		context := cli.String("context")
		laddr, err := net.ResolveTCPAddr("tcp", localAddr)
		errorhandler.PanicHandling(err)
		raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
		errorhandler.PanicHandling(err)
		listener, err := net.ListenTCP("tcp", laddr)
		errorhandler.PanicHandling(err)

		logger.Info("Debugger from %v\nIDE      from %v\n", localAddr, remoteAddr)

		veryverbose := cli.Bool("vv")
		verbose := cli.Bool("verbose") || veryverbose

		config := &config.Config{
			Context:     context,
			Verbose:     verbose,
			VeryVerbose: veryverbose,
		}

		pathMapper := flowpathmapper.PathMapper{
			Config: config,
		}

		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				logger.Warn("Failed to accept connection '%s'\n", err)
				continue
			}

			proxy := &xdebugproxy.Proxy{
				Lconn:      conn,
				Raddr:      raddr,
				PathMapper: &pathMapper,
				Config:     config,
			}
			go proxy.Start()
		}
	}

	app.Run(os.Args)
}
