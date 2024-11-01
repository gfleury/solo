package utils

import (
	"bytes"
	"compress/gzip"
	"io"
)

func Compress(b []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(b)
	gz.Close()

	return buf.Bytes()
}

func Decompress(b []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	return io.ReadAll(r)
}
