package fd

import (
	"fmt"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/event"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/stat"
	"sort"
	"sync"
)

type Worker struct {
	pid     int
	stop    chan struct{}
	store   sync.Map
	max     int
	watcher Watcher
	event   chan event.Events
}

func NewWorker(pid int) (w *Worker) {
	w = &Worker{
		pid:   pid,
		stop:  make(chan struct{}),
		event: make(chan event.Events),
	}
	w.watcher = NewWatcher(pid)
	go w.Work()
	return
}

func (w *Worker) Work() {
	go w.watcher.Watch(w.event)
	for {
		select {
		case <-w.stop:
			return
		case e := <-w.event:
			w.do(e)
		}
	}
}

// Close and wait watcher close
func (w *Worker) Close() {
	// wait watcher closed first to prevent w.event block
	w.watcher.Close()
	w.stop <- struct{}{}
}

func (w *Worker) do(events event.Events) {
	if len(events) == 0 {
		return
	}
	var changed []*stat.Stat
	var closed []int
	for _, e := range events {
		switch e.Type {
		case event.Open, event.Change:
			e.Stat.Change(true)
			changed = append(changed, e.Stat)
			w.store.Store(e.Fd, e.Stat)
		case event.Close:
			closed = append(closed, e.Fd)
			w.store.Store(e.Fd, e.Stat)
		default:
			panic("unhandled default case")
		}
	}
	w.print()
	for _, s := range changed {
		s.Change(false)
	}
	for _, fd := range closed {
		w.store.Delete(fd)
	}
}

func (w *Worker) print() {
	var keys []int
	w.store.Range(func(key any, value any) bool {
		fd := key.(int)
		keys = append(keys, fd)
		return true
	})
	sort.Ints(keys)
	for _, fd := range keys {
		s, _ := w.store.Load(fd)
		fmt.Println(s)
	}
	fmt.Println("----------------")
}
