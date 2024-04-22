package fd

import (
	"fmt"
	"github.com/ssst0n3/fd-listener/pkg"
	"os"
	"sort"
	"strconv"
	"sync"
)

type Worker struct {
	pid   int
	Stop  chan bool
	store sync.Map
}

func NewWorker(pid int) (w *Worker) {
	w = &Worker{
		pid:  pid,
		Stop: make(chan bool),
	}
	go w.Work()
	return
}

func (l *Worker) Work() {
	for {
		select {
		case <-l.Stop:
			return
		default:
			l.do()
		}
	}
}

func (l *Worker) stat(fd int) (stat Stat, err error) {
	fdPath := fmt.Sprintf("/proc/%d/fd/%d", l.pid, fd)
	realPath, _ := os.Readlink(fdPath)
	if realPath == "" {
		realPath = "?"
	}
	stat = Stat{
		FdPath:   fdPath,
		RealPath: realPath,
	}
	leak, err := Leak(fdPath, realPath)
	if err != nil {
		return
	}
	flags, err := pkg.ReadFlags(fmt.Sprintf("/proc/%d/fdinfo/%d", l.pid, fd))
	if err != nil {
		return
	}
	stat = Stat{
		FdPath:   fdPath,
		RealPath: realPath,
		Leak:     leak,
		Flags:    flags,
	}
	return
}

func (l *Worker) do() {
	_, err := os.Lstat(fmt.Sprintf("/proc/%d/", l.pid))
	if os.IsNotExist(err) {
		return
	}
	fds, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", l.pid))
	if err != nil {
		fmt.Printf("open /proc/%d/fd failed\n", l.pid)
		return
	}
	changed := false
	for _, name := range fds {
		fd, err := strconv.Atoi(name.Name())
		if err != nil {
			continue
		}
		stat, _ := l.stat(fd)
		if old, ok := l.store.Load(fd); !ok {
			l.store.Store(fd, stat)
			changed = true
		} else {
			if old != stat {
				l.store.Store(fd, stat)
				changed = true
			}
		}
	}
	if changed {
		l.print()
	}
}

func (l *Worker) print() {
	var keys []int
	l.store.Range(func(key any, value any) bool {
		fd := key.(int)
		keys = append(keys, fd)
		return true
	})
	sort.Ints(keys)
	for _, fd := range keys {
		stat, _ := l.store.Load(fd)
		fmt.Println(stat)
	}
	fmt.Println("----------------")
}
