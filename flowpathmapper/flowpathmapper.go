// Copyright 2015 Dominique Feyer <dfeyer@ttree.ch>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flowpathmapper

import (
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/errorhandler"
	"github.com/dfeyer/flow-debugproxy/logger"
	"github.com/dfeyer/flow-debugproxy/pathmapperfactory"
	"github.com/dfeyer/flow-debugproxy/pathmapping"

	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	h                = "%s"
	framework        = "flow"
	cachePathPattern = "@base@/Data/Temporary/@context@/Cache/Code/Flow_Object_Classes/@filename@.php"
)

var (
	regexpPhpFile         = regexp.MustCompile(`(?://)?(/[^ ]*\.php)`)
	regexpFilename        = regexp.MustCompile(`filename=["]?file://(\S+)/Data/Temporary/.+?/Cache/Code/Flow_Object_Classes/([^"]*)\.php`)
	regexpPathAndFilename = regexp.MustCompile(`(?m)^# PathAndFilename: (.*)$`)
	regexpPackageClass    = regexp.MustCompile(`(.*?)/Packages/[^/]*/(.*?)/Classes/(.*).php`)
	regexpDot             = regexp.MustCompile(`[\./]`)
)

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
	return p.doTextPathMapping(message)
}

// ApplyMappingToXML change file path in xDebug XML protocol
func (p *PathMapper) ApplyMappingToXML(message []byte) []byte {
	message = p.doXMLPathMapping(message)

	// update xml length count
	s := strings.Split(string(message), "\x00")
	i, err := strconv.Atoi(s[0])
	errorhandler.PanicHandling(err, p.logger)
	l := len(s[1])
	if i != l {
		message = bytes.Replace(message, []byte(strconv.Itoa(i)), []byte(strconv.Itoa(l)), 1)
	}

	return message
}

func (p *PathMapper) doTextPathMapping(message []byte) []byte {
	var processedMapping = map[string]string{}
	for _, match := range regexpPhpFile.FindAllStringSubmatch(string(message), -1) {
		originalPath := match[1]
		path := p.mapPath(originalPath)
		p.logger.Debug("doTextPathMapping %s >>> %s", path, originalPath)
		processedMapping[path] = originalPath
	}

	for path, originalPath := range processedMapping {
		message = bytes.Replace(message, []byte(p.getRealFilename(originalPath)), []byte(p.getRealFilename(path)), -1)
	}

	return message
}

func (p *PathMapper) getCachePath(base, filename string) string {
	cachePath := strings.Replace(cachePathPattern, "@base@", base, 1)
	context := p.config.Context
	if strings.Contains(context, "/") {
		contextParts := strings.Split(context, "/")
		contextParts[1] = "SubContext" + contextParts[1]
		context = strings.Join(contextParts, "/")
	}
	cachePath = strings.Replace(cachePath, "@context@", context, 1)
	return strings.Replace(cachePath, "@filename@", filename, 1)
}

func (p *PathMapper) doXMLPathMapping(b []byte) []byte {
	var processedMapping = map[string]string{}
	for _, match := range regexpFilename.FindAllStringSubmatch(string(b), -1) {
		path := p.getCachePath(match[1], match[2])
		if _, ok := processedMapping[path]; ok == false {
			if originalPath, exist := p.pathMapping.Get(path); exist {
				if p.config.VeryVerbose {
					p.logger.Info("Umpa Lumpa can help you, he know the mapping\n%s\n%s\n", p.logger.Colorize(">>> "+fmt.Sprintf(h, path), "yellow"), p.logger.Colorize(">>> "+fmt.Sprintf(h, p.getRealFilename(originalPath)), "green"))
				}
				processedMapping[path] = originalPath
				p.logger.Debug("doXMLPathMapping mapping exist %s >>> %s", path, originalPath)
			} else {
				originalPath = p.readOriginalPathFromCache(path, match[1])
				processedMapping[path] = originalPath
				p.logger.Debug("doXMLPathMapping missing mapping %s >>> %s", path, originalPath)
			}
		}
	}

	for path, originalPath := range processedMapping {
		path = p.getRealFilename(path)
		originalPath = p.getRealFilename(originalPath)
		b = bytes.Replace(b, []byte(path), []byte(originalPath), -1)
	}

	return b
}

// getRealFilename removes file:// protocol from the given path
func (p *PathMapper) getRealFilename(path string) string {
	return strings.TrimPrefix(path, "file://")
}

func (p *PathMapper) mapPath(originalPath string) string {
	if strings.Contains(originalPath, "/Packages/") {
		p.logger.Debug("Path %s is a Flow Package file", originalPath)
		cachePath := p.getCachePath(p.buildClassNameFromPath(originalPath))
		realPath := p.getRealFilename(cachePath)
		var err error
		if len(p.config.LocalRoot) == 0 {
			_, err = os.Stat(realPath)
		}
		if err == nil {
			return p.setPathMapping(realPath, originalPath)
		}
	}

	return originalPath
}

func (p *PathMapper) setPathMapping(path string, originalPath string) string {
	if p.config.Verbose {
		p.logger.Info(h, "Our Umpa Lumpa take care of your mapping and they did a great job, they found a proxy for you:")
		p.logger.Info(">>> %s\n", path)
	}

	if p.pathMapping.Has(path) == false {
		p.pathMapping.Set(path, originalPath)
	}
	return path
}

func (p *PathMapper) readOriginalPathFromCache(path, basePath string) string {
	localPath := path
	if len(p.config.LocalRoot) > 0 {
		localPath = strings.Replace(path, basePath, p.config.LocalRoot, 1)
	}
	p.logger.Debug("readOriginalPathFromCache %s", localPath)
	dat, err := ioutil.ReadFile(localPath)
	errorhandler.PanicHandling(err, p.logger)
	match := regexpPathAndFilename.FindStringSubmatch(string(dat))
	if len(match) == 2 {
		originalPath := match[1]
		if len(p.config.LocalRoot) > 0 {
			originalPath = strings.Replace(strings.Replace(originalPath, "\\", "/", -1), p.config.LocalRoot, basePath, 1)
		}
		if p.config.VeryVerbose {
			p.logger.Info("Umpa Lumpa need to work harder, need to reverse this one\n>>> %s\n>>> %s\n", p.logger.Colorize(fmt.Sprintf(h, path), "yellow"), p.logger.Colorize(fmt.Sprintf(h, originalPath), "green"))
		}
		p.logger.Debug("readOriginalPathFromCache %s >>> %s", path, originalPath)
		p.setPathMapping(path, originalPath)
		return originalPath
	}
	return path
}

func (p *PathMapper) buildClassNameFromPath(path string) (string, string) {
	basePath, className := pathToClassPath(path)
	if className == "" {
		// Other (vendor) packages, todo add support for vendor package with Flow proxy class
		p.logger.Warn(h, "Vendor package detected")
		p.logger.Warn("Class mapping not supported currently for path: %s, \n", path)
	}
	return basePath, className
}

// Convert absolute path to class path (internal use only)
func pathToClassPath(path string) (string, string) {
	var (
		basePath  string
		classPath string
	)
	match := regexpPackageClass.FindStringSubmatch(path)
	if len(match) == 4 {
		// Flow standard packages
		packagePath := regexpDot.ReplaceAllString(match[2], "/")
		classPath = match[3]
		if strings.Contains(classPath, packagePath) == false {
			// Quick'n dirty PSR4 support
			classPath = packagePath + "/" + classPath
		}
		basePath = match[1]
		classPath = regexpDot.ReplaceAllString(classPath, "_")
	} else {
		// Other (vendor) packages, todo add support for vendor package with Flow proxy class
		basePath = path
		classPath = ""
	}
	return basePath, classPath
}
