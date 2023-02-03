package model

import (
	"reflect"
	"testing"
)

func TestYesNo_validate(t *testing.T) {
	tests := []struct {
		name    string
		v       YesNo
		wantErr bool
	}{
		{"Blank", "", false},
		{"Unknown", "Unknown", false},
		{"Yes", "Yes", false},
		{"No", "No", false},
		{"Invalid", "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManga_validate(t *testing.T) {
	tests := []struct {
		name    string
		v       Manga
		wantErr bool
	}{
		{"Blank", "", false},
		{"Unknown", "Unknown", false},
		{"No", "No", false},
		{"Yes", "Yes", false},
		{"YesAndRightToLeft", "YesAndRightToLeft", false},
		{"Invalid", "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgeRating_validate(t *testing.T) {
	tests := []struct {
		name    string
		v       AgeRating
		wantErr bool
	}{
		{"Blank", "", false},
		{"Unknown", "Unknown", false},
		{"Adults Only 18+", "Adults Only 18+", false},
		{"Early Childhood", "Early Childhood", false},
		{"Everyone", "Everyone", false},
		{"Everyone 10+", "Everyone 10+", false},
		{"G", "G", false},
		{"Kids to Adults", "Kids to Adults", false},
		{"M", "M", false},
		{"MA15+", "MA15+", false},
		{"Mature 17+", "Mature 17+", false},
		{"PG", "PG", false},
		{"R18+", "R18+", false},
		{"Rating Pending", "Rating Pending", false},
		{"Teen", "Teen", false},
		{"X18+", "X18+", false},
		{"Invalid", "foo", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComicPageType_validate(t *testing.T) {
	tests := []struct {
		name    string
		v       ComicPageType
		wantErr bool
	}{
		{"Blank", "", false},
		{"FrontCover", "FrontCover", false},
		{"InnerCover", "InnerCover", false},
		{"Roundup", "Roundup", false},
		{"Story", "Story", false},
		{"Advertisement", "Advertisement", false},
		{"Editorial", "Editorial", false},
		{"Letters", "Letters", false},
		{"Preview", "Preview", false},
		{"BackCover", "BackCover", false},
		{"Other", "Other", false},
		{"Deleted", "Deleted", false},
		{"Invalid", "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{"AgeRating", args{"AgeRating", "PG"}, "PG", false},
		// int fields
		{"Count", args{"Count", "1"}, int64(1), false},
		{"Volume", args{"Volume", "2"}, int64(2), false},
		{"AlternativeCount", args{"AlternativeCount", "3"}, int64(3), false},
		{"Year", args{"Year", "4"}, int64(4), false},
		{"Month", args{"Month", "5"}, int64(5), false},
		{"Day", args{"Day", "6"}, int64(6), false},
		{"PageCount", args{"PageCount", "7"}, int64(7), false},
		{"Not an int", args{"PageCount", "not an int"}, int64(0), true},
		// float fields
		{"CommunityRating", args{"CommunityRating", "1.5"}, 1.5, false},
		{"Not a float", args{"CommunityRating", "not a float"}, 0.0, true},
		// bool fields
		{"DoublePage", args{"DoublePage", "true"}, true, false},
		{"Not a bool", args{"DoublePage", "green"}, false, true},
		// everything else is a string
		{"Strings", args{"AnyRandomField", "abc"}, "abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.args.name, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Convert() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComicInfo_String(t *testing.T) {

	info := ComicInfo{
		Title:  "Great Comic",
		Series: "Great Series",
	}

	got := info.String()

	want :=
		`<ComicInfo>
 <Title>Great Comic</Title>
 <Series>Great Series</Series>
</ComicInfo>`

	if want != got {
		t.Fatalf("want: %v, got: %v", want, got)
	}
}
