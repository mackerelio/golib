package uptime

import "syscall"

var getTickCount = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount")

func get() (float64, error) {
	r, _, err := getTickCount.Call()
	return float64(r), err
}
