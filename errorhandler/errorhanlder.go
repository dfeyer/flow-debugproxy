// Copyright 2015 Dominique Feyer <dfeyer@ttree.ch>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errorhandler

import (
	"github.com/dfeyer/flow-debugproxy/logger"

	log "github.com/Sirupsen/logrus"
	"os"
)

// PanicHandling handle error and output log message
func PanicHandling(err error, logger *logger.Logger) {
	if err != nil {
		logger.Warn(err.Error())
		os.Exit(1)
	}
}

// PanicHandler handle error and output log message
func PanicHandler(err error) {
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
}
