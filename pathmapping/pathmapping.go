package pathmapping

var mapping = map[string]string{}

// PathMapping is a simple key store for class and proxy class mapping
type PathMapping struct{}

func (p *PathMapping) Set(path string, originalPath string) {
	mapping[path] = originalPath
}

func (p *PathMapping) Get(path string) (string, bool) {
	if p.Has(path) {
		return mapping[path], true
	}
	return "", false
}

func (p *PathMapping) Has(path string) bool {
	if _, exist := mapping[path]; exist {
		return true
	}
	return false
}
