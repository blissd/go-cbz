package main

// Models the ComicInfo.xml schema file.
// Schema is currently at version 2.0: https://github.com/anansi-project/comicinfo/blob/main/schema/v2.0/ComicInfo.xsd

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
