package process

import (
	"github.com/ssst0n3/fd-listener/pkg/event"
	"github.com/ssst0n3/fd-listener/pkg/util"
	"os"
	"strconv"
)

func (l *Listener) process() {
	lastPid, err := util.LastPid()
	if err != nil {
		return
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		if pid >= lastPid-10 {
			valid, err := l.filter(pid)
			if err != nil {
				continue
			}
			if valid {
				l.known.Store(pid, true)
				l.Event <- event.Event{
					Type: event.ProcessNew,
					Pid:  pid,
				}
			}
		}
	}
}
