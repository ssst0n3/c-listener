package ebpf

import (
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/event"
)

type Watcher struct {
	msg chan Message
}

func New(pid int) *Watcher {
	return &Watcher{
		msg: make(chan Message),
	}
}

func (w Watcher) Init() {
	// register
	// receive IPC message
}

func (w Watcher) Enable() (enabled bool) {
	return
}

func (w Watcher) Watch(stop <-chan struct{}, event chan<- event.Events) {
	select {
	case <-stop:
		return
	default:
		w.do(event)
	}
}

func (w Watcher) do(e chan<- event.Events) {
	for {
		msg := <-w.msg
		events, err := msg.parse()
		if err != nil {
			continue
		}
		e <- events
	}
}
