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
	thread     bool
	Event      chan event.Event
	watcher    Watcher
	start      chan int
	exit       chan int
}

func New(allow, deny []string, exceptSelf, thread bool, stepLength int) (l *Listener) {
	l = &Listener{
		allow:      allow,
		deny:       deny,
		stepLength: DefaultStepLength,
		self:       os.Args[0],
		known:      sync.Map{},
		thread:     thread,
		Event:      make(chan event.Event),
		start:      make(chan int),
		exit:       make(chan int),
	}
	if exceptSelf {
		l.deny = append(l.deny, l.self)
	}
	if stepLength > 0 {
		l.stepLength = stepLength
	}
	l.watcher = NewWatcher(l.thread, l.stepLength, &l.known)
	return
}

func (l *Listener) Listen() {
	err := Watch(l.watcher, l.start, l.exit)
	if err != nil {
		return
	}
	go l.Start()
	go l.Exit()
}

func (l *Listener) filter(pid int) (valid bool, err error) {
	if _, ok := l.known.Load(pid); ok {
		return
	}
	// make sure process exists
	content, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		if !util.NotAlive(err) {
			awesome_error.CheckErr(err)
		}
		return
	}
	raw := string(content)
	cmdline := strings.Join(strings.Split(raw, "\x00"), " ")
	for _, allow := range l.allow {
		if !strings.Contains(cmdline, allow) {
			return
		}
	}
	for _, deny := range l.deny {
		if strings.Contains(cmdline, deny) {
			return
		}
	}
	valid = true
	color.Green("[+] new process: %d, %s", pid, cmdline)
	return
}

func (l *Listener) Start() {
	for {
		pid := <-l.start
		valid, err := l.filter(pid)
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
}

func (l *Listener) Exit() {
	for {
		pid := <-l.exit
		old, loaded := l.known.LoadOrStore(pid, false)
		if loaded && old.(bool) {
			//color.Green("[+] stop on: %d, %v", pid, old)
			l.Event <- event.Event{
				Type: event.ProcessExit,
				Pid:  pid,
			}
		}
	}
}
