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

	if _, ok := compressor.Compressors[c.compressor]; !ok {
		return NotFoundCompressorError
	}
	reqBody, err := c.serializer.Marshal(param)
	if err != nil {
		return err
	}
	compressedReqBody, err := compressor.Compressors[c.compressor].Zip(reqBody)
	if err != nil {
		return err
	}
	return c.buildSendRequest(r, compressedReqBody)
}

// buildSendRequest Build request header and send frame
func (c *clientCodec) buildSendRequest(r *rpc.Request, reqBody []byte) error {
	h := header.RequestPool.Get().(*header.RequestHeader)
	defer func() {
		h.ResetHeader()
		header.RequestPool.Put(h)
	}()

	h.ID = r.Seq
	h.Method = r.ServiceMethod
	h.RequestLen = uint32(len(reqBody))
	h.CompressType = c.compressor
	h.Checksum = crc32.ChecksumIEEE(reqBody)

	if err := sendFrame(c.w, h.Marshal()); err != nil {
		return err
	}
	if err := write(c.w, reqBody); err != nil {
		return err
	}
	c.w.(*bufio.Writer).Flush()
	return nil
}

func (c *clientCodec) ReadResponseHeader(response *rpc.Response) error {
	//TODO implement me
	panic("implement me")
}

func (c *clientCodec) ReadResponseBody(a any) error {
	//TODO implement me
	panic("implement me")
}

func (c *clientCodec) Close() error {
	//TODO implement me
	panic("implement me")
}

// NewClientCodec Create a new client codec
func NewClientCodec(conn io.ReadWriteCloser,
	compressType compressor.CompressType, serializer serializer.Serializer) rpc.ClientCodec {

	return &clientCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		compressor: compressType,
		serializer: serializer,
		pending:    make(map[uint64]string),
	}
}
