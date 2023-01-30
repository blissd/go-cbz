package main

// Models the ComicInfo.xml schema file.
// Schema is currently at version 2.0: https://github.com/anansi-project/comicinfo/blob/main/schema/v2.0/ComicInfo.xsd

type YesNo string
type AgeRating string
type Manga string
type Rating float32

type ComicPageType string

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
	Count               int                  `xml:",omitempty"`
	Volume              int                  `xml:",omitempty"`
	AlternativeSeries   string               `xml:",omitempty"`
	AlternativeNumber   string               `xml:",omitempty"`
	AlternativeCount    int                  `xml:",omitempty"`
	Summary             string               `xml:",omitempty"`
	Notes               string               `xml:",omitempty"`
	Year                int                  `xml:",omitempty"`
	Month               int                  `xml:",omitempty"`
	Day                 int                  `xml:",omitempty"`
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
	PageCount           int                  `xml:",omitempty"`
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
