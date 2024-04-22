package fd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ssst0n3/fd-listener/pkg"
)

type Stat struct {
	FdPath     string
	RealPath   string
	SocketPath string
	Leak       bool
	Flags      int64
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
