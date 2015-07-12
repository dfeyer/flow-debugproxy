// Copyright 2015 Dominique Feyer <dfeyer@ttree.ch>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathmapping

var mapping = map[string]string{}

// PathMapping is a simple key store for class and proxy class mapping
type PathMapping struct{}

// Set a path mapping
func (p *PathMapping) Set(path string, originalPath string) {
	mapping[path] = originalPath
}

// Get a path mapping
func (p *PathMapping) Get(path string) (string, bool) {
	if p.Has(path) {
		return mapping[path], true
	}
	return "", false
}

// Has check if the path mapping exist
func (p *PathMapping) Has(path string) bool {
	if _, exist := mapping[path]; exist {
		return true
	}
	return false
}
