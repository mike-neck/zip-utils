package main

import "testing"

func TestUserOption_Validate(t *testing.T) {
	type fields struct {
		ZipFilename   string
		FileToExtract TargetFile
		ExtractDir    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test for fileToExtract when specify with archiveName[SUCCESS]",
			fields: fields{
				ZipFilename: "test.zip",
				FileToExtract: TargetFile{
					ArchiveName: "archive",
					ArchiveHash: "",
					ExtractName: "",
				},
				ExtractDir: "dir",
			},
			wantErr: false,
		},
		{
			name: "test for fileToExtract when specify with archiveName[FAIL]",
			fields: fields{
				ZipFilename: "test.zip",
				FileToExtract: TargetFile{
					ArchiveName: "",
					ArchiveHash: "",
					ExtractName: "",
				},
				ExtractDir: "dir",
			},
			wantErr: true,
		},
		{
			name: "test for fileToExtract when specify with hash then [SUCCESS]",
			fields: fields{
				ZipFilename: "test.zip",
				FileToExtract: TargetFile{
					ArchiveName: "",
					ArchiveHash: "1a2b3c4d5e",
					ExtractName: "extract",
				},
				ExtractDir: "dir",
			},
			wantErr: false,
		},
		{
			name: "test for fileToExtract when specify with hash then [SUCCESS]",
			fields: fields{
				ZipFilename: "test.zip",
				FileToExtract: TargetFile{
					ArchiveName: "",
					ArchiveHash: "1a2b3c4d5e",
					ExtractName: "",
				},
				ExtractDir: "dir",
			},
			wantErr: true,
		},
		{
			name: "test for fileToExtract when specify with hash then [SUCCESS]",
			fields: fields{
				ZipFilename: "test.zip",
				FileToExtract: TargetFile{
					ArchiveName: "",
					ArchiveHash: "",
					ExtractName: "ext",
				},
				ExtractDir: "dir",
			},
			wantErr: true,
		},
		{
			name: "no zipFilename name[FAILURE]",
			fields: fields{
				ZipFilename: "",
				FileToExtract: TargetFile{
					ArchiveName: "",
					ArchiveHash: "1a2b3c",
					ExtractName: "ext",
				},
				ExtractDir: "dir",
			},
			wantErr: true,
		},
		{
			name: "no extractDir name[FAILURE]",
			fields: fields{
				ZipFilename: "test.zip",
				FileToExtract: TargetFile{
					ArchiveName: "",
					ArchiveHash: "1a2b3c",
					ExtractName: "ext",
				},
				ExtractDir: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uo := UserOption{
				ZipFilename:   tt.fields.ZipFilename,
				FileToExtract: tt.fields.FileToExtract,
				ExtractDir:    tt.fields.ExtractDir,
			}
			if err := uo.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type testTargetFilelGetFileNameGetTimeFunc string

func (f *testTargetFilelGetFileNameGetTimeFunc) GetTime() string {
	return string(*f)
}

func TestTargetFile_GetFileName(t *testing.T) {
	type fields struct {
		ArchiveName string
		ArchiveHash string
		ExtractName string
		GetTimeFunc GetTimeFunc
	}
	mockGetTimeFunc := testTargetFilelGetFileNameGetTimeFunc("test-time")
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "With ArchiveName[-> ArchiveName]",
			fields: fields{
				ArchiveName: "test-name",
				ArchiveHash: "",
				ExtractName: "",
				GetTimeFunc: nil,
			},
			want: "test-name",
		},
		{
			name: "With Hash/ExtractName[-> ExtractName]",
			fields: fields{
				ArchiveName: "",
				ArchiveHash: "1a2b3c4d5e6f",
				ExtractName: "ext-name",
				GetTimeFunc: nil,
			},
			want: "ext-name",
		},
		{
			name: "With ArchiveName/ExtractName[-> ExtractName]",
			fields: fields{
				ArchiveName: "archive-name",
				ArchiveHash: "",
				ExtractName: "ext-name",
				GetTimeFunc: nil,
			},
			want: "ext-name",
		},
		{
			name: "Without params[-> GetTimeFunc]",
			fields: fields{
				ArchiveName: "",
				ArchiveHash: "",
				ExtractName: "",
				GetTimeFunc: &mockGetTimeFunc,
			},
			want: "test-time",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := TargetFile{
				ArchiveName: tt.fields.ArchiveName,
				ArchiveHash: tt.fields.ArchiveHash,
				ExtractName: tt.fields.ExtractName,
				GetTimeFunc: tt.fields.GetTimeFunc,
			}
			if got := tf.GetFileName(); got != tt.want {
				t.Errorf("GetFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
