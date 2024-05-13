package fd

import (
	"github.com/ssst0n3/fd-listener/pkg/event"
)

type Listener struct {
	workers map[int]*Worker
	event   chan event.Event
}

func New(event chan event.Event) (l *Listener) {
	return &Listener{
		workers: make(map[int]*Worker),
		event:   event,
	}
}

func (l *Listener) Handle() {
	for {
		e := <-(l.event)
		switch e.Type {
		case event.ProcessNew:
			l.workers[e.Pid] = NewWorker(e.Pid)
		case event.ProcessExit:
			worker := l.workers[e.Pid]
			if worker != nil {
				worker.Close()
				delete(l.workers, e.Pid)
			}
		default:
			panic("unhandled default case")
		}
	}
}
