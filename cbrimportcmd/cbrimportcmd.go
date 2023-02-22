package cbrimportcmd

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"github.com/gen2brain/go-unarr"
	"github.com/peterbourgon/ff/v3/ffcli"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type config struct {
	out io.Writer
}

// New creates a ffcli.Command for updating the metadata in a ComicInfo.xml file.
// Can update multiple fields at once. Operates on multiple CBZ files sequentially.
func New(out io.Writer) *ffcli.Command {

	cfg := config{
		out: out,
	}
	fs := flag.NewFlagSet("cbz import", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "import",
		ShortUsage: "cbz import <comic.cbr>",
		ShortHelp:  "Imports a CBR file and converts it into a CBZ file.",
		FlagSet:    fs,
		Exec:       cfg.exec,
	}
}

// exec is the callback for ffcli.Command
func (c *config) exec(_ context.Context, args []string) error {

	rarFileName := args[0]

	input, err := unarr.NewArchive(rarFileName)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	outputZip, err := os.CreateTemp(filepath.Dir(rarFileName), filepath.Base(rarFileName))
	if err != nil {
		return fmt.Errorf("failed creating temporary file: %w", err)
	}
	defer outputZip.Close()

	output := zip.NewWriter(outputZip)
	defer output.Close()

	for {
		err := input.Entry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed moving to next RAR entry: %w", err)
		}

		bs, err := input.ReadAll()
		if err != nil {
			return fmt.Errorf("failed reading RAR entry: %w", err)
		}

		w, err := output.Create(filepath.Base(input.Name()))
		if _, err = w.Write(bs); err != nil {
			return fmt.Errorf("failed to write ZIP entry: %w", err)
		}
	}

	output.Close()
	outputZip.Close()

	{
		dir := filepath.Dir(rarFileName)
		base := filepath.Base(rarFileName)
		base = strings.TrimSuffix(base, ".cbr")
		cbzName := fmt.Sprintf("%v/%v.cbz", dir, base)
		err = os.Rename(outputZip.Name(), cbzName)
		if err != nil {
			return fmt.Errorf("failed moving file: %w", err)
		}
	}

	return nil
}
