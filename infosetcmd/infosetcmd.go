package infosetcmd

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/blissd/cbz/model"
	"github.com/peterbourgon/ff/v3/ffcli"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

type config struct {
	out io.Writer

	// Compute the "pages" array
	computePages bool

	// Detect double page spreads from file names such as "Page 005-006.jpeg" Implies computePages is true.
	// TODO fix this as too many comics don't follow this convention.
	inferDoublePages bool
}

// New creates a ffcli.Command for updating the metadata in a ComicInfo.xml file.
// Can update multiple fields at once. Operates on multiple CBZ files sequentially.
func New(out io.Writer) *ffcli.Command {

	cfg := config{
		out: out,
	}
	fs := flag.NewFlagSet("cbz set", flag.ExitOnError)
	fs.BoolVar(&cfg.computePages, "p", false, "compute values for the 'pages' element")
	fs.BoolVar(&cfg.inferDoublePages, "d", false, "infer double page spreads. Implies -p.")

	return &ffcli.Command{
		Name:       "set",
		ShortUsage: "cbz set <field=value> <comic.cbz>",
		ShortHelp:  "Set an field value in ComicInfo.xml. e.g., cbz meta set AgeRating=M comic.cbz",
		FlagSet:    fs,
		Exec:       cfg.exec,
	}
}

// exec is the callback for ffcli.Command
func (c *config) exec(_ context.Context, args []string) error {

	zipFileNames := []string{}

	for _, v := range args {
		if strings.HasSuffix(v, ".cbz") {
			zipFileNames = append(zipFileNames, v)
		}
	}

	// remove file names from argument list so only metadata name=value pairs are left
	args = args[:len(args)-len(zipFileNames)]

	actions := make([]comicInfoAction, len(args), len(args)+2) // leave space for Validate and (optional) printXml actions

	for i, v := range args {
		nameAndValue := strings.Split(v, "=")
		if len(nameAndValue) != 2 {
			return fmt.Errorf("malformed metadata: '%v'", v)
		}
		typedValue, err := model.Convert(nameAndValue[0], nameAndValue[1])
		if err != nil {
			return fmt.Errorf("field %s has invalid value %s: %w", nameAndValue[0], nameAndValue[1], err)
		}
		actions[i] = setField(nameAndValue[0], typedValue)
	}

	action := join(append(actions, validate, c.printXml)) // TODO only add this comicInfoAction with a -v "verbose" flag

	for _, name := range zipFileNames {
		err := c.updateZip(name, action)
		if err != nil {
			return fmt.Errorf("failed updating comic archive '%s': %w", name, err)
		}
	}

	return nil
}

// updateZip updates a single zip file to transform the ComicInfo.xml file.
// Source file will be replaced with updated version.
func (c *config) updateZip(zipFileName string, action comicInfoAction) error {

	updatedZip, err := os.CreateTemp(filepath.Dir(zipFileName), filepath.Base(zipFileName))
	if err != nil {
		return fmt.Errorf("failed creating temporary file: %w", err)
	}

	err = c.applyActions(zipFileName, action, updatedZip)
	if err != nil {
		updatedZip.Close()
		os.Remove(updatedZip.Name())
		return fmt.Errorf("failed processing comic book archive: %w", err)
	}

	// Success, so replace original file with updated file.
	updatedZip.Close()

	os.Rename(updatedZip.Name(), zipFileName)
	if err != nil {
		return fmt.Errorf("failed moving file: %w", err)
	}

	return nil
}

// applyActions applies a series of actions to files in a zip archive.
func (c *config) applyActions(zipFileName string, action comicInfoAction, output io.Writer) error {
	input, err := zip.OpenReader(zipFileName)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	outputZip := zip.NewWriter(output)
	defer outputZip.Close()

	// Will write the ComicInfo.xml file as the last entry of the CBZ archive
	var info *model.ComicInfo = &model.ComicInfo{}

	// Note any existing page data in the ComicInfo.xml will be lost!
	pages := make([]model.ComicPageInfo, 0, len(input.File))

	addPage := noopAddPage
	if c.computePages || c.inferDoublePages {
		addPage = c.addPage
	}

	for _, file := range input.File {
		if file.Name == model.ComicInfoXmlName {
			info, err = model.Unmarshal(file)
			if err != nil {
				return fmt.Errorf("failed to unmarshal ComicInfo.xml: %w", err)
			}
		} else {
			pages, err = addPage(pages, file)
			if err != nil {
				return fmt.Errorf("failed to process page file: %w", err)
			}

			// Copies source file as-is. No-decompression/validation/re-compression.
			err = outputZip.Copy(file)
			if err != nil {
				return fmt.Errorf("failed to add %s: %w", file.Name, err)
			}
		}
	}

	// TODO only do this if len(info.Pages) is zero? Or can they be merged?
	info.Pages = pages

	err = action(info)
	if err != nil {
		return fmt.Errorf("failed to apply comicInfoAction to ComicInfo.xml: %w", err)
	}

	err = info.Validate()
	if err != nil {
		return fmt.Errorf("failed to produce a valid ComicInfo.xml: %w", err)
	}

	bs, err := xml.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ComicInfo.xml: %w", err)
	}
	w, err := outputZip.Create(model.ComicInfoXmlName)
	if _, err = w.Write(bs); err != nil {
		return fmt.Errorf("failed to write ComicInfo.xml: %w", err)
	}

	return nil
}

// comicInfoAction performs an comicInfoAction on a ComicInfo, such as printing a value, setting a value, or removing a value.
type comicInfoAction func(info *model.ComicInfo) error

// printXml is an comicInfoAction that prints the ComicInfo.xml to stdout.
func (c *config) printXml(info *model.ComicInfo) error {
	fmt.Fprintln(c.out, info)
	return nil
}

// validate is an comicInfoAction that validates a ComicInfo.xml.
func validate(info *model.ComicInfo) error {
	return info.Validate()
}

// join many Actions together into a single comicInfoAction.
func join(actions []comicInfoAction) comicInfoAction {
	return func(info *model.ComicInfo) error {
		for _, action := range actions {
			if err := action(info); err != nil {
				return fmt.Errorf("failed applying comicInfoAction: %w", err)
			}
		}
		return nil
	}
}

// setField overwrites the value of a named field in a ComicInfo.
// Uses reflection... like a monster.
func setField(name string, value any) comicInfoAction {
	return func(info *model.ComicInfo) error {
		rv := reflect.Indirect(reflect.ValueOf(info))
		f := rv.FieldByName(name)
		switch v := value.(type) {
		case string:
			f.SetString(v)
		case int64:
			f.SetInt(v)
		case float64:
			f.SetFloat(v)
		case bool:
			f.SetBool(v)
		default:
			return fmt.Errorf("field %s has unsupported data type: %v", name, v)
		}
		return nil
	}
}

// addPage accumulates information about pages to build the Pages.
type addPage func(pages []model.ComicPageInfo, file *zip.File) ([]model.ComicPageInfo, error)

func noopAddPage(pages []model.ComicPageInfo, file *zip.File) ([]model.ComicPageInfo, error) {
	return pages, nil
}

func (c *config) addPage(pages []model.ComicPageInfo, file *zip.File) ([]model.ComicPageInfo, error) {
	if file.Name == model.ComicInfoXmlName {
		return pages, nil
	}

	isJpeg := false
	isPng := false
	switch {
	case strings.HasSuffix(file.Name, ".jpg"):
		isJpeg = true
	case strings.HasSuffix(file.Name, ".jpeg"):
		isJpeg = true
	case strings.HasSuffix(file.Name, ".png"):
		isPng = true
	}

	isImage := isJpeg || isPng

	if isImage {
		page := model.ComicPageInfo{
			Image: len(pages),
			Type:  "Story",
		}

		ir, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open image file '%v': %w", file.Name, err)
		}

		var img image.Image = nil

		switch {
		case isJpeg:
			img, err = jpeg.Decode(ir)
		case isPng:
			img, err = png.Decode(ir)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to decode image '%v': %w", file.Name, err)
		}

		if img == nil {
			return nil, fmt.Errorf("nil decode for image '%v'", file.Name)
		}

		page.ImageWidth = img.Bounds().Dx()
		page.ImageHeight = img.Bounds().Dy()

		if c.inferDoublePages {
			// Assume that double spread pages have names containing a hyphen surrounded by numbers
			match, err := regexp.MatchString("^.*[0-9]-[0-9].*$", file.Name)
			if err != nil {
				return nil, fmt.Errorf("failed double-page inference regex: %w", err)
			}
			page.DoublePage = match
		}

		pages = append(pages, page)
	}
	return pages, nil
}
