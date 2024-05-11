package stat

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ssst0n3/fd-listener/pkg"
	"os"
)

type Stat struct {
	FdPath     string
	RealPath   string
	SocketPath string
	Leak       bool
	Flags      int64
}

func New(pid, fd int) (stat Stat, err error) {
	fdPath := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
	realPath, _ := os.Readlink(fdPath)
	if realPath == "" {
		realPath = "?"
	}
	stat = Stat{
		FdPath:   fdPath,
		RealPath: realPath,
	}
	socketPath, err := Socket(pid, realPath)
	if err != nil {
		return
	}
	stat.SocketPath = socketPath
	leak, err := Leak(fdPath, realPath)
	if err != nil {
		return
	}
	stat.Leak = leak
	flags, err := pkg.ReadFlags(fmt.Sprintf("/proc/%d/fdinfo/%d", pid, fd))
	if err != nil {
		return
	}
	stat.Flags = flags
	return
}

func (s Stat) String() (content string) {
	leaked := ""
	if s.Leak {
		leaked = "leaked!"
	}
	flags := pkg.ParseFlags(s.Flags)
	var socketPath string
	if s.SocketPath != "" {
		socketPath = " -> " + s.SocketPath
	}
	content = fmt.Sprintf("%s -> %s%s\t; %s\t%s", s.FdPath, s.RealPath, socketPath, leaked, flags)
	if s.Leak {
		content = color.RedString(content)
	}
	return
}
