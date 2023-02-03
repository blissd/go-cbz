package infoshowcmd

import (
	"archive/zip"
	"cbz/model"
	"context"
	"fmt"
	"github.com/peterbourgon/ff/v3/ffcli"
	"log"
)

type config struct{}

func New() *ffcli.Command {
	c := config{}

	return &ffcli.Command{
		Name:       "show",
		ShortUsage: "cbz show <comic.cbz>",
		ShortHelp:  "Show the raw ComicInfo.xml file in a comic archive",
		Exec:       c.exec,
	}
}

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

			fmt.Println(info)
			return nil // early
		}
	}

	return fmt.Errorf("no ComicInfo.xml file found")
}