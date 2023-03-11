package renamecmd

import (
	"github.com/blissd/cbz/model"
	"io"
	"testing"
)

func Test_config_inferFileName(t *testing.T) {
	type fields struct {
		out          io.Writer
		includeTitle bool
	}
	type args struct {
		c *model.ComicInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{"Series v01", fields{}, args{&model.ComicInfo{Series: "Series", Title: "Title", Volume: 1}}, "Series v01", false},
		{"Series v01 - Title", fields{includeTitle: true}, args{&model.ComicInfo{Series: "Series", Title: "Title", Volume: 1}}, "Series v01 - Title", false},
		{"Series v01 #2", fields{}, args{&model.ComicInfo{Series: "Series", Volume: 1, Number: "2"}}, "Series v01 #2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{
				out:          io.Discard,
				includeTitle: tt.fields.includeTitle,
			}
			got, err := cfg.inferFileName(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("inferFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("inferFileName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
