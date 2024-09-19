package walk

import (
	"fmt"
	"github.com/ssst0n3/awesome_libs/awesome_error"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/event"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd/stat"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Watcher struct {
	pid   int
	store sync.Map
	event event.Events
	max   int
	stop  chan struct{}
}

func New(pid int) (w *Watcher) {
	w = &Watcher{
		pid:   pid,
		store: sync.Map{},
		event: event.Events{},
		stop:  make(chan struct{}),
	}
	return
}

func (w *Watcher) Enable() (enabled bool) {
	return true
}

func (w *Watcher) Watch(c chan<- event.Events) {
	go w.mmwatch()
	for {
		select {
		case <-w.stop:
			return
		default:
			e := event.Events{}
			_ = w.do(&e)
			if len(e) > 0 {
				c <- e
			}
		}
	}
}

func (w *Watcher) Close() {
	w.stop <- struct{}{}
}

func (w *Watcher) mmwatch() {
	_, err := os.Lstat("/proc/mmwatch")
	if err != nil {
		return
	}
	var last string
	for {
		err := os.WriteFile("/proc/mmwatch", []byte(fmt.Sprintf("%d", w.pid)), 0666)
		if err != nil {
			awesome_error.CheckErr(err)
			return
		}
		content, err := os.ReadFile("/proc/mmwatch")
		if err != nil {
			awesome_error.CheckErr(err)
			return
		}
		if last == string(content) {
			continue
		}
		last = string(content)
		if strings.Contains(string(content), "Process not found") {
			return
		} else {
			fmt.Print(string(content))
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (w *Watcher) do(e *event.Events) (err error) {
	fds, err := fds(w.pid)
	if err != nil {
		return
	}
	w.closeFd(fds, e)
	w.openOrChangeFd(fds, e)
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
func (w *Watcher) closeFd(fds []int, e *event.Events) {
	if len(fds) == 0 {
		return
	}
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
			*e = append(*e, event.Event{
				Type: event.Close,
				Pid:  w.pid,
				Fd:   fd,
				Stat: &stat.Stat{
					FdPath:   fmt.Sprintf("/proc/%d/fd/%d", w.pid, fd),
					RealPath: "[FD CLOSED]",
					Changed:  true,
				},
			})
		}
	}
	w.max = fds[len(fds)-1]
	return
}

func (w *Watcher) openOrChangeFd(fds []int, e *event.Events) {
	for _, fd := range fds {
		s, _ := stat.New(w.pid, fd)
		if old, ok := w.store.LoadOrStore(fd, s); !ok {
			*e = append(*e, event.Event{
				Type: event.Open,
				Pid:  w.pid,
				Fd:   fd,
				Stat: s,
			})
		} else {
			if !old.(*stat.Stat).Equals(s) {
				*e = append(*e, event.Event{
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
