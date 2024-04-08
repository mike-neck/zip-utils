package main

import (
	"archive/zip"
	"fmt"
	"github.com/mike-neck/zip-utils"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func main() {
	// list-zip コマンド
	// github.com/urfave/cli/v2 を使ってコマンドラインアプリケーションを組み立てる
	// オプション・オプションなし引数を使って ListOption 構造体を組み立てる
	// オプション指定なしのパラメーターを ListOption の FilePath(必須)に指定する(For implementation Do not use cli.StringFlag)
	// -r/--raw-string オプションが ListOption の RawString(指定無しの場合は falseで、ファイルの名前を ShiftJIS->UTF-8 に変換して表示する)
	// -s/--show-hash オプションが ListOption の ShowHash(指定無しの場合は false)
	// 組み立てられた ListOption の listZipFileEntries() 関数を呼び出すことがアプリケーション本体の内容
	// 上記の仕様を満たすプログラムを書いてください。
	rawString := false
	showHash := false
	app := &cli.App{
		Name:  "list-zip",
		Usage: "List the entries in a zip file",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "raw-string",
				Aliases:     []string{"r"},
				Usage:       "Display file names in raw string format (default false: converts file names from ShiftJIS to UTF-8)",
				Destination: &rawString,
			},
			&cli.BoolFlag{
				Name:        "show-hash",
				Aliases:     []string{"s"},
				Usage:       "Display hash value of each entry",
				Destination: &showHash,
			},
		},
		Action: func(c *cli.Context) error {
			lo := ListOption{
				FilePath:  c.Args().First(),
				RawString: rawString,
				ShowHash:  showHash,
			}
			return lo.listZipFileEntries()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
}

type ListOption struct {
	FilePath  string
	RawString bool
	ShowHash  bool
}

func (lo ListOption) String() string {
	return fmt.Sprintf("ListOption[FilePath: %s, RawString: %v, ShowHash: %v]", lo.FilePath, lo.RawString, lo.ShowHash)
}

func (lo ListOption) listZipFileEntries() error {
	r, err := zip.OpenReader(lo.FilePath)
	if err != nil {
		return err
	}
	defer r.Close()

	// エントリーの一覧を取得して表示する
	es := make([]string, 0)
	for i, f := range r.File {
		// エントリーのファイル名をShiftJISからUTF-8に変換する
		filename, err := lo.ExtractFileName(f.Name)
		if err != nil {
			es = append(es, fmt.Sprintf("    %s", err))
		} else {
			outputs := lo.MakePrintFormat(i, filename, f.FileHeader)
			fmt.Println(outputs)
		}
	}
	if len(es) > 0 {
		all := strings.Join(es, "\n")
		return fmt.Errorf("errors occurred during conversion: \n%s", all)
	}
	return nil
}

func (lo ListOption) ExtractFileName(name string) (string, error) {
	if lo.RawString {
		return name, nil
	} else {
		return charsets.SJISToUtf8(name)
	}
}

func (lo ListOption) MakePrintFormat(index int, name string, f zip.FileHeader) string {
	outputs := make([]string, 0)
	if lo.ShowHash {
		hash := CalculateHash(index, f)
		outputs = append(outputs, fmt.Sprintf("%08x", hash))
	}
	outputs = append(outputs, name)
	return strings.Join(outputs, " ")
}
