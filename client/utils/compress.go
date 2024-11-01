package utils

import (
	"bytes"
	"io"

	"github.com/klauspost/compress/zstd"
)

func Compress(b []byte) []byte {
	var buf bytes.Buffer
	enc, err := zstd.NewWriter(&buf)
	if err != nil {
		return nil
	}
	_, err = enc.Write(b)
	if err != nil {
		return nil
	}
	enc.Close()

	return buf.Bytes()
}

func Decompress(b []byte) ([]byte, error) {
	dec, err := zstd.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer dec.Close()

	return io.ReadAll(dec)
}
