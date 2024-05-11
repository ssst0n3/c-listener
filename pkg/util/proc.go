package util

import (
	"github.com/ssst0n3/awesome_libs/awesome_error"
	"os"
	"strconv"
	"strings"
)

func LastPid() (lastPid int, err error) {
	content, err := os.ReadFile("/proc/sys/kernel/ns_last_pid")
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	pid, err := strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64)
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	lastPid = int(pid)
	return
}

func NotAlive(err error) bool {
	return os.IsNotExist(err) || strings.Contains(err.Error(), "no such process")
}
