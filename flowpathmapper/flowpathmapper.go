package flowpathmapper

import (
	"github.com/dfeyer/flow-debugproxy/config"
	"github.com/dfeyer/flow-debugproxy/errorhandler"
	"github.com/dfeyer/flow-debugproxy/logger"

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
	cachePathPattern = "@base@/Data/Temporary/@context@/Cache/Code/Flow_Object_Classes/@filename@.php"
)

var (
	mapping               = map[string]string{}
	regexpPhpFile         = regexp.MustCompile(`(?://)?(/[^ ]*\.php)`)
	regexpFilename        = regexp.MustCompile(`filename=["]?file://(\S+)/Data/Temporary/[^/]*/Cache/Code/Flow_Object_Classes/([^"]*)\.php`)
	regexpPathAndFilename = regexp.MustCompile(`(?m)^# PathAndFilename: (.*)$`)
	regexpPackageClass    = regexp.MustCompile(`(.*?)/Packages/(.*?)/Classes/(.*).php`)
	regexpDot             = regexp.MustCompile(`[\./]`)
)

// PathMapper handle the mapping between real code and proxy
type PathMapper struct {
	Config *config.Config
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
	errorhandler.PanicHandling(err)
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
		if _, ok := processedMapping[path]; ok == false {
			processedMapping[path] = originalPath
		}
	}

	for path, originalPath := range processedMapping {
		path = p.getRealFilename(path)
		originalPath = p.getRealFilename(originalPath)
		message = bytes.Replace(message, []byte(originalPath), []byte(path), -1)
	}

	return message
}

func (p *PathMapper) getCachePath(base, filename string) string {
	cachePath := strings.Replace(cachePathPattern, "@base", base, 1)
	cachePath = strings.Replace(cachePath, "@context", p.Config.Context, 1)
	return strings.Replace(cachePath, "@filename", filename, 1)
}

func (p *PathMapper) doXMLPathMapping(b []byte) []byte {
	var processedMapping = map[string]string{}
	for _, match := range regexpFilename.FindAllStringSubmatch(string(b), -1) {
		path := p.getCachePath(match[1], match[2])
		if _, ok := processedMapping[path]; ok == false {
			if originalPath, exist := mapping[path]; exist {
				if p.Config.VeryVerbose {
					logger.Info("Umpa Lumpa can help you, he know the mapping\n%s\n%s\n", logger.Colorize(">>> "+fmt.Sprintf(h, path), "yellow"), logger.Colorize(">>> "+fmt.Sprintf(h, p.getRealFilename(originalPath)), "green"))
				}
				processedMapping[path] = originalPath
			} else {
				originalPath = p.readOriginalPathFromCache(path)
				processedMapping[path] = originalPath
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
		parts := p.buildClassNameFromPath(originalPath)
		cachePath := p.getCachePath(parts[0], parts[1])
		realPath := p.getRealFilename(cachePath)
		if _, err := os.Stat(realPath); err == nil {
			return p.registerPathMapping(realPath, originalPath)
		}
	}

	return originalPath
}

func (p *PathMapper) registerPathMapping(path string, originalPath string) string {
	dat, err := ioutil.ReadFile(path)
	errorhandler.PanicHandling(err)
	// check if file contains flow annotation
	if strings.Contains(string(dat), "@Flow\\") {
		if p.Config.Verbose {
			logger.Info("%s", "Our Umpa Lumpa take care of your mapping and they did a great job, they found a proxy for you:")
			logger.Info(">>> %s\n", path)
		}

		if _, exist := mapping[path]; exist == false {
			mapping[path] = originalPath
		}
		return path
	}
	return originalPath
}

func (p *PathMapper) readOriginalPathFromCache(path string) string {
	dat, err := ioutil.ReadFile(path)
	errorhandler.PanicHandling(err)
	match := regexpPathAndFilename.FindStringSubmatch(string(dat))
	if len(match) == 2 {
		originalPath := match[1]
		if p.Config.VeryVerbose {
			logger.Info("Umpa Lumpa need to work harder, need to reverse this one\n>>> %s\n>>> %s\n", logger.Colorize(fmt.Sprintf(h, path), "yellow"), logger.Colorize(fmt.Sprintf(h, originalPath), "green"))
		}
		p.registerPathMapping(path, originalPath)
		return originalPath
	}
	return path
}

func (p *PathMapper) buildClassNameFromPath(path string) []string {
	// todo add support for PSR4
	match := regexpPackageClass.FindStringSubmatch(path)
	basePath := match[1]
	className := regexpDot.ReplaceAllString(match[3], "_")
	return []string{basePath, className}
}
