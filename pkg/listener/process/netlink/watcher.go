package netlink

import (
	"github.com/ssst0n3/awesome_libs/awesome_error"
	"github.com/vishvananda/netlink"
	"os"
)

type Watcher struct {
	thread bool
	done   chan struct{}
	event  chan netlink.ProcEvent
}

func New(thread bool) *Watcher {
	return &Watcher{
		thread: thread,
		done:   make(chan struct{}),
		event:  make(chan netlink.ProcEvent),
	}
}

func (w Watcher) init(event chan netlink.ProcEvent, done chan struct{}) (err error) {
	errChan := make(chan error)
	err = netlink.ProcEventMonitor(event, done, errChan)
	if err != nil {
		if os.IsPermission(err) {
			return
		}
		awesome_error.CheckErr(err)
		return
	}
	return
}

func (w Watcher) Init() (err error) {
	return w.init(w.event, w.done)
}

func (w Watcher) Enable() (enabled bool) {
	event := make(chan netlink.ProcEvent)
	done := make(chan struct{})
	defer close(done)
	err := w.init(event, done)
	if err != nil {
		if os.IsPermission(err) {
			return
		}
		awesome_error.CheckErr(err)
		return
	}
	return true
}

func (w Watcher) Start(c chan int) (err error) {
	for {
		e := <-w.event
		switch msg := e.Msg.(type) {
		case *netlink.ForkProcEvent:
			if w.thread {
				c <- int(msg.ChildPid)
			}
		case *netlink.ExecProcEvent:
			if !w.thread {
				c <- int(msg.Pid())
			}
		}
	}
}

func (w Watcher) Exit(c chan int) (err error) {
	for {
		e := <-w.event
		switch msg := e.Msg.(type) {
		case *netlink.ExitProcEvent:
			if !w.thread {
				c <- int(msg.Pid())
			}
		}
	}
}
