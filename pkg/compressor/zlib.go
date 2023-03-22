package compressor

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
)

// ZlibCompressor implements the Compressor interface
type ZlibCompressor struct {
}

// Zip .
func (c ZlibCompressor) Zip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := zlib.NewWriter(buf)
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
func (c ZlibCompressor) Unzip(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err = io.ReadAll(r)
	if err != nil && errors.Is(err, io.EOF) && errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	return data, nil
}
