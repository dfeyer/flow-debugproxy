package flowpathmapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildClassNameFromPathSupportPSR2(t *testing.T) {
	basePath, className := pathToClassPath("/your/path/sites/dev/master-dev.neos-workplace.dev/Packages/Application/Ttree.FlowDebugProxyHelper/Classes/Ttree/FlowDebugProxyHelper/ProxyClassMapperComponent.php")
	assert.Equal(t, "/your/path/sites/dev/master-dev.neos-workplace.dev", basePath, "they should be equal")
	assert.Equal(t, "Ttree_FlowDebugProxyHelper_ProxyClassMapperComponent", className, "they should be equal")
}

func TestBuildClassNameFromPathSupportPSR4(t *testing.T) {
	basePath, className := pathToClassPath("/your/path/sites/dev/master-dev.neos-workplace.dev/Packages/Application/Ttree.FlowDebugProxyHelper/Classes/ProxyClassMapperComponent.php")
	assert.Equal(t, "/your/path/sites/dev/master-dev.neos-workplace.dev", basePath, "they should be equal")
	assert.Equal(t, "Ttree_FlowDebugProxyHelper_ProxyClassMapperComponent", className, "they should be equal")
}
