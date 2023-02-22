package main

import (
	"context"
	"flag"
	"github.com/blissd/cbz/cbrimportcmd"
	"github.com/blissd/cbz/infosetcmd"
	"github.com/blissd/cbz/infoshowcmd"
	"github.com/peterbourgon/ff/v3/ffcli"
	"log"
	"os"
)

func main() {
	root := &ffcli.Command{
		Name:       "root",
		ShortUsage: "cbz <subcommand>",
		Subcommands: []*ffcli.Command{
			infoshowcmd.New(os.Stdout),
			infosetcmd.New(os.Stdout),
			cbrimportcmd.New(os.Stdout),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
