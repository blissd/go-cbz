package main

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/peterbourgon/ff/v3/ffcli"
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

func pipeline(zipFileName string, action Action) error {
	input, err := zip.OpenReader(zipFileName)
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
			err = action(info)
			if err != nil {
				return fmt.Errorf("failed to apply action to ComicInfo.xml: %w", err)
			}
		}
	}

	return nil
}

func showComicInfo(_ context.Context, args []string) error {
	return pipeline(args[0], func(info *ComicInfo) error {
		marshal, err := xml.MarshalIndent(&info, "", " ")
		if err != nil {
			return fmt.Errorf("failed to XML marshal ComicInfo.xml: %w", err)
		}

		fmt.Println(string(marshal))
		return nil
	})
}

func setComicInfoField(_ context.Context, args []string) error {
	fmt.Println("meta set:", args)

	nameAndValue := strings.Split(args[0], "=")

	return pipeline(args[1], setField(nameAndValue[0], nameAndValue[1]))
}

// Action performs an action on a ComicInfo, such as printing a value, setting a value, or removing a value.
type Action func(info *ComicInfo) error

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
