package walk

import (
	"fmt"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/event"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/stat"
	"os"
	"sort"
	"strconv"
	"sync"
)

type Watcher struct {
	store   *sync.Map
	changed bool
	event   event.Events
	pid     int
	max     int
}

func New(pid int, store *sync.Map) *Watcher {
	return &Watcher{
		pid:   pid,
		store: store,
	}
}

func (w Watcher) Enable() (enabled bool) {
	return true
}

func (w Watcher) Watch(stop <-chan struct{}, event chan<- event.Events) {
	for {
		select {
		case <-stop:
			return
		default:
			_ = w.do(event)
			if len(w.event) > 0 {
				event <- w.event
			}
		}
	}
}

func (w Watcher) do(e chan<- event.Events) (err error) {
	w.changed = false
	fds, err := fds(w.pid)
	if err != nil {
		return
	}
	w.close(fds)
	w.openOrChange(fds)
	return
}

func fds(pid int) (fds []int, err error) {
	_, err = os.Lstat(fmt.Sprintf("/proc/%d/", pid))
	if os.IsNotExist(err) {
		return
	}
	entries, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
	if err != nil {
		fmt.Printf("open /proc/%d/fd failed\n", pid)
		return
	}
	for _, fd := range entries {
		fd, err := strconv.Atoi(fd.Name())
		if err != nil {
			continue
		}
		fds = append(fds, fd)
	}
	sort.Ints(fds)
	return
}

// fds has been sorted
func (w Watcher) close(fds []int) {
	m := max(w.max, fds[len(fds)-1])
	var missing []int
	for i := 0; i < len(fds)-1; i++ {
		for j := fds[i] + 1; j < fds[i+1]; j++ {
			missing = append(missing, j)
		}
	}
	for i := fds[len(fds)-1] + 1; i <= m; i++ {
		missing = append(missing, i)
	}
	for _, fd := range missing {
		if _, ok := w.store.LoadAndDelete(fd); ok {
			w.changed = true
			w.event = append(w.event, event.Event{
				Type: event.Close,
				Pid:  w.pid,
				Fd:   fd,
			})
		}
	}
	w.max = fds[len(fds)-1]
	return
}

func (w Watcher) openOrChange(fds []int) {
	for _, fd := range fds {
		s, _ := stat.New(w.pid, fd)
		if old, ok := w.store.LoadOrStore(fd, s); !ok {
			w.event = append(w.event, event.Event{
				Type: event.Open,
				Pid:  w.pid,
				Fd:   fd,
				Stat: s,
			})
		} else {
			if old != s {
				w.event = append(w.event, event.Event{
					Type: event.Change,
					Pid:  w.pid,
					Fd:   fd,
					Stat: s,
				})
				w.store.Store(fd, s)
			}
		}
	}
}
