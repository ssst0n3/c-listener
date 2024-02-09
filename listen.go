package main

import (
	"fmt"
	"github.com/ctrsploit/sploit-spec/pkg/version"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type Listener struct {
	self    string
	PidList chan int
	Store   sync.Map
	Fd      chan [2]string
}

func New() *Listener {
	return &Listener{
		self:    os.Args[0],
		PidList: make(chan int, 10),
		Fd:      make(chan [2]string, 10),
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
	_, err := os.Lstat(fmt.Sprintf("/proc/%d/", pid))
	if os.IsNotExist(err) {
		return
	}
	v, _ := l.Store.LoadOrStore(pid, &sync.Map{})
	store := v.(*sync.Map)
	leaked := map[int]bool{}
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
		fdPath := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
		realPath, _ := os.Readlink(fdPath)
		if realPath == "" {
			realPath = "?"
		}

		if detect(fdPath, realPath) {
			leaked[fd] = true
		}

		if old, ok := store.Load(fd); !ok {
			store.Store(fd, realPath)
			//l.Fd <- [2]string{fdPath, realPath}
			changed = true
		} else {
			if old != realPath {
				store.Store(fd, realPath)
				//l.Fd <- [2]string{fdPath, realPath}
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
			fdPath := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
			realPath, _ := store.Load(fd)
			if _, ok := leaked[fd]; ok {
				color.Red(fmt.Sprintf("%s -> %s\t; leaked!\n", fdPath, realPath))
			} else {
				fmt.Printf("%s -> %s\n", fdPath, realPath)
			}
		}
		fmt.Println("----------------")
	}
}

func detect(fdPath, realPath string) (leaked bool) {
	fdFI, err := os.Stat(fdPath)
	if err != nil {
		return
	}
	fdStat, _ := fdFI.Sys().(*syscall.Stat_t)
	realFI, err := os.Stat(realPath)
	if err != nil {
		return
	}
	realStat, _ := realFI.Sys().(*syscall.Stat_t)
	if fdStat.Ino == realStat.Ino {
		leaked = true
	}
	return
}

func (l *Listener) Detect() {
	for {
		paths := <-l.Fd
		fdPath, realPath := paths[0], paths[1]

		if detect(fdPath, realPath) {
			color.Red(fmt.Sprintf("[!] leaked path: %s -> %s\n", fdPath, realPath))
		}
	}
}

func main() {
	listener := &cli.App{
		Name: "fd-listener",
		Commands: []*cli.Command{
			version.Command,
		},
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
			go l.Detect()
			for {
				pid := <-l.PidList
				l.listFd(pid)
			}
		},
	}
	err := listener.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
