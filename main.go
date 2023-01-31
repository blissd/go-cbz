package main

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/peterbourgon/ff/v3/ffcli"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

const comicInfoXmlName = "ComicInfo.xml"

func main() {

	metaShow := &ffcli.Command{
		Name:       "show",
		ShortUsage: "cbz meta show <comic.cbz>",
		ShortHelp:  "Show the raw ComicInfo.xml file in a comic archive",
		Exec:       showComicInfo,
	}

	metaSet := &ffcli.Command{
		Name:       "set",
		ShortUsage: "cbz meta set <field=value> <comic.cbz>",
		ShortHelp:  "Set an field value. e.g., cbz meta set AgeRating=M comic.cbz",
		Exec:       setComicInfoField,
	}

	meta := &ffcli.Command{
		Name:        "meta",
		ShortUsage:  "cbz meta <subcommand>",
		ShortHelp:   "Display and manipulate ComicInfo.xml file in a comic archive",
		Subcommands: []*ffcli.Command{metaShow, metaSet},
	}

	root := &ffcli.Command{
		Name:        "root",
		ShortUsage:  "cbz <subcommand>",
		Subcommands: []*ffcli.Command{meta},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

// process applies a series of actions to files in a zip archive.
func process(zipFileName string, action Action, output io.Writer) error {
	input, err := zip.OpenReader(zipFileName)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	outputZip := zip.NewWriter(output)
	defer outputZip.Close()

	// Will write the ComicInfo.xml file as the last entry of the CBZ archive
	var info *ComicInfo

	for _, file := range input.File {
		if err != nil {
			return fmt.Errorf("failed creating file in zip archive: %w", err)
		}

		if file.Name == comicInfoXmlName {
			if err != nil {
				return fmt.Errorf("failed to show ComicInfo.xml: %w", err)
			}

			info, err = unmarshallComicInfoXml(file)
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

	if info == nil {
		info = &ComicInfo{}
	}

	err = action(info)
	if err != nil {
		return fmt.Errorf("failed to apply action to ComicInfo.xml: %w", err)
	}

	err = info.validate()
	if err != nil {
		return fmt.Errorf("failed to produce a valid ComicInfo.xml: %w", err)
	}

	bs, err := xml.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ComicInfo.xml: %w", err)
	}
	w, err := outputZip.Create(comicInfoXmlName)
	if _, err = w.Write(bs); err != nil {
		return fmt.Errorf("failed to write ComicInfo.xml: %w", err)
	}

	return nil
}

func showComicInfo(_ context.Context, args []string) error {
	input, err := zip.OpenReader(args[0])
	if err != nil {
		log.Fatalln("Failed to open file:", err)
	}
	defer input.Close()

	for _, file := range input.File {
		if file.Name == "ComicInfo.xml" {
			if err != nil {
				return fmt.Errorf("failed to show ComicInfo.xml: %w", err)
			}

			info, err := unmarshallComicInfoXml(file)
			err = printXml(info)
			if err != nil {
				return fmt.Errorf("failed to apply action to ComicInfo.xml: %w", err)
			}
			return nil // early
		}
	}

	return fmt.Errorf("no ComicInfo.xml file found")
}

var intFieldNames = []string{
	"Count",
	"Volume",
	"AlternativeCount",
	"Year",
	"Month",
	"Day",
	"PageCount",
}

var floatFieldNames = []string{
	"CommunityRating",
}
var boolFieldNames = []string{
	"DoublePage",
}

func convertFieldValue(name string, value string) (any, error) {
	for _, n := range intFieldNames {
		if n == name {
			return strconv.ParseInt(value, 10, 64)
		}
	}
	for _, n := range floatFieldNames {
		if n == name {
			return strconv.ParseFloat(value, 32)
		}
	}
	for _, n := range boolFieldNames {
		if n == name {
			return strconv.ParseBool(value)
		}
	}
	return value, nil
}

func setComicInfoField(_ context.Context, args []string) error {

	zipFileNames := []string{}

	for _, v := range args {
		if strings.HasSuffix(v, ".cbz") {
			zipFileNames = append(zipFileNames, v)
		}
	}

	args = args[:len(zipFileNames)]
	actions := make([]Action, len(args), len(args)+2) // leave space for validate and (optional) printXml actions

	for i, v := range args {
		nameAndValue := strings.Split(v, "=")
		typedValue, err := convertFieldValue(nameAndValue[0], nameAndValue[1])
		if err != nil {
			return fmt.Errorf("field %s has invalid value %s: %w", nameAndValue[0], nameAndValue[1], err)
		}
		actions[i] = setField(nameAndValue[0], typedValue)
	}

	action := join(append(actions, validate, printXml)) // TODO only add this action with a -v "verbose" flag

	for _, name := range zipFileNames {
		err := updateZip(name, action)
		if err != nil {
			return fmt.Errorf("failed updating comic archive '%s': %w", name, err)
		}
	}

	return nil
}

func updateZip(zipFileName string, action Action) error {

	updatedZip, err := os.CreateTemp(filepath.Dir(zipFileName), filepath.Base(zipFileName))
	if err != nil {
		return fmt.Errorf("failed creating temporary file: %w", err)
	}

	err = process(zipFileName, action, updatedZip)
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

func outputCbzName(sourcePath string) string {
	dir := filepath.Dir(sourcePath)
	base := filepath.Base(sourcePath)
	ext := filepath.Ext(sourcePath)
	baseNoExt := strings.TrimSuffix(base, ext)
	return fmt.Sprintf("%s/%s-updated%s", dir, baseNoExt, ext)
}

// Action performs an action on a ComicInfo, such as printing a value, setting a value, or removing a value.
type Action func(info *ComicInfo) error

// printXml prints the ComicInfo.xml to stdout
func printXml(info *ComicInfo) error {
	marshal, err := xml.MarshalIndent(&info, "", " ")
	if err != nil {
		return fmt.Errorf("failed to XML marshal ComicInfo.xml: %w", err)
	}

	fmt.Println(string(marshal))
	return nil
}

// setField overwrites the value of a named field in a ComicInfo.
// Uses reflection... like a monster.
func setField(name string, value any) Action {
	return func(info *ComicInfo) error {
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

func validate(info *ComicInfo) error {
	return info.validate()
}

// join many actions together into a single action
func join(actions []Action) Action {
	return func(info *ComicInfo) error {
		for _, action := range actions {
			if err := action(info); err != nil {
				return fmt.Errorf("failed applying action: %w", err)
			}
		}
		return nil
	}
}
