package process

import (
	"github.com/fatih/color"
	"github.com/ssst0n3/fd-listener/pkg/listener/process/netlink"
	"github.com/ssst0n3/fd-listener/pkg/listener/process/walk"
	"sync"
)

type Watcher interface {
	Init() (err error)
	Enable() (enabled bool)
	Start(c chan int) (err error)
	Exit(c chan int) (err error)
}

func NewWatcher(thread bool, step int, known *sync.Map) (w Watcher) {
	w = netlink.New(thread)
	enabled := w.Enable()
	color.White("[+] netlink (faster) enabled: %t\n", enabled)
	if enabled {
		return
	} else {
		w = walk.New(thread, step, known)
		return
	}
}

func Watch(w Watcher, start chan int, exit chan int) (err error) {
	err = w.Init()
	if err != nil {
		return
	}
	go func() {
		for {
			err := w.Start(start)
			if err != nil {
				break
			}
		}
	}()
	go func() {
		for {
			err := w.Exit(exit)
			if err != nil {
				break
			}
		}
	}()
	return
}
