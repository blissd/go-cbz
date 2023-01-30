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
)

func main() {

	metaShow := &ffcli.Command{
		Name:       "show",
		ShortUsage: "cbz meta show <comic.cbz>",
		ShortHelp:  "Show the raw ComicInfo.xml file in a comic archive",
		Exec:       showComicInfo,
	}

	meta := &ffcli.Command{
		Name:        "meta",
		ShortUsage:  "cbz meta <subcommand>",
		ShortHelp:   "Display and manipulate ComicInfo.xml file in a comic archive",
		Subcommands: []*ffcli.Command{metaShow},
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
	fmt.Println(args)

	input, err := zip.OpenReader(args[0])
	if err != nil {
		log.Fatalln("Failed to open file:", err)
	}

	for _, file := range input.File {
		if file.Name == "ComicInfo.xml" {
			err := showComicInfoXml(file)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	return nil
}

func showComicInfoXml(file *zip.File) error {
	r, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip %s for reading: %w", file.Name, err)
	}
	defer r.Close()

	bs, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", file.Name, err)
	}

	info := ComicInfo{}
	err = xml.Unmarshal(bs, &info)
	if err != nil {
		return fmt.Errorf("failed to XML unmarshal %s: %w", file.Name, err)
	}

	//fmt.Println("raw info:", info)

	marshal, err := xml.MarshalIndent(&info, "", " ")
	if err != nil {
		return fmt.Errorf("failed to XML marshal %s: %w", file.Name, err)
	}

	fmt.Println("marshalled XML:", string(marshal))

	return nil
}
