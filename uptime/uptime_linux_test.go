package uptime

import (
	"testing"
)

func TestCalcUptime(t *testing.T) {
	v, err := calcUptime("481453.56 1437723.27\n")
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
	if v != float64(481453.56) {
		t.Errorf("uptime should be 481453.56, but: %f", v)
	}
}
