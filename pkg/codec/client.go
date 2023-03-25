package codec

import (
	"bufio"
	"github.com/alilestera/tinyrpc/pkg/compressor"
	"github.com/alilestera/tinyrpc/pkg/header"
	"github.com/alilestera/tinyrpc/pkg/serializer"
	"hash/crc32"
	"io"
	"net/rpc"
	"sync"
)

type clientCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	compressor compressor.CompressType // rpc compress type(raw,gzip,snappy,zlib)
	serializer serializer.Serializer
	response   header.ResponseHeader // rpc response header
	mutex      sync.Mutex            // protect pending map
	pending    map[uint64]string
}

// WriteRequest Write the rpc request header and body to the io stream
func (c *clientCodec) WriteRequest(r *rpc.Request, param any) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()

	// check compressor whether exist
	if _, ok := compressor.Compressors[c.compressor]; !ok {
		return NotFoundCompressorError
	}
	// call serializer
	reqBody, err := c.serializer.Marshal(param)
	if err != nil {
		return err
	}
	// zip request body
	compressedReqBody, err := compressor.Compressors[c.compressor].Zip(reqBody)
	if err != nil {
		return err
	}
	return c.buildSendRequest(r, compressedReqBody)
}

// buildSendRequest build request and send frame
// reqBody is compressed
func (c *clientCodec) buildSendRequest(r *rpc.Request, reqBody []byte) error {
	// get header from pool
	h := header.RequestPool.Get().(*header.RequestHeader)
	defer func() {
		h.ResetHeader()
		header.RequestPool.Put(h)
	}()

	// fill header
	h.ID = r.Seq
	h.Method = r.ServiceMethod
	h.RequestLen = uint32(len(reqBody))
	h.CompressType = c.compressor
	h.Checksum = crc32.ChecksumIEEE(reqBody)

	// send header
	if err := sendFrame(c.w, h.Marshal()); err != nil {
		return err
	}
	// send body
	if err := write(c.w, reqBody); err != nil {
		return err
	}
	c.w.(*bufio.Writer).Flush()

	return nil
}

// ReadResponseHeader read the rpc response header from the io stream
func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	// reset clientCodec header
	c.response.ResetHeader()
	// read response header
	data, err := recvFrame(c.r)
	if err != nil {
		return err
	}
	// unmarshal header
	err = c.response.Unmarshal(data)
	if err != nil {
		return err
	}
	// fill response
	c.mutex.Lock()
	r.Seq = c.response.ID
	r.Error = c.response.Error
	r.ServiceMethod = c.pending[r.Seq]
	delete(c.pending, r.Seq)
	c.mutex.Unlock()
	return nil
}

// ReadResponseBody read the rpc response body from the io stream
func (c *clientCodec) ReadResponseBody(param any) error {
	if param == nil {
		if c.response.ResponseLen != 0 { // discard excess
			if err := read(c.r, make([]byte, c.response.ResponseLen)); err != nil {
				return err
			}
		}
		return nil
	}
	// read ResponseLen length bytes
	respBody := make([]byte, c.response.ResponseLen)
	err := read(c.r, respBody)
	if err != nil {
		return err
	}

	// check
	if c.response.Checksum != 0 {
		if crc32.ChecksumIEEE(respBody) != c.response.Checksum {
			return UnexpectedChecksumError
		}
	}
	// check compressor whether is match
	if c.response.GetCompressType() != c.compressor {
		return CompressorTypeMismatchError
	}
	// unzip response body
	resp, err := compressor.Compressors[c.response.GetCompressType()].Unzip(respBody)
	if err != nil {
		return err
	}

	// unmarshal unzip response body
	return c.serializer.Unmarshal(resp, param)
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}

// NewClientCodec Create a new client codec
func NewClientCodec(conn io.ReadWriteCloser, compressType compressor.CompressType, serializer serializer.Serializer) rpc.ClientCodec {
	return &clientCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		compressor: compressType,
		serializer: serializer,
		pending:    make(map[uint64]string),
	}
}
