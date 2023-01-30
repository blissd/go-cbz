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
	"reflect"
	"strings"
)

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
		ShortHelp:  "Set an field value. e.g., cbz meta set AgeRating=M",
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
func process(zipFileName string, action Action) error {
	input, err := zip.OpenReader(zipFileName)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	output, err := os.Create("out.cbz")
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer output.Close()

	outputZip := zip.NewWriter(output)
	defer outputZip.Close()

	for _, file := range input.File {
		w, err := outputZip.Create(file.Name)
		if err != nil {
			return fmt.Errorf("failed creating file in zip archive: %w", err)
		}

		if file.Name == "ComicInfo.xml" {
			if err != nil {
				return fmt.Errorf("failed to show ComicInfo.xml: %w", err)
			}

			info, err := unmarshallComicInfoXml(file)
			if err != nil {
				return fmt.Errorf("failed to unmarshal ComicInfo.xml: %w", err)
			}

			err = action(info)
			if err != nil {
				return fmt.Errorf("failed to apply action to ComicInfo.xml: %w", err)
			}

			bs, err := xml.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal ComicInfo.xml: %w", err)
			}

			if _, err = w.Write(bs); err != nil {
				return fmt.Errorf("failed to write ComicInfo.xml: %w", err)
			}
		} else {
			err = copyFile(w, file)
			if err != nil {
				return fmt.Errorf("failed to add %s: %w", file.Name, err)
			}
		}
	}

	return nil
}

func copyFile(w io.Writer, src *zip.File) error {
	r, err := src.Open()
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", src.Name, err)
	}
	defer r.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("failed to copy %s: %w", src.Name, err)
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

func printXml(info *ComicInfo) error {
	marshal, err := xml.MarshalIndent(&info, "", " ")
	if err != nil {
		return fmt.Errorf("failed to XML marshal ComicInfo.xml: %w", err)
	}

	fmt.Println(string(marshal))
	return nil
}

func setComicInfoField(_ context.Context, args []string) error {
	fmt.Println("meta set:", args)

	zipFileName := args[len(args)-1]
	args = args[:len(args)-1]
	actions := make([]Action, len(args), len(args)+1) // leave space for (optional) printXml action
	for i, v := range args {
		fmt.Println(v)
		nameAndValue := strings.Split(v, "=")
		fmt.Println("split:", nameAndValue)
		actions[i] = setField(nameAndValue[0], nameAndValue[1])
	}

	actions = append(actions, printXml)

	return process(zipFileName, join(actions))
}

// Action performs an action on a ComicInfo, such as printing a value, setting a value, or removing a value.
type Action func(info *ComicInfo) error

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
		case bool:
			f.SetBool(v)
		default:
			return fmt.Errorf("unsupported data type: %v", v)
		}
		return nil
	}
}
