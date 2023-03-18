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
	"strings"
)

type config struct {
	out io.Writer

	// Compute the "pages" array
	computePages bool

	// inferDoublePages compute if a page is double width based on some simple heuristics
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

	setActions := make([]comicInfoAction, len(args), len(args)+2) // leave space for Validate and (optional) printXml actions

	for i, v := range args {
		nameAndValue := strings.Split(v, "=")
		if len(nameAndValue) != 2 {
			return fmt.Errorf("malformed metadata: '%v'", v)
		}
		typedValue, err := model.Convert(nameAndValue[0], nameAndValue[1])
		if err != nil {
			return fmt.Errorf("field %s has invalid value %s: %w", nameAndValue[0], nameAndValue[1], err)
		}
		setActions[i] = setField(nameAndValue[0], typedValue)
	}

	for _, name := range zipFileNames {
		actions := make([]comicInfoAction, 0, len(setActions)+3)
		for _, a := range setActions {
			actions = append(actions, a)
		}

		if c.inferDoublePages {
			actions = append(actions, c.inferDoubles(name))
		}

		action := join(append(actions, validate, c.printXml)) // TODO only add this comicInfoAction with a -v "verbose" flag

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

	err = os.Rename(updatedZip.Name(), zipFileName)
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

	for _, file := range input.File {
		if file.Name == model.ComicInfoXmlName {
			info, err = model.Unmarshal(file)
			if err != nil {
				return fmt.Errorf("failed to unmarshal ComicInfo.xml: %w", err)
			}
		}

		// Copies source file as-is. No-decompression/validation/re-compression.
		err = outputZip.Copy(file)
		if err != nil {
			return fmt.Errorf("failed to add %s: %w", file.Name, err)
		}
	}

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
	_, _ = fmt.Fprintln(c.out, info)
	return nil
}

// validate is an comicInfoAction that validates a ComicInfo.xml.
func validate(info *model.ComicInfo) error {
	return info.Validate()
}

func (cfg *config) inferDoubles(zipFileName string) comicInfoAction {
	return func(info *model.ComicInfo) error {
		input, err := zip.OpenReader(zipFileName)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer input.Close()

		var pageCount int
		for _, file := range input.File {
			if isImage(file.Name) {
				pageCount++
			}
		}

		pages := make([]model.ComicPageInfo, pageCount, pageCount)

		// copy data from any existing pages
		pageIndex := 0
		for _, file := range input.File {
			if !isImage(file.Name) {
				continue
			}

			if pageIndex < len(info.Pages) {
				pages[pageIndex] = info.Pages[pageIndex]
			}
			pages[pageIndex].Image = pageIndex
			_, err = cfg.updatePage(&pages[pageIndex], file)
			if err != nil {
				return fmt.Errorf("failed updating page: %w", err)
			}
			pageIndex++
		}

		// compute average page width and a range with tolerance for double page width
		var totalWidth int
		for _, p := range pages {
			totalWidth = totalWidth + p.ImageWidth
		}
		avgWidth := totalWidth / len(pages)
		avgWidth *= 2 // double average width for double pages
		loWidth, hiWidth := int(float64(avgWidth)*0.8), int(float64(avgWidth)*1.2)

		fmt.Println("lo/avg/hi", loWidth, "/", avgWidth, "/", hiWidth)

		for i, _ := range pages {
			if pages[i].ImageWidth >= loWidth && pages[i].ImageWidth <= hiWidth {
				pages[i].DoublePage = true
				//fmt.Println(pages[i])
				fmt.Printf("Computed double page for page %d with width=%d and height=%d\n", i, pages[i].ImageWidth, pages[i].ImageHeight)
			}
		}
		info.Pages = pages
		return nil
	}
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

func (c *config) updatePage(page *model.ComicPageInfo, file *zip.File) (*model.ComicPageInfo, error) {
	ir, err := file.Open()
	if err != nil {
		return page, fmt.Errorf("failed to open image file '%v': %w", file.Name, err)
	}
	defer ir.Close()

	var img image.Image = nil

	switch {
	case isJpeg(file.Name):
		img, err = jpeg.Decode(ir)
	case isPng(file.Name):
		img, err = png.Decode(ir)
	}

	if err != nil {
		return page, fmt.Errorf("failed to decode image '%v': %w", file.Name, err)
	}

	if img == nil {
		return page, fmt.Errorf("nil decode for image '%v'", file.Name)
	}

	page.ImageWidth = img.Bounds().Dx()
	page.ImageHeight = img.Bounds().Dy()

	return page, nil
}
func isImage(fileName string) bool {
	return isJpeg(fileName) || isPng(fileName)
}

func isJpeg(fileName string) bool {
	switch {
	case strings.HasSuffix(fileName, ".jpg"):
		return true
	case strings.HasSuffix(fileName, ".jpeg"):
		return true
	}
	return false
}

func isPng(fileName string) bool {
	switch {
	case strings.HasSuffix(fileName, ".png"):
		return true
	}
	return false
}
