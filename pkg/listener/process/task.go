package process

import (
	"github.com/ssst0n3/fd-listener/pkg/event"
	"github.com/ssst0n3/fd-listener/pkg/util"
)

func (l *Listener) task() {
	lastPid, err := util.LastPid()
	if err != nil {
		return
	}
	for pid := lastPid; pid < lastPid+l.stepLength; pid++ {
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
