package pkg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var (
	FlagsDescriptions = []struct {
		flag int
		desc string
	}{
		{flag: os.O_RDONLY, desc: "O_RDONLY"},
		{flag: os.O_WRONLY, desc: "O_WRONLY"},
		{flag: os.O_RDWR, desc: "O_RDWR"},
		{flag: os.O_APPEND, desc: "O_APPEND"},
		{flag: syscall.O_DSYNC, desc: "O_DSYNC"},
		{flag: syscall.O_RSYNC, desc: "O_RSYNC"},
		{flag: syscall.O_SYNC, desc: "O_SYNC"},
		{flag: syscall.O_NDELAY, desc: "O_NDELAY"},
		{flag: syscall.O_NONBLOCK, desc: "O_NONBLOCK"},
		{flag: syscall.O_ASYNC, desc: "O_ASYNC"},
		{flag: syscall.O_CLOEXEC, desc: "O_CLOEXEC"},
		{flag: syscall.O_DIRECT, desc: "O_DIRECT"},
		{flag: syscall.O_NOATIME, desc: "O_NOATIME"},
		//{flag: syscall.O_PATH, desc: "O_PATH"},
	}
)

func ParseFlags(flagValue int64) string {
	var setFlags []string
	for _, flag := range FlagsDescriptions {
		if int(flagValue)&flag.flag == flag.flag {
			setFlags = append(setFlags, flag.desc)
		}
	}

	if len(setFlags) == 0 {
		return "No flags set"
	}

	return strings.Join(setFlags, ", ")
}

func ReadFlags(path string) (flags int64, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "flags:") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				err = fmt.Errorf("unexpected format for flags")
				return
			}
			flags, err = strconv.ParseInt(parts[1], 8, 64)
			if err != nil {
				return
			}
			return
		}
	}
	return
}
