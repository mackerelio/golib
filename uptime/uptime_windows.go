package uptime

import "syscall"

var (
	getTickCount = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount")
)

func get() (float64, error) {
	r, _, err := getTickCount.Call()
	if errno, ok := err.(syscall.Errno); ok {
		if errno != 0 {
			return 0, err
		}
		return float64(r) / 1000, nil
	}
	return 0, err
}
