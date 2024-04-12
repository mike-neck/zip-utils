package main

import (
	_ "embed"
	"testing"
)

//go:embed test/data/utf-8.txt
var testListOptionExtractFileNameUTF8 string

//go:embed test/data/shift-jis.txt
var testListOptionExtractFileNameShiftJIS string

func TestListOption_ExtractFileName(t *testing.T) {
	type fields struct {
		RawString  bool
		AutoDetect bool
	}
	type args struct {
		name   string
		isUTF8 bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "run with --raw-string",
			fields: fields{
				RawString:  true,
				AutoDetect: false,
			},
			args: args{
				name:   testListOptionExtractFileNameUTF8,
				isUTF8: false,
			},
			want:    "UTF8文字列",
			wantErr: false,
		},
		{
			name: "run with --auto-detect on UTF-8",
			fields: fields{
				RawString:  false,
				AutoDetect: true,
			},
			args: args{
				name:   testListOptionExtractFileNameUTF8,
				isUTF8: true,
			},
			want:    "UTF8文字列",
			wantErr: false,
		},
		{
			name: "run with no option on ShiftJIS",
			fields: fields{
				RawString:  false,
				AutoDetect: false,
			},
			args: args{
				name:   testListOptionExtractFileNameShiftJIS,
				isUTF8: false,
			},
			want:    "ShiftJIS文字列",
			wantErr: false,
		},
		{
			name: "run with --auto-detect on ShiftJIS",
			fields: fields{
				RawString:  false,
				AutoDetect: true,
			},
			args: args{
				name:   testListOptionExtractFileNameShiftJIS,
				isUTF8: false,
			},
			want:    "ShiftJIS文字列",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lo := ListOption{
				RawString:  tt.fields.RawString,
				AutoDetect: tt.fields.AutoDetect,
			}
			got, err := lo.ExtractFileName(tt.args.name, tt.args.isUTF8)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractFileName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
