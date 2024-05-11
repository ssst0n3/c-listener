package netlink2

import (
	"fmt"
	"github.com/slimtoolkit/slim/pkg/pdiscover"
	"github.com/ssst0n3/awesome_libs/awesome_error"
	"github.com/ssst0n3/fd-listener/pkg/util"
	"log"
	"os"
)

type Watcher struct {
	thread bool
	exec   chan *pdiscover.ProcEventExec
	fork   chan *pdiscover.ProcEventFork
	exit   chan *pdiscover.ProcEventExit
}

func New(thread bool) *Watcher {
	return &Watcher{
		thread: thread,
		exec:   make(chan *pdiscover.ProcEventExec, 1000),
		fork:   make(chan *pdiscover.ProcEventFork, 1000),
		exit:   make(chan *pdiscover.ProcEventExit, 1000),
	}
}

func (w Watcher) init() (watcher *pdiscover.Watcher, err error) {
	watcher, err = pdiscover.NewAllWatcher(pdiscover.PROC_EVENT_ALL)
	if err != nil {
		awesome_error.CheckErr(err)
		return
	}
	return
}

func (w Watcher) Init() (err error) {
	watcher, err := w.init()
	if err != nil {
		return
	}
	go func() {
		for {
			select {
			case <-watcher.Fork:
				continue
			case ev := <-watcher.Exec:
				w.exec <- ev
			case ev := <-watcher.Exit:
				w.exit <- ev
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	return
}

func (w Watcher) Enable() (enabled bool) {
	listener, err := w.init()
	if err != nil {
		return
	}
	err = listener.Close()
	if err != nil {
		return
	}
	return true
}

func (w Watcher) Start(c chan int) (err error) {
	for {
		e := <-w.exec
		c <- e.Pid
	}
}

func (w Watcher) Exit(c chan int) (err error) {
	for {
		e := <-w.exit
		pid := e.Pid
		_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
		if err != nil && util.NotAlive(err) {
			c <- pid
		} else {
			fmt.Printf("[!] pid %d alive\n", pid)
		}
	}
}
