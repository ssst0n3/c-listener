package walk

import (
	"fmt"
	"github.com/ssst0n3/awesome_libs/awesome_error"
	"github.com/ssst0n3/fd-listener/pkg/util"
	"os"
	"strconv"
	"sync"
	"time"
)

const DefaultStepLength = 20

type Watcher struct {
	thread bool
	step   int
	known  *sync.Map // PID -> Alive
}

func New(thread bool, step int, known *sync.Map) *Watcher {
	if step <= 0 {
		step = DefaultStepLength
	}
	return &Watcher{
		thread: thread,
		step:   step,
		known:  known,
	}
}

func (w Watcher) Init() (err error) {
	return
}

func (w Watcher) Enable() (enabled bool) {
	return true
}

func (w Watcher) Start(c chan int) (err error) {
	lastPid, err := util.LastPid()
	if err != nil {
		panic(err)
	}
	if w.thread {
		for pid := lastPid; pid < lastPid+w.step; pid++ {
			c <- pid
		}
	} else {
		entries, err := os.ReadDir("/proc")
		if err != nil {
			panic(err)
		}
		for _, entry := range entries {
			pid, err := strconv.Atoi(entry.Name())
			if err != nil {
				continue
			}
			if pid >= lastPid-10 {
				c <- pid
			}
		}
	}
	return
}

func (w Watcher) Exit(c chan int) (err error) {
	for {
		w.known.Range(func(k, v interface{}) bool {
			pid := k.(int)
			alive := v.(bool)
			if !alive {
				return true
			}
			_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
			if err != nil {
				if util.NotAlive(err) {
					c <- pid
				} else {
					awesome_error.CheckErr(err)
				}
			}
			return true
		})
		time.Sleep(1 * time.Second)
	}
}
