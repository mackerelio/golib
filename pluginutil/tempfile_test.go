package pluginutil

import (
	"os"
	"testing"
)

func TestGenerateTempfilePathWithBase(t *testing.T) {
	origDir := os.Getenv("MACKEREL_PLUGIN_WORKDIR")
	os.Setenv("MACKEREL_PLUGIN_WORKDIR", "")
	defer os.Setenv("MACKEREL_PLUGIN_WORKDIR", origDir)

	expect1 := os.TempDir()
	defaultPath := PluginWorkDir()
	if defaultPath != expect1 {
		t.Errorf("PluginWorkDir() should be %s, but: %s", expect1, defaultPath)
	}

	os.Setenv("MACKEREL_PLUGIN_WORKDIR", "/SOME-SPECIAL-PATH")

	expect2 := "/SOME-SPECIAL-PATH"
	pathFromEnv := PluginWorkDir()
	if pathFromEnv != expect2 {
		t.Errorf("PluginWorkDir() should be %s, but: %s", expect2, pathFromEnv)
	}
}
