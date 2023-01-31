package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
)

// Models the ComicInfo.xml schema file.
// Schema is currently at version 2.0: https://github.com/anansi-project/comicinfo/blob/main/schema/v2.0/ComicInfo.xsd

type YesNo string

func (v YesNo) validate() error {
	switch v {
	case "":
	case "Unknown":
	case "No":
	case "Yes":
	default:
		return fmt.Errorf("invalid value for YesNo: %v", v)
	}
	return nil
}

type AgeRating string

func (v AgeRating) validate() error {
	// The order age ratings in the XSD is... somewhat random.
	// This order is taken from the Kavita source code: https://github.com/Kareadita/Kavita/blob/develop/API/Entities/Enums/AgeRating.cs
	switch v {
	case "":
	case "Unknown":
	case "Rating Pending":
	case "Early Childhood":
	case "Everyone":
	case "G":
	case "Everyone 10+":
	case "PG":
	case "Kids to Adults":
	case "Teen":
	case "MA15+":
	case "Mature 17+":
	case "M":
	case "R18+":
	case "Adults Only 18+":
	case "X18+":
	default:
		return fmt.Errorf("invalid value for AgeRating: %v", v)
	}
	return nil
}

type Manga string

func (v Manga) validate() error {
	switch v {
	case "":
	case "Unknown":
	case "No":
	case "Yes":
	case "YesAndRightToLeft":
	default:
		return fmt.Errorf("invalid value for Manga: %v", v)
	}
	return nil
}

type Rating float64

type ComicPageType string

func (v ComicPageType) validate() error {
	switch v {
	case "":
	case "FrontCover":
	case "InnerCover":
	case "Roundup":
	case "Story":
	case "Advertisement":
	case "Editorial":
	case "Letters":
	case "Preview":
	case "BackCover":
	case "Other":
	case "Deleted":
	default:
		return fmt.Errorf("invalid value for ComicPageType: %v", v)
	}
	return nil
}

type ComicPageInfo struct {
	Image       int           `xml:",attr"`
	Type        ComicPageType `xml:",attr,omitempty"`
	DoublePage  bool          `xml:",attr,omitempty"`
	ImageSize   int64         `xml:",attr,omitempty"`
	Key         string        `xml:",attr,omitempty"`
	Bookmark    string        `xml:",attr,omitempty"`
	ImageWidth  int           `xml:",attr,omitempty"`
	ImageHeight int           `xml:",attr,omitempty"`
}

type ArrayOfComicPageInfo []ComicPageInfo

type ComicInfo struct {
	Title               string               `xml:",omitempty"`
	Series              string               `xml:",omitempty"`
	Number              string               `xml:",omitempty"`
	Count               int64                `xml:",omitempty"`
	Volume              int64                `xml:",omitempty"`
	AlternativeSeries   string               `xml:",omitempty"`
	AlternativeNumber   string               `xml:",omitempty"`
	AlternativeCount    int64                `xml:",omitempty"`
	Summary             string               `xml:",omitempty"`
	Notes               string               `xml:",omitempty"`
	Year                int64                `xml:",omitempty"`
	Month               int64                `xml:",omitempty"`
	Day                 int64                `xml:",omitempty"`
	Writer              string               `xml:",omitempty"`
	Penciller           string               `xml:",omitempty"`
	Inker               string               `xml:",omitempty"`
	Colorist            string               `xml:",omitempty"`
	Letterer            string               `xml:",omitempty"`
	CoverArtist         string               `xml:",omitempty"`
	Editor              string               `xml:",omitempty"`
	Publisher           string               `xml:",omitempty"`
	Imprint             string               `xml:",omitempty"`
	Genre               string               `xml:",omitempty"`
	Web                 string               `xml:",omitempty"`
	PageCount           int64                `xml:",omitempty"`
	LanguageISO         string               `xml:",omitempty"`
	Format              string               `xml:",omitempty"`
	BlackAndWhite       YesNo                `xml:",omitempty"`
	Manga               Manga                `xml:",omitempty"`
	Characters          string               `xml:",omitempty"`
	Teams               string               `xml:",omitempty"`
	Locations           string               `xml:",omitempty"`
	ScanInformation     string               `xml:",omitempty"`
	StoryArc            string               `xml:",omitempty"`
	SeriesGroup         string               `xml:",omitempty"`
	AgeRating           AgeRating            `xml:",omitempty"`
	Pages               ArrayOfComicPageInfo `xml:",omitempty"`
	CommunityRating     Rating               `xml:",omitempty"`
	MainCharacterOrTeam string               `xml:",omitempty"`
	Review              string               `xml:",omitempty"`
}

func (v *ComicInfo) validate() error {
	if err := v.AgeRating.validate(); err != nil {
		return fmt.Errorf("invalid value for AgeRating: %w", err)
	}
	if err := v.BlackAndWhite.validate(); err != nil {
		return fmt.Errorf("invalid value for BlackAndWhite: %w", err)
	}
	if err := v.Manga.validate(); err != nil {
		return fmt.Errorf("invalid value for Manga: %w", err)
	}

	for _, p := range v.Pages {
		if err := p.Type.validate(); err != nil {
			return fmt.Errorf("invalid value for Pages.Type: %w", err)
		}
	}

	return nil
}

func unmarshallComicInfoXml(file *zip.File) (*ComicInfo, error) {
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
