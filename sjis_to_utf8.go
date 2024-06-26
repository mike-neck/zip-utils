package ziputils

import (
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"hash/fnv"
	"time"
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

func CalculateHash(index int, z ZipEntry) uint32 {
	hf := fnv.New32a()
	// time.Time 型の f.Modified を文字列にする
	modified := z.Modified.Format("2006-01-02T15:04:05")
	entry := fmt.Sprintf("%04d:%-60s:%s", index, z.Name, modified)
	_, _ = hf.Write([]byte(entry))
	hash := hf.Sum32()
	return hash
}

type ZipEntry struct {
	Name     string
	Modified time.Time
}
