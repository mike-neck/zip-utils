package main

import (
	"archive/zip"
	"fmt"
	"github.com/mike-neck/zip-utils"
	"log"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		_, _ = fmt.Fprintln(os.Stderr, "no file specified")
		os.Exit(1)
	} else if 2 < len(os.Args) {
		_, _ = fmt.Fprintln(os.Stderr, "only a single file is acceptable.")
		os.Exit(2)
	}

	// zipファイルを開く
	r, err := zip.OpenReader(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// エントリーの一覧を取得して表示する
	for _, f := range r.File {
		// エントリーのファイル名をShiftJISからUTF-8に変換する
		filename, err := charsets.SJISToUtf8(f.Name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(filename)
	}
}
