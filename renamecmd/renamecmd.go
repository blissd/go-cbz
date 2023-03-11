package renamecmd

import (
	"archive/zip"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/blissd/cbz/model"
	"github.com/peterbourgon/ff/v3/ffcli"
	"io"
	fsys "io/fs"
	"os"
	"path/filepath"
	"strings"
)

type config struct {
	out io.Writer

	// includeTitle include Title in file name in addition to series
	includeTitle bool

	// includeNumber include Number field in file name.
	includeNumber bool

	// dryRun disables applying renames and just prints new names instead.
	dryRun bool
}

func New(out io.Writer) *ffcli.Command {
	cfg := config{
		out: out,
	}
	fs := flag.NewFlagSet("cbz rename", flag.ExitOnError)
	fs.BoolVar(&cfg.includeTitle, "t", false, "include comic title in file name.")
	fs.BoolVar(&cfg.includeNumber, "n", false, "include comic number in file name.")
	fs.BoolVar(&cfg.dryRun, "d", false, "dry-run")

	return &ffcli.Command{
		Name:       "rename",
		ShortUsage: "cbz rename <comic.cbz>",
		ShortHelp:  "Renames CBZ file based on ComicInfo.xml metadata",
		FlagSet:    fs,
		Exec:       cfg.exec,
	}
}

func (cfg *config) exec(_ context.Context, args []string) error {

	var zipFileNames []string

	for _, v := range args {
		if strings.HasSuffix(v, ".cbz") {
			zipFileNames = append(zipFileNames, v)
		}
	}

	for _, name := range zipFileNames {
		err := cfg.rename(name)
		if err != nil {
			return fmt.Errorf("failed updating comic archive '%s': %w", name, err)
		}
	}

	return nil
}

// rename computes the new name for a file, but doesn't rename the file.
func (cfg *config) rename(fileName string) error {

	zipFile, err := zip.OpenReader(fileName)
	if err != nil {
		return fmt.Errorf("failed opening zip file: %w", err)
	}

	var comicInfoFile *zip.File
	for i, f := range zipFile.File {
		if f.Name == model.ComicInfoXmlName {
			comicInfoFile = zipFile.File[i]
		}
	}

	if comicInfoFile == nil {
		return fmt.Errorf("no ComicInfo.xml file")
	}

	comicInfo, err := model.Unmarshal(comicInfoFile)
	if err != nil {
		return fmt.Errorf("failed unmarshalling ComicInfo.xml: %w", err)
	}

	inferredFileName, err := cfg.inferFileName(comicInfo)
	if err != nil {
		return fmt.Errorf("failed inferring name: %w", err)
	}

	dir := filepath.Dir(fileName)
	newPath := filepath.Join(dir, inferredFileName)
	newPath = fmt.Sprintf("%s.cbz", newPath)

	// Renaming should be non-destructive, so fail if the destination file exists.
	// We can tell it exists if we _don't_ get an error!
	if _, err = os.Stat(newPath); !errors.Is(err, fsys.ErrNotExist) {
		if err != nil {
			return fmt.Errorf("file exists '%v'", newPath)
		}
		return fmt.Errorf("file already exists: '%v'", newPath)
	}

	if cfg.dryRun {
		fmt.Printf("Dry-run: would rename '%s' to '%s'\n", fileName, newPath)
	} else {
		err = os.Rename(fileName, newPath)
		if err != nil {
			return fmt.Errorf("failed renaming file '%v' to '%v': %w", fileName, newPath, err)
		}
	}
	return nil
}

func (cfg *config) inferFileName(c *model.ComicInfo) (string, error) {
	b := strings.Builder{}

	if c.Series != "" {
		b.WriteString(c.Series)
	} else if c.Title != "" {
		b.WriteString(c.Title)
	}

	if c.Volume > 0 {
		b.WriteString(fmt.Sprintf(" v%02d", c.Volume))
	}

	// If there is _no_ Volume but there is a Number, then include Number anyway.
	if (cfg.includeNumber && c.Number != "") || (c.Number != "" && c.Volume == 0) {
		b.WriteString(" #")
		b.WriteString(c.Number)
	}

	if cfg.includeTitle && c.Title != "" && c.Series != "" {
		if b.Len() > 0 {
			b.WriteString(" - ")
		}
		b.WriteString(c.Title)
	}

	if b.Len() == 0 {
		return "", fmt.Errorf("not enough metadata to rename file")
	}

	return b.String(), nil
}
