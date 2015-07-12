// Copyright 2015 Dominique Feyer <dfeyer@ttree.ch>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dummypathmapper

import (
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/logger"
	"github.com/dfeyer/flow-debugproxy/pathmapperfactory"
	"github.com/dfeyer/flow-debugproxy/pathmapping"
)

const framework = "dummy"

func init() {
	p := &PathMapper{}
	pathmapperfactory.Register(framework, p)
}

// PathMapper handle the mapping between real code and proxy
type PathMapper struct {
	config      *config.Config
	logger      *logger.Logger
	pathMapping *pathmapping.PathMapping
}

// Initialize the path mapper dependencies
func (p *PathMapper) Initialize(c *config.Config, l *logger.Logger, m *pathmapping.PathMapping) {
	p.config = c
	p.logger = l
	p.pathMapping = m
}

// ApplyMappingToTextProtocol change file path in xDebug text protocol
func (p *PathMapper) ApplyMappingToTextProtocol(message []byte) []byte {
	return message
}

// ApplyMappingToXML change file path in xDebug XML protocol
func (p *PathMapper) ApplyMappingToXML(message []byte) []byte {
	return message
}
