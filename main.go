// Copyright 2015 Dominique Feyer <dfeyer@ttree.ch>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/dfeyer/flow-debugproxy/config"

	"github.com/dfeyer/flow-debugproxy/errorhandler"
	"github.com/dfeyer/flow-debugproxy/logger"
	"github.com/dfeyer/flow-debugproxy/pathmapperfactory"
	"github.com/dfeyer/flow-debugproxy/pathmapping"
	"github.com/dfeyer/flow-debugproxy/xdebugproxy"

	// Register available path mapper
	_ "github.com/dfeyer/flow-debugproxy/dummypathmapper"
	_ "github.com/dfeyer/flow-debugproxy/flowpathmapper"

	log "github.com/Sirupsen/logrus"
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
	app.Version = "0.9.0"

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
			Usage: "Framework support, currently on Flow framework (flow) or Dummy (dummy) is supported",
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

		logger := &logger.Logger{
			Config: c,
		}

		laddr, raddr, listener := setupNetworkConnection(cli.String("xdebug"), cli.String("ide"), logger)

		log.WithFields(log.Fields{
			"debugger": laddr,
			"ide":      raddr,
		}).Info("Flow Debug Proxy started")

		pathMapping := &pathmapping.PathMapping{}
		pathMapper, err := pathmapperfactory.Create(c, pathMapping)
		errorhandler.PanicHandling(err, logger)

		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				logger.Warn("Failed to accept connection '%s'\n", err)
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

func setupNetworkConnection(xdebugAddr string, ideAddr string, logger *logger.Logger) (*net.TCPAddr, *net.TCPAddr, *net.TCPListener) {
	laddr, err := net.ResolveTCPAddr("tcp", xdebugAddr)
	errorhandler.PanicHandling(err, logger)

	raddr, err := net.ResolveTCPAddr("tcp", ideAddr)
	errorhandler.PanicHandling(err, logger)

	listener, err := net.ListenTCP("tcp", laddr)
	errorhandler.PanicHandling(err, logger)

	return laddr, raddr, listener
}
