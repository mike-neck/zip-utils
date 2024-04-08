package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"github.com/mike-neck/zip-utils"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UserOption 3つのパラメーターをまとめたデータ
type UserOption struct {
	ZipFilename   string
	FileToExtract string
	ExtractDir    string
}

// パラメーターをパースして、3つのパラメーターをまとめたデータを返す
// エラーが発生した場合は、エラーを返す
func parseParams() (*UserOption, error) {
	// zipファイル名を指定するパラメーター
	zipFilename := flag.String("i", "", "input zip file")

	// zipファイルから取り出すファイル名を指定するパラメーター
	fileToExtract := flag.String("f", "", "file to extract from zip")

	// 展開先のディレクトリを指定するパラメーター
	extractDir := flag.String("d", "", "directory to extract file to")

	// コマンドライン引数を解析する
	flag.Parse()

	// 各パラメーターの値を取得する
	if *zipFilename == "" || *fileToExtract == "" || *extractDir == "" {
		slice := make([]string, 0)
		slice = append(slice, "usage: pick-zip -i <zip file> -f <file to extract> -d <extract directory>\n")
		if *zipFilename == "" {
			slice = append(slice, "-z <zip file>")
		} else {
			slice = append(slice, fmt.Sprintf("-z %s", *zipFilename))
		}
		if *fileToExtract == "" {
			slice = append(slice, "-f <file to extract>")
		} else {
			slice = append(slice, fmt.Sprintf("-f %s", *fileToExtract))
		}
		if *extractDir == "" {
			slice = append(slice, "-d <extract directory>")
		} else {
			slice = append(slice, fmt.Sprintf("-d %s", *extractDir))
		}
		return nil, errors.New(strings.Join(slice, "\n"))
	}

	return &UserOption{*zipFilename, *fileToExtract, *extractDir}, nil
}

func main() {
	uo, err := parseParams()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	zipfile := uo.ZipFilename
	filename := uo.FileToExtract
	destdir := uo.ExtractDir

	if i, err := os.Stat(destdir); err != nil || !i.IsDir() {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(os.Stderr, "Error: destination directory %s does not exist.\n", destdir)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(2)
	}

	// zipファイルを開く
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(os.Stderr, "Error: zip file %s does not exist.\n", zipfile)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(3)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer r.Close()

	err = pickupEntry(&r.Reader, filename, destdir)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(10)
	} else {
		fmt.Println(filename, "picked up")
	}
}

func pickupEntry(r *zip.Reader, targetFile string, destdir string) error {
	for _, f := range r.File {
		// エントリーのファイル名をShiftJISからUTF-8に変換する
		entryFilename, err := ziputils.SJISToUtf8(f.Name)
		if err != nil {
			return fmt.Errorf("error at PickupEntry#SJISToUtf8: %w", err)
		}

		// 展開するファイルと一致する場合に展開する
		if entryFilename == targetFile {
			_, file := filepath.Split(targetFile)
			err = extractFile(f, file, destdir)
			if err != nil {
				return fmt.Errorf("error at PickupEntry#extractFile: %w", err)
			}
			fmt.Printf("%s was extracted to %s\n", targetFile, file)
			return nil
		}
	}
	return fmt.Errorf("file not found: %s", targetFile)
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
