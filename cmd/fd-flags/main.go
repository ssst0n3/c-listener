package main

import (
	"fmt"
	"github.com/ctrsploit/sploit-spec/pkg/version"
	"github.com/ssst0n3/fd-listener/pkg"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name: "fd-flags",
		Commands: []*cli.Command{
			version.Command,
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "pid",
				Aliases: []string{"p"},
			},
		},
		Action: func(context *cli.Context) (err error) {
			entries, err := os.ReadDir(fmt.Sprintf("/proc/%s/fd", context.String("pid")))
			if err != nil {
				return
			}
			for _, entry := range entries {
				flags, err := pkg.ReadFlags(fmt.Sprintf("/proc/%s/fdinfo/%s", context.String("pid"), entry.Name()))
				if err != nil {
					return
				}
				parsed := pkg.ParseFlags(flags)
				fmt.Println(entry.Name(), parsed)
			}
			return
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
