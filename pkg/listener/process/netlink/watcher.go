package netlink

import (
	"github.com/ssst0n3/awesome_libs/awesome_error"
	"github.com/vishvananda/netlink"
	"os"
)

type Watcher struct {
	thread bool
	done   chan struct{}
	start  chan int
	exit   chan int
}

func New(thread bool) *Watcher {
	return &Watcher{
		thread: thread,
		done:   make(chan struct{}),
		start:  make(chan int),
		exit:   make(chan int),
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
	go func() {
		for {
			err := <-errChan
			if err != nil {
				//awesome_error.CheckErr(err)
			}
		}
	}()
	return
}

func (w Watcher) Init() (err error) {
	event := make(chan netlink.ProcEvent)
	err = w.init(event, w.done)
	if err != nil {
		return
	}
	go func() {
		for {
			e := <-event
			switch msg := e.Msg.(type) {
			case *netlink.ExecProcEvent:
				w.start <- int(msg.Tgid())
			case *netlink.ExitProcEvent:
				w.exit <- int(msg.Pid())
			}
		}
	}()
	return
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
		c <- <-w.start
	}
}

func (w Watcher) Exit(c chan int) (err error) {
	for {
		c <- <-w.exit
	}
}
