// +build freebsd netbsd darwin

package uptime

import "testing"

func TestParseMetrics(t *testing.T) {
	v, err := calcUptime("{ sec = 1455448176, usec = 0 } Sun Feb 14 20:09:36 2016\n", 1456242880)
	if err != nil {
		t.Errorf("error should be nil but: %s", err)
	}
	if v != float64(794704) {
		t.Errorf("uptime should be 794704.0 but: %f", v)
	}
}
