package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Listener struct {
	self    string
	PidList chan int
	Store   sync.Map
}

func New() *Listener {
	return &Listener{
		self:    os.Args[0],
		PidList: make(chan int, 10),
	}
}

func valid(cmdline string, allows, denys []string) bool {
	for _, target := range allows {
		if !strings.Contains(cmdline, target) {
			return false
		}
	}
	for _, deny := range denys {
		if strings.Contains(cmdline, deny) {
			return false
		}
	}
	return true
}

func (l *Listener) Listen(passSelf bool, allows []string, denys []string) {
	for {
		dirs, err := os.ReadDir("/proc")
		if err != nil {
			panic(err)
		}
		for _, dir := range dirs {
			pid, err := strconv.ParseInt(dir.Name(), 10, 32)
			if err != nil {
				continue
			}

			//exe, err := filepath.EvalSymlinks(fmt.Sprintf("/proc/%d/exe", pid))
			//if err != nil {
			//	continue
			//}
			content, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
			if err != nil {
				continue
			}
			cmdline := string(content)
			if passSelf && strings.Contains(cmdline, l.self) {
				continue
			}
			if !valid(cmdline, allows, denys) {
				continue
			}

			if _, ok := l.Store.Load(int(pid)); !ok {
				fmt.Printf("[+] Found the PID: %d, %q\n", pid, strings.Split(cmdline, "\x00"))
			}
			l.PidList <- int(pid)
		}
	}
}

func (l *Listener) listFd(pid int) {
	v, _ := l.Store.LoadOrStore(pid, &sync.Map{})
	store := v.(*sync.Map)
	fds, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
	if err != nil {
		fmt.Printf("open /proc/%d/fd failed\n", pid)
		return
	}
	changed := false
	for _, name := range fds {
		fd, err := strconv.Atoi(name.Name())
		if err != nil {
			continue
		}
		path, _ := os.Readlink(fmt.Sprintf("/proc/%d/fd/%d", pid, fd))
		if path == "" {
			path = "?"
		}

		if old, ok := store.Load(fd); !ok {
			store.Store(fd, path)
			changed = true
		} else {
			if old != path && path != "?" {
				store.Store(fd, path)
				changed = true
			}
		}
	}
	if changed {
		var keys []int
		store.Range(func(key any, value any) bool {
			fd := key.(int)
			keys = append(keys, fd)
			return true
		})
		sort.Ints(keys)
		for _, fd := range keys {
			path, _ := store.Load(fd)
			fmt.Printf("/proc/%d/fd/%d -> %s\n", pid, fd, path)
		}
	}
}

func main() {
	app := &cli.App{
		Name: "fd-listener",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "allows",
				Aliases: []string{"a"},
			},
			&cli.StringSliceFlag{
				Name:    "denys",
				Aliases: []string{"d"},
			},
		},
		Action: func(context *cli.Context) error {
			l := New()
			go l.Listen(true, context.StringSlice("allows"), context.StringSlice("denys"))
			for {
				pid := <-l.PidList
				l.listFd(pid)
			}
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
