package pluginutil

import (
	"os"
	"testing"
)

func TestGenerateTempfilePathWithBase(t *testing.T) {
	tempDir := os.TempDir()
	tests := []struct {
		name string
		env  string
		s    string
	}{
		{"empty", "", tempDir},
		{"specified", "/SOME-SPECIAL-PATH", "/SOME-SPECIAL-PATH"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origDir := os.Getenv("MACKEREL_PLUGIN_WORKDIR")
			os.Setenv("MACKEREL_PLUGIN_WORKDIR", tt.env) // nolint
			t.Cleanup(func() {
				os.Setenv("MACKEREL_PLUGIN_WORKDIR", origDir) // nolint
			})

			defaultPath := PluginWorkDir()
			if defaultPath != tt.s {
				t.Errorf("PluginWorkDir() should be %s, but: %s", tt.s, defaultPath)
			}
		})
	}
}
