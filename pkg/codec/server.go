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

type requestContext struct {
	requestID   uint64
	compareType compressor.CompressType
}

type serverCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	request    header.RequestHeader
	serializer serializer.Serializer
	mutex      sync.Mutex
	seq        uint64
	pending    map[uint64]*requestContext
}

// ReadRequestHeader read the rpc request header from the io stream
func (s *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	// reset serverCodec request header
	s.request.ResetHeader()
	// read request header
	data, err := recvFrame(s.r)
	if err != nil {
		return err
	}
	// unmarshal data
	err = s.request.Unmarshal(data)
	if err != nil {
		return err
	}
	// fill field
	s.mutex.Lock()
	s.seq++
	s.pending[s.seq] = &requestContext{
		requestID:   s.request.ID,
		compareType: s.request.CompressType,
	}
	r.ServiceMethod = s.request.Method
	r.Seq = s.seq
	s.mutex.Unlock()
	return nil
}

// ReadRequestBody read the rpc request body from the io stream
func (s *serverCodec) ReadRequestBody(param any) error {
	if param == nil {
		if s.request.RequestLen != 0 { // discard excess
			if err := read(s.r, make([]byte, s.request.RequestLen)); err != nil {
				return err
			}
		}
		return nil
	}

	// read ResponseLen length bytes
	reqBody := make([]byte, s.request.RequestLen)
	err := read(s.r, reqBody)
	if err != nil {
		return err
	}

	// check
	if s.request.Checksum != 0 {
		if crc32.ChecksumIEEE(reqBody) != s.request.Checksum {
			return UnexpectedChecksumError
		}
	}
	// check compressor whether exist
	if _, ok := compressor.
		Compressors[s.request.GetCompressType()]; !ok {
		return NotFoundCompressorError
	}
	// unzip request body
	req, err := compressor.Compressors[s.request.GetCompressType()].Unzip(reqBody)
	if err != nil {
		return err
	}
	return s.serializer.Unmarshal(req, param)
}

// WriteResponse Write the rpc response header and body to the io stream
func (s *serverCodec) WriteResponse(r *rpc.Response, param any) error {
	s.mutex.Lock()
	reqCtx, ok := s.pending[r.Seq]
	if !ok {
		s.mutex.Unlock()
		return InvalidSequenceError
	}
	delete(s.pending, r.Seq)
	s.mutex.Unlock()

	// call rpc get wrong
	if r.Error != "" {
		param = nil
	}
	// check compressor whether exist
	if _, ok := compressor.Compressors[reqCtx.compareType]; !ok {
		return NotFoundCompressorError
	}

	// marshal
	var respBody []byte
	var err error
	if param != nil {
		respBody, err = s.serializer.Marshal(param)
		if err != nil {
			return err
		}
	}
	// zip response body
	compressedRespBody, err := compressor.Compressors[reqCtx.compareType].Zip(respBody)
	if err != nil {
		return err
	}
	return s.buildSendResponse(reqCtx, r, compressedRespBody)
}

// buildSendResponse build response and send frame
// reqBody is compressed
func (s *serverCodec) buildSendResponse(reqCtx *requestContext, r *rpc.Response, reqBody []byte) error {
	// get header from pool
	h := header.ResponsePool.Get().(*header.ResponseHeader)
	defer func() {
		h.ResetHeader()
		header.ResponsePool.Put(h)
	}()
	// fill header
	h.ID = reqCtx.requestID
	h.Error = r.Error
	h.ResponseLen = uint32(len(reqBody))
	h.Checksum = crc32.ChecksumIEEE(reqBody)
	h.CompressType = reqCtx.compareType

	var err error
	// send header
	if err = sendFrame(s.w, h.Marshal()); err != nil {
		return err
	}
	// send body
	if err = write(s.w, reqBody); err != nil {
		return err
	}
	s.w.(*bufio.Writer).Flush()
	return nil
}

func (s *serverCodec) Close() error {
	return s.c.Close()
}

// NewServerCodec Create a new server codec
func NewServerCodec(conn io.ReadWriteCloser, serializer serializer.Serializer) rpc.ServerCodec {
	return &serverCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		serializer: serializer,
		pending:    make(map[uint64]*requestContext),
	}
}
