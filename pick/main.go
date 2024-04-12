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
	"time"
)

type GetTimeFunc interface {
	GetTime() string
}

// TargetFile 解凍するファイル
type TargetFile struct {
	// ArchiveName はファイル名称によるファイル指定方法
	ArchiveName string
	// ArchiveHash はハッシュ値によるファイル指定方法
	ArchiveHash string
	// ExtractName は解凍したファイルの名前(デフォルトは ArchiveName)
	ExtractName string

	// GetTimeFunc は現在時刻を[分]まで取得する関数(デフォルトは nil )
	GetTimeFunc
}

var (
	ExtractNameMissing = errors.New("missing extract name")
	IdentifierMissing  = errors.New("missing identifier, archive-name or archive-hash required")
)

func (tf TargetFile) Validate() error {
	if tf.ArchiveName != "" {
		return nil
	}
	if tf.ArchiveHash != "" && tf.ExtractName != "" {
		return nil
	}
	if tf.ArchiveHash != "" && tf.ExtractName == "" {
		return ExtractNameMissing
	}
	return IdentifierMissing
}

func (tf TargetFile) Matches(index int, zipEntry *zip.File) (bool, error) {
	if tf.ArchiveName != "" {
		if tf.ArchiveName == zipEntry.Name {
			return true, nil
		}
		utf8, err := ziputils.SJISToUtf8(zipEntry.Name)
		if err != nil {
			return false, err
		}
		return tf.ArchiveName == utf8, nil
	}
	entry := ziputils.ZipEntry{
		Name:     zipEntry.Name,
		Modified: zipEntry.Modified,
	}
	hash := ziputils.CalculateHash(index, entry)
	h := fmt.Sprintf("%x", hash)
	return tf.ArchiveHash == h, nil
}

func (tf TargetFile) GetFileName() string {
	if tf.ExtractName != "" {
		return tf.ExtractName
	}
	if tf.ArchiveName != "" {
		_, fileName := filepath.Split(tf.ArchiveName)
		return fileName
	}
	if tf.GetTimeFunc == nil {
		now := time.Now()
		return now.Format("2006-01-02T1504-07")
	} else {
		return tf.GetTime()
	}
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
			Name:        "s",
			Value:       "",
			Usage:       "a hash of a file to extract from zip",
			Destination: &uo.FileToExtract.ArchiveHash,
		},
		&cli.PathFlag{
			Name:        "n",
			Value:       "",
			Usage:       "a name of extracted file",
			Destination: &uo.FileToExtract.ExtractName,
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
	err := uo.FileToExtract.Validate()
	if err != nil {
		switch {
		case errors.Is(err, IdentifierMissing):
			slice = append(slice, "-f <file to extract> or -x <hash of file to extract> + -n <file name>")
			break
		case errors.Is(err, ExtractNameMissing):
			slice = append(slice, "-x option requires a more option -n <file name>")
		}
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
	for i, f := range r.File {
		// エントリーのファイル名をShiftJISからUTF-8に変換する
		matches, err := uo.FileToExtract.Matches(i, f)
		if err != nil {
			return fmt.Errorf("error at PickupEntry#SJISToUtf8: %w", err)
		}

		// 指定したファイルでない場合は次のファイルへ
		if !matches {
			continue
		}

		// 展開するファイルと一致する場合に展開する
		file := uo.FileToExtract.GetFileName()
		err = extractFile(f, file, uo.ExtractDir)
		if err != nil {
			return fmt.Errorf("error at PickupEntry#extractFile: %w", err)
		}
		fmt.Printf("%s was extracted to %s\n", uo.FileToExtract, file)
		return nil
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
