package fd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ssst0n3/fd-listener/pkg"
)

type Stat struct {
	FdPath   string
	RealPath string
	Leak     bool
	Flags    int64
}

func (s Stat) String() (content string) {
	leaked := ""
	if s.Leak {
		leaked = "leaked!"
	}
	flags := pkg.ParseFlags(s.Flags)
	content = fmt.Sprintf("%s -> %s\t; %s\t%s", s.FdPath, s.RealPath, leaked, flags)
	if s.Leak {
		content = color.RedString(content)
	}
	return
}
