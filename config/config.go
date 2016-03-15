// Copyright 2015 Dominique Feyer <dfeyer@ttree.ch>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

// Config store the proxy configuration
type Config struct {
	Context     string
	Framework   string
	LocalRoot   string
	Verbose     bool
	VeryVerbose bool
	Debug       bool
}
