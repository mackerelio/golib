// +build freebsd netbsd darwin

package uptime

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func get() (float64, error) {
	out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err != nil {
		return 0.0, fmt.Errorf("faild to fetch uptime: %s", err)
	}
	return calcUptime(string(out), time.Now().Unix())
}

func calcUptime(str string, nowEpoch int64) (float64, error) {
	// { sec = 1455448176, usec = 0 } Sun Feb 14 20:09:36 2016
	cols := strings.Split(str, " ")
	epoch, err := strconv.ParseInt(strings.Trim(cols[3], ","), 10, 64)
	if err != nil {
		return 0.0, fmt.Errorf("Faild to parse uptime: %s", err)
	}
	return float64(nowEpoch - epoch), nil
}
