package uptime

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func get() (float64, error) {
	contentbytes, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return 0.0, fmt.Errorf("Faild to fetch uptime metrics: %s", err)
	}
	return calcUptime(string(contentbytes))
}

func calcUptime(str string) (float64, error) {
	cols := strings.Split(str, " ")
	f, err := strconv.ParseFloat(cols[0], 64)
	if err != nil {
		return 0.0, fmt.Errorf("Faild to fetch uptime metrics: %s", err)
	}
	return f, nil
}
