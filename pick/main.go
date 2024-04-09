package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/mike-neck/zip-utils"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TargetFile 解凍するファイル
type TargetFile struct {
	// ArchiveName はファイル名称によるファイル指定方法
	ArchiveName string
	// ArchiveHash はハッシュ値によるファイル指定方法
	ArchiveHash string
	// ExtractName は解凍したファイルの名前(デフォルトは ArchiveName)
	ExtractName string
}

// パラメーターをパースして、3つのパラメーターをまとめたデータを返す
// エラーが発生した場合は、エラーを返す
func parseParams() (*UserOption, []cli.Flag) {
	var uo UserOption
	return &uo, []cli.Flag{
		&cli.PathFlag{
			Name:        "i",
			Value:       "",
			Usage:       "input zip file",
			Destination: &uo.ZipFilename,
		},
		&cli.PathFlag{
			Name:        "f",
			Value:       "",
			Usage:       "file to extract from zip",
			Destination: &uo.FileToExtract.ArchiveName,
		},
		&cli.PathFlag{
			Name:        "d",
			Value:       "",
			Usage:       "directory to extract file to",
			Destination: &uo.ExtractDir,
		},
	}
}

func main() {
	uo, flg := parseParams()

	app := cli.App{
		Name:  "pick-zip",
		Usage: "extracts a single file from zip archive",
		Flags: flg,
		Action: func(context *cli.Context) error {
			return uo.DoMain()
		},
		Before: func(context *cli.Context) error {
			return uo.Validate()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}

// UserOption 3つのパラメーターをまとめたデータ
type UserOption struct {
	ZipFilename   string
	FileToExtract TargetFile
	ExtractDir    string
}

func (uo UserOption) DoMain() error {
	err := uo.PrepareDestDir()
	if err != nil {
		return fmt.Errorf("error at preparing dest dir: %w", err)
	}

	// zipファイルを開く
	r, err := uo.OpenZipFile()
	if err != nil {
		return fmt.Errorf("error at openning zip file: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer r.Close()

	err = uo.PickUpEntry(&r.Reader)
	if err != nil {
		return fmt.Errorf("error at picking up file: %w", err)
	}
	fmt.Println(uo.FileToExtract, "picked up")
	return nil
}

func (uo UserOption) Validate() error {
	slice := make([]string, 0)
	slice = append(slice, "missing parameters")
	if uo.ZipFilename == "" {
		slice = append(slice, "-i <zip file>")
	}
	if uo.FileToExtract.ArchiveName == "" {
		slice = append(slice, "-f <file to extract>")
	}
	if uo.ExtractDir == "" {
		slice = append(slice, "-d <directory to extract file to>")
	}
	if len(slice) > 1 {
		return errors.New(strings.Join(slice, "\n"))
	}
	return nil
}

func (uo UserOption) PrepareDestDir() error {
	if i, err := os.Stat(uo.ExtractDir); err != nil || !i.IsDir() {
		if os.IsNotExist(err) {
			err := os.MkdirAll(uo.ExtractDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create dir[%s]: %w", uo.ExtractDir, err)
			}
		} else if i.IsDir() {
			return fmt.Errorf("%s is a file, not directory", uo.ExtractDir)
		} else {
			return fmt.Errorf("failed to get dir[%s]: %w", uo.ExtractDir, err)
		}
	}
	return nil
}

func (uo UserOption) OpenZipFile() (*zip.ReadCloser, error) {
	r, err := zip.OpenReader(uo.ZipFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to open a zip file[%s]: file does not exist", uo.ZipFilename)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(3)
	}
	return r, nil
}

func (uo UserOption) PickUpEntry(r *zip.Reader) error {
	for _, f := range r.File {
		// エントリーのファイル名をShiftJISからUTF-8に変換する
		entryFilename, err := ziputils.SJISToUtf8(f.Name)
		if err != nil {
			return fmt.Errorf("error at PickupEntry#SJISToUtf8: %w", err)
		}

		// 展開するファイルと一致する場合に展開する
		if entryFilename == uo.FileToExtract.ArchiveName {
			_, file := filepath.Split(uo.FileToExtract.ArchiveName)
			err = extractFile(f, file, uo.ExtractDir)
			if err != nil {
				return fmt.Errorf("error at PickupEntry#extractFile: %w", err)
			}
			fmt.Printf("%s was extracted to %s\n", uo.FileToExtract, file)
			return nil
		}
	}
	return fmt.Errorf("file not found: %s", uo.FileToExtract)
}

func extractFile(f *zip.File, targetFile string, destdir string) error {
	r, err := f.Open()
	if err != nil {
		return fmt.Errorf("error at extractFile#zipFile.Open: %w", err)
	}
	destPath := filepath.Join(destdir, targetFile)
	w, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error at extractFile#os.create: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer w.Close()

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("error at extractFile#io.copy: %w", err)
	}
	err = os.Chmod(destPath, f.Mode())
	if err != nil {
		return fmt.Errorf("error at extractFile#os.chmod: %w", err)
	}
	return nil
}
