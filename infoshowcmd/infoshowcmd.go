package infoshowcmd

import (
	"archive/zip"
	"context"
	"fmt"
	"github.com/blissd/cbz/model"
	"github.com/peterbourgon/ff/v3/ffcli"
	"io"
	"log"
)

type config struct {
	out io.Writer
}

// New creates a new ffcli.Command for showing a ComicInfo.xml file.
// Operates on a single CBZ file.
func New(out io.Writer) *ffcli.Command {
	c := config{
		out: out,
	}

	return &ffcli.Command{
		Name:       "show",
		ShortUsage: "cbz show <comic.cbz>",
		ShortHelp:  "Show the raw ComicInfo.xml file in a comic archive",
		Exec:       c.exec,
	}
}

// exec is the callback for ffcli.Command
func (c *config) exec(ctx context.Context, args []string) error {
	input, err := zip.OpenReader(args[0])
	if err != nil {
		log.Fatalln("Failed to open file:", err)
	}
	defer input.Close()

	for _, file := range input.File {
		if file.Name == model.ComicInfoXmlName {
			info, err := model.Unmarshal(file)
			if err != nil {
				return fmt.Errorf("failed to unmarshal ComicInfo.xml: %w", err)
			}

			fmt.Fprintln(c.out, info)
			return nil // early
		}
	}

	return fmt.Errorf("no ComicInfo.xml file found")
}
