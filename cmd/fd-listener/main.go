package main

import (
	"github.com/ctrsploit/sploit-spec/pkg/version"
	"github.com/ssst0n3/fd-listener/pkg/listener/fd"
	"github.com/ssst0n3/fd-listener/pkg/listener/process"
	"github.com/urfave/cli/v2"
	"os"
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
			allow := context.StringSlice("allow")
			deny := context.StringSlice("deny")
			processListener := process.New(allow, deny, true, 0)
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
