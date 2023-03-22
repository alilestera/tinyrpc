package compressor

// RawCompressor implements the Compressor interface
type RawCompressor struct {
}

// Zip .
func (c RawCompressor) Zip(data []byte) ([]byte, error) {
	return data, nil
}

// Unzip .
func (c RawCompressor) Unzip(data []byte) ([]byte, error) {
	return data, nil
}

var _ Compressor = (*RawCompressor)(nil)
