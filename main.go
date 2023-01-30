package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
)

func main() {

	input, err := zip.OpenReader("comic.cbz")
	if err != nil {
		log.Fatalln("Failed to open file:", err)
	}

	for _, file := range input.File {
		fmt.Println("name: ", file.Name)
		if file.Name == "ComicInfo.xml" {
			fmt.Println("Has metadata")
			err := showComicInfo(file)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

type YesNo string
type AgeRating string
type Manga string
type Rating float32

type ComicPageType string

type ComicPageInfo struct {
	Image       int            `xml:",attr"`
	Type        *ComicPageType `xml:",attr"`
	DoublePage  *bool          `xml:",attr"`
	ImageSize   *int64         `xml:",attr"`
	Key         *string        `xml:",attr"`
	Bookmark    *string        `xml:",attr"`
	ImageWidth  *int           `xml:",attr"`
	ImageHeight *int           `xml:",attr"`
}

type ArrayOfComicPageInfo []ComicPageInfo

type ComicInfo struct {
	Title               *string
	Series              *string
	Number              *string
	Count               *int
	Volume              *int
	AlternativeSeries   *string
	AlternativeNumber   *string
	AlternativeCount    *int
	Summary             *string
	Notes               *string
	Year                *int
	Month               *int
	Day                 *int
	Writer              *string
	Penciller           *string
	Inker               *string
	Colorist            *string
	Letterer            *string
	CoverArtist         *string
	Editor              *string
	Publisher           *string
	Imprint             *string
	Genre               *string
	Web                 *string
	PageCount           *int
	LanguageISO         *string
	Format              *string
	BlackAndWhite       *YesNo
	Manga               *Manga
	Characters          *string
	Teams               *string
	Locations           *string
	ScanInformation     *string
	StoryArc            *string
	SeriesGroup         *string
	AgeRating           *AgeRating
	Pages               ArrayOfComicPageInfo
	CommunityRating     *Rating
	MainCharacterOrTeam *string
	Review              *string
}

func showComicInfo(file *zip.File) error {
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
