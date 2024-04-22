package fd

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

func Leak(fdPath, realPath string) (leak bool, err error) {
	fdFI, err := os.Stat(fdPath)
	if err != nil {
		return
	}
	fdStat, _ := fdFI.Sys().(*syscall.Stat_t)
	realFI, err := os.Stat(realPath)
	if err != nil {
		if os.IsNotExist(err) {
			// stat anon_inode:[eventpoll]: no such file or directory
			err = nil
		}
		return
	}
	realStat, _ := realFI.Sys().(*syscall.Stat_t)
	if fdStat.Ino == realStat.Ino {
		leak = true
	}
	return
}

func Socket(pid int, realPath string) (socketPath string, err error) {
	if !strings.Contains(realPath, "socket:[") {
		return
	}
	id := strings.TrimSuffix(strings.TrimPrefix(realPath, "socket:["), "]")
	content, err := os.ReadFile(fmt.Sprintf("/proc/%d/net/unix", pid))
	if err != nil {
		return
	}
	socket := strings.Split(string(content), "\n")
	for _, line := range socket {
		if strings.Contains(line, id) {
			data := strings.Split(line, " ")
			socketPath = data[len(data)-1]
			return
		}
	}
	return
}
