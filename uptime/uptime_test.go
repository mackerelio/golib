package uptime

import (
	"testing"
)

func TestGet(t *testing.T) {
	v, err := Get()
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
	t.Logf("uptime: %f", v)
}
