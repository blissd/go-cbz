package main

import (
	"cbz/infosetcmd"
	"cbz/infoshowcmd"
	"context"
	"flag"
	"github.com/peterbourgon/ff/v3/ffcli"
	"log"
	"os"
)

func main() {
	root := &ffcli.Command{
		Name:       "root",
		ShortUsage: "cbz <subcommand>",
		Subcommands: []*ffcli.Command{
			infoshowcmd.New(),
			infosetcmd.New(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
