package process

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ssst0n3/awesome_libs/awesome_error"
	"github.com/ssst0n3/fd-listener/pkg/event"
	"github.com/ssst0n3/fd-listener/pkg/util"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DefaultStepLength = 20
)

type Listener struct {
	allow      []string
	deny       []string
	stepLength int
	self       string
	known      sync.Map // PID -> Alive
	Event      chan event.Event
}

func New(allow, deny []string, exceptSelf bool, stepLength int) (l *Listener) {
	l = &Listener{
		allow:      allow,
		deny:       deny,
		stepLength: DefaultStepLength,
		self:       os.Args[0],
		known:      sync.Map{},
		Event:      make(chan event.Event),
	}
	if exceptSelf {
		l.deny = append(l.deny, l.self)
	}
	if stepLength > 0 {
		l.stepLength = stepLength
	}
	return
}

func (l *Listener) Listen() {
	go l.New()
	go l.Exit()
}

func (l *Listener) exit(pid int) {
	old, loaded := l.known.LoadOrStore(pid, false)
	if loaded && old.(bool) {
		//color.Green("[+] stop on: %d, %v", pid, old)
		l.Event <- event.Event{
			Type: event.ProcessExit,
			Pid:  pid,
		}
	}
}

func (l *Listener) handleError(pid int, err error) error {
	if os.IsNotExist(err) || strings.Contains(err.Error(), "no such process") {
		l.exit(pid)
		return nil
	} else {
		awesome_error.CheckErr(err)
		return err
	}
}

func (l *Listener) filter(pid int) (valid bool, err error) {
	if _, ok := l.known.Load(pid); ok {
		return
	}
	{
		// make sure process exists
		_, err = os.Stat(fmt.Sprintf("/proc/%d", pid))
		if err != nil {
			return false, l.handleError(pid, err)
		}
	}
	content, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return false, l.handleError(pid, err)
	}
	cmdline := string(content)
	for _, target := range l.allow {
		if !strings.Contains(cmdline, target) {
			return
		}
	}
	for _, deny := range l.deny {
		if strings.Contains(cmdline, deny) {
			return
		}
	}
	valid = true
	color.Green("[+] new process: %d, %q", pid, strings.Split(cmdline, "\x00"))
	return
}

func (l *Listener) New() {
	for {
		// only search new started process
		lastPid, err := util.LastPid()
		if err != nil {
			return
		}
		for pid := lastPid; pid < lastPid+l.stepLength; pid++ {
			valid, err := l.filter(lastPid)
			if err != nil {
				continue
			}
			if valid {
				l.known.Store(pid, true)
				l.Event <- event.Event{
					Type: event.ProcessNew,
					Pid:  pid,
				}
			}
		}
		time.Sleep(10 * time.Microsecond)
	}
}

func (l *Listener) Exit() {
	for {
		l.known.Range(func(k, v interface{}) bool {
			pid := k.(int)
			alive := v.(bool)
			if !alive {
				return true
			}
			_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
			if err != nil {
				_ = l.handleError(pid, err)
			}
			return true
		})
		time.Sleep(1 * time.Second)
	}
}
