package renamecmd

import (
	"github.com/blissd/cbz/model"
	"io"
	"testing"
)

func Test_config_inferFileName(t *testing.T) {
	type args struct {
		c *model.ComicInfo
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Series v01", args{&model.ComicInfo{Series: "Series", Volume: 1}}, "Series v01", false},
		{"Series v01 #2", args{&model.ComicInfo{Series: "Series", Volume: 1, Number: "2"}}, "Series v01 #2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{
				out: io.Discard,
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
