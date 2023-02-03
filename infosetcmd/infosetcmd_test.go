package infosetcmd

import (
	"cbz/model"
	"reflect"
	"testing"
)

func Test_setField(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name string
		args args
		want model.ComicInfo
	}{
		{
			name: "AgeRating",
			args: args{"AgeRating", "M"},
			want: model.ComicInfo{
				AgeRating: "M",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := model.ComicInfo{}
			setter := setField(tt.args.name, tt.args.value)
			err := setter(&info)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(info, tt.want) {
				t.Errorf("setField() = %v, want %v", info, tt.want)
			}
		})
	}
}
