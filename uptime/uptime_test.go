package uptime

import (
	"testing"
)

func TestGet(t *testing.T) {
	v, err := Get()
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
	if v <= 0.0 {
		t.Errorf("uptime should be positive but got: %f", v)
	}
	t.Logf("uptime: %f", v)
}
