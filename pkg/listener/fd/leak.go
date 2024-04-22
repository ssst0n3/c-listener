package fd

import (
	"os"
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
