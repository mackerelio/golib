package pluginutil

import (
	"os"
)

// PluginWorkDir returns path for mackerel-agent plugins' cache / tempfiles
func PluginWorkDir() string {
	dir := os.Getenv("MACKEREL_PLUGIN_WORKDIR")
	if dir == "" {
		dir = os.TempDir()
	}
	return dir
}
