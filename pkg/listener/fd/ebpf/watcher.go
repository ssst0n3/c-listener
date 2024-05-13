package ebpf

import (
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/event"
)

type Watcher struct {
	msg  chan Message
	stop chan struct{}
}

func New(pid int) *Watcher {
	return &Watcher{
		msg: make(chan Message),
	}
}

func (w Watcher) Init() {
	w.register()
	go w.receive()
}

func (w Watcher) Enable() (enabled bool) {
	return
}

// register to kernel
func (w Watcher) register() {
	// register with os.GetPid()
}

// receive IPC message
func (w Watcher) receive() {
	for {
		w.msg <- Message{}
	}
}

func (w Watcher) Watch(event chan<- event.Events) {
	select {
	case <-w.stop:
		return
	default:
		w.do(event)
	}
}

func (w Watcher) Close() {
	w.stop <- struct{}{}
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
