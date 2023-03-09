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

	// includeTitle include title in file name in addition to series
	includeTitle bool
}

func New(out io.Writer) *ffcli.Command {
	cfg := config{
		out: out,
	}
	fs := flag.NewFlagSet("cbz rename", flag.ExitOnError)

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

	_, err = os.Stat(newPath)

	if errors.Is(err, fsys.ErrExist) {
		return fmt.Errorf("destination file already exists: %v", newPath)
	}

	err = os.Rename(fileName, newPath)
	if err != nil {
		return fmt.Errorf("failed renaming file '%v' to '%v': %w", fileName, newPath, err)
	}
	return nil
}

func (cfg *config) inferFileName(c *model.ComicInfo) (string, error) {
	b := strings.Builder{}

	if c.Series != "" {
		b.WriteString(c.Series)
	}
	if c.Title != "" {
		if b.Len() > 0 {
			b.WriteString(" - ")
		}
		b.WriteString(c.Title)
	}

	if c.Volume > 0 {
		b.WriteString(fmt.Sprintf(" v%02d", c.Volume))
	}

	if c.Number != "" {
		b.WriteString(" #")
		b.WriteString(c.Number)
	}

	return b.String(), nil
}
