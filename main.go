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

			info, err := readComicInfo(file)

			marshal, err := xml.MarshalIndent(&info, "", " ")
			if err != nil {
				return fmt.Errorf("failed to XML marshal ComicInfo.xml: %w", err)
			}

			fmt.Println(string(marshal))
		}
	}

	return nil
}

func readComicInfo(file *zip.File) (*ComicInfo, error) {
	r, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open zip %s for reading: %w", file.Name, err)
	}
	defer r.Close()

	bs, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", file.Name, err)
	}

	info := ComicInfo{}
	err = xml.Unmarshal(bs, &info)

	if err != nil {
		return nil, fmt.Errorf("failed to XML unmarshal %s: %w", file.Name, err)
	}

	return &info, nil
}

func setComicInfoField(_ context.Context, args []string) error {
	fmt.Println("meta set:", args)

	nameAndValue := strings.Split(args[0], "=")

	input, err := zip.OpenReader(args[1])
	if err != nil {
		log.Fatalln("Failed to open file:", err)
	}
	defer input.Close()

	for _, file := range input.File {
		if file.Name == "ComicInfo.xml" {
			if err != nil {
				return fmt.Errorf("failed to show ComicInfo.xml: %w", err)
			}

			info, err := readComicInfo(file)

			updater := SetField(nameAndValue[0], nameAndValue[1])
			updater(info)

			marshal, err := xml.MarshalIndent(&info, "", " ")
			if err != nil {
				return fmt.Errorf("failed to XML marshal ComicInfo.xml: %w", err)
			}

			fmt.Println(string(marshal))
		}
	}

	return nil
}

type FieldUpdater func(info *ComicInfo) error

func SetField(name string, value string) FieldUpdater {
	return func(info *ComicInfo) error {
		rv := reflect.Indirect(reflect.ValueOf(info))
		f := rv.FieldByName(name)
		f.SetString(value)
		return nil
	}
}
