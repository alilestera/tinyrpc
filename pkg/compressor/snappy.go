package compressor

import (
	"bytes"
	"errors"
	"github.com/golang/snappy"
	"io"
)

// SnappyCompressor implements the Compressor interface
type SnappyCompressor struct {
}

// Zip .
func (c SnappyCompressor) Zip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := snappy.NewBufferedWriter(buf)
	defer w.Close()
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unzip .
func (c SnappyCompressor) Unzip(data []byte) ([]byte, error) {
	r := snappy.NewReader(bytes.NewBuffer(data))
	data, err := io.ReadAll(r)
	if err != nil && !errors.Is(err, io.EOF) && errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	return data, nil
}
