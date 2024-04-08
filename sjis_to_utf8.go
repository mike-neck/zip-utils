package charsets

import (
	"archive/zip"
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"hash/fnv"
)

// SJISToUtf8 ShiftJISからUTF-8に変換する関数
func SJISToUtf8(s string) (string, error) {
	// ShiftJISをUTF-8に変換するためのTransformerを作成する
	transformer := japanese.ShiftJIS.NewDecoder()
	utf8Bytes, _, err := transform.Bytes(transformer, []byte(s))
	if err != nil {
		return "", err
	}
	return string(utf8Bytes), nil
}
