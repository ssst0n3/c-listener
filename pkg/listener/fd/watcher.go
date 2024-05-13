package fd

import (
	"github.com/fatih/color"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/ebpf"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/event"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/walk"
)

type Watcher interface {
	Watch(e chan<- event.Events)
	Enable() (enabled bool)
	Close()
}

func NewWatcher(pid int) (watcher Watcher) {
	watcher = ebpf.New(pid)
	enabled := watcher.Enable()
	color.White("[+] ebpf (faster) enabled: %t\n", enabled)
	if enabled {
		return
	} else {
		watcher = walk.New(pid)
		return
	}
}

func Watch(watcher Watcher, events chan<- event.Events) {
	watcher.Watch(events)
}
