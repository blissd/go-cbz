package main

import "testing"

func TestYesNo_validate(t *testing.T) {
	tests := []struct {
		name    string
		v       YesNo
		wantErr bool
	}{
		{"Blank", "Yes", false},
		{"Unknown", "Unknown", false},
		{"Yes", "Yes", false},
		{"No", "No", false},
		{"Invalid", "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
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
		{"Blank", "Yes", false},
		{"Unknown", "Unknown", false},
		{"No", "No", false},
		{"Yes", "Yes", false},
		{"YesAndRightToLeft", "YesAndRightToLeft", false},
		{"Invalid", "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
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
		{"", "", false},
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
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
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
		{"", "", false},
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
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
