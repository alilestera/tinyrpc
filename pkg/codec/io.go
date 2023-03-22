package codec

import (
	"encoding/binary"
	"io"
	"net"
)

func sendFrame(w io.Writer, data []byte) error {
	var size [binary.MaxVarintLen64]byte

	if len(data) == 0 {
		n := binary.PutUvarint(size[:], 0)
		return write(w, size[:n])
	}

	n := binary.PutUvarint(size[:], uint64(len(data)))
	if err := write(w, size[:n]); err != nil {
		return err
	}
	return write(w, data)
}

func recvFrame(r io.Reader) (data []byte, err error) {
	size, err := binary.ReadUvarint(r.(io.ByteReader))
	if err != nil {
		return nil, err
	}
	if size != 0 {
		data = make([]byte, size)
		if err = read(r, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func write(w io.Writer, data []byte) error {
	for i := 0; i < len(data); {
		n, err := w.Write(data[i:])
		if err != nil {
			if _, ok := err.(net.Error); !ok {
				return err
			}
		}
		i += n
	}
	return nil
}

func read(r io.Reader, data []byte) error {
	for i := 0; i < len(data); {
		n, err := r.Read(data[i:])
		if err != nil {
			if _, ok := err.(net.Error); !ok {
				return err
			}
		}
		i += n
	}
	return nil
}
