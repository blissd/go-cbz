package infosetcmd

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/blissd/cbz/model"
	"github.com/peterbourgon/ff/v3/ffcli"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type config struct {
	out io.Writer
}

// New creates a ffcli.Command for updating the metadata in a ComicInfo.xml file.
// Can update multiple fields at once. Operates on multiple CBZ files sequentially.
func New(out io.Writer) *ffcli.Command {

	c := config{
		out: out,
	}

	return &ffcli.Command{
		Name:       "set",
		ShortUsage: "cbz set <field=value> <comic.cbz>",
		ShortHelp:  "Set an field value in ComicInfo.xml. e.g., cbz meta set AgeRating=M comic.cbz",
		Exec:       c.exec,
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

	actions := make([]action, len(args), len(args)+2) // leave space for Validate and (optional) printXml actions

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

	action := join(append(actions, validate, c.printXml)) // TODO only add this action with a -v "verbose" flag

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
func (c *config) updateZip(zipFileName string, action action) error {

	updatedZip, err := os.CreateTemp(filepath.Dir(zipFileName), filepath.Base(zipFileName))
	if err != nil {
		return fmt.Errorf("failed creating temporary file: %w", err)
	}

	err = applyActions(zipFileName, action, updatedZip)
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
func applyActions(zipFileName string, action action, output io.Writer) error {
	input, err := zip.OpenReader(zipFileName)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	outputZip := zip.NewWriter(output)
	defer outputZip.Close()

	// Will write the ComicInfo.xml file as the last entry of the CBZ archive
	var info *model.ComicInfo

	for _, file := range input.File {
		if file.Name == model.ComicInfoXmlName {
			info, err = model.Unmarshal(file)
			if err != nil {
				return fmt.Errorf("failed to unmarshal ComicInfo.xml: %w", err)
			}
		} else {
			// Copies source file as-is. No-decompression/validation/re-compression.
			err = outputZip.Copy(file)
			if err != nil {
				return fmt.Errorf("failed to add %s: %w", file.Name, err)
			}
		}
	}

	// If no ComicInfo.xml file was found then create a new one.
	if info == nil {
		info = &model.ComicInfo{}
	}

	err = action(info)
	if err != nil {
		return fmt.Errorf("failed to apply action to ComicInfo.xml: %w", err)
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

// action performs an action on a ComicInfo, such as printing a value, setting a value, or removing a value.
type action func(info *model.ComicInfo) error

// printXml is an action that prints the ComicInfo.xml to stdout.
func (c *config) printXml(info *model.ComicInfo) error {
	fmt.Fprintln(c.out, info)
	return nil
}

// validate is an action that validates a ComicInfo.xml.
func validate(info *model.ComicInfo) error {
	return info.Validate()
}

// join many Actions together into a single action.
func join(actions []action) action {
	return func(info *model.ComicInfo) error {
		for _, action := range actions {
			if err := action(info); err != nil {
				return fmt.Errorf("failed applying action: %w", err)
			}
		}
		return nil
	}
}

// setField overwrites the value of a named field in a ComicInfo.
// Uses reflection... like a monster.
func setField(name string, value any) action {
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
