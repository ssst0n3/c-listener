package main

import (
	"fmt"
	"github.com/ctrsploit/sploit-spec/pkg/version"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd"
	"github.com/ssst0n3/fd-listener/pkg/listener/process"
	"github.com/urfave/cli/v2"
	"os"
	"runtime"
	"time"
)

func main() {
	app := &cli.App{
		Name: "fd",
		Commands: []*cli.Command{
			version.Command,
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "allow",
				Aliases: []string{"a"},
			},
			&cli.StringSliceFlag{
				Name:    "deny",
				Aliases: []string{"d"},
			},
			&cli.BoolFlag{
				Name:    "task",
				Aliases: []string{"t"},
			},
		},
		Action: func(context *cli.Context) error {
			go func() {
				for {
					fmt.Println("routine", runtime.NumGoroutine())
					time.Sleep(5 * time.Second)
				}
			}()
			allow := context.StringSlice("allow")
			deny := context.StringSlice("deny")
			task := context.Bool("task")
			processListener := process.New(allow, deny, true, task, 0)
			fdListener := fd.New(processListener.Event)
			processListener.Listen()
			fdListener.Handle()
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
