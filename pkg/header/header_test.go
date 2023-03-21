package header_test

import (
	"github.com/alilestera/tinyrpc/pkg/header"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// TestRequestHeader_Marshal .
func TestRequestHeader_Marshal(t *testing.T) {
	tests := []struct {
		name   string
		h      *header.RequestHeader
		expect []byte
	}{
		{
			"marshalRequestHeaderCase-1",
			&header.RequestHeader{
				CompressType: 0,
				Method:       "Add",
				ID:           12455,
				RequestLen:   266,
				Checksum:     3845236589,
			},
			[]byte{0x0, 0x0, 0x3, 0x41, 0x64, 0x64,
				0xa7, 0x61, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.h.Marshal()
			assert.Equal(t, tt.expect, data)
		})
	}
}

// TestRequestHeader_Unmarshal .
func TestRequestHeader_Unmarshal(t *testing.T) {
	type expect struct {
		header *header.RequestHeader
		err    error
	}
	tests := []struct {
		name   string
		data   []byte
		expect expect
	}{
		{
			"unmarshalRequestHeaderNormalData",
			[]byte{0x0, 0x0, 0x3, 0x41, 0x64, 0x64,
				0xa7, 0x61, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5},
			expect{&header.RequestHeader{
				CompressType: 0,
				Method:       "Add",
				ID:           12455,
				RequestLen:   266,
				Checksum:     3845236589,
			}, nil},
		},
		{
			"unmarshalRequestHeaderNilData",
			nil,
			expect{&header.RequestHeader{},
				header.ErrUnmarshal},
		},
		{
			"unmarshalRequestHeaderTooShortData",
			[]byte{0x0},
			expect{&header.RequestHeader{},
				header.ErrUnmarshal},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &header.RequestHeader{}
			err := h.Unmarshal(tt.data)
			assert.Equal(t, true, reflect.DeepEqual(tt.expect.header, h))
			assert.Equal(t, tt.expect.err, err)
		})
	}
}

// TestRequestHeader_ResetHeader .
func TestRequestHeader_ResetHeader(t *testing.T) {
	tests := []struct {
		name string
		h    *header.RequestHeader
	}{
		{
			"resetRequestHeaderCase-1",
			&header.RequestHeader{
				CompressType: 0,
				Method:       "Add",
				ID:           12455,
				RequestLen:   266,
				Checksum:     3845236589,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.ResetHeader()
			assert.Equal(t, true, reflect.DeepEqual(tt.h, &header.RequestHeader{}))
		})
	}
}

// TestResponseHeader_Marshal .
func TestResponseHeader_Marshal(t *testing.T) {
	tests := []struct {
		name   string
		h      *header.ResponseHeader
		expect []byte
	}{
		{
			"marshalResponseHeaderCase-1",
			&header.ResponseHeader{
				CompressType: 0,
				Error:        "error",
				ID:           12455,
				ResponseLen:  266,
				Checksum:     3845236589,
			},
			[]byte{0x0, 0x0, 0xa7, 0x61, 0x5, 0x65, 0x72,
				0x72, 0x6f, 0x72, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.h.Marshal()
			assert.Equal(t, tt.expect, data)
		})
	}
}

// TestResponseHeader_Unmarshal .
func TestResponseHeader_Unmarshal(t *testing.T) {
	type expect struct {
		header *header.ResponseHeader
		err    error
	}
	tests := []struct {
		name   string
		data   []byte
		expect expect
	}{
		{
			"unmarshalResponseHeaderNormalData",
			[]byte{0x0, 0x0, 0xa7, 0x61, 0x5, 0x65, 0x72,
				0x72, 0x6f, 0x72, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5},
			expect{&header.ResponseHeader{
				CompressType: 0,
				Error:        "error",
				ID:           12455,
				ResponseLen:  266,
				Checksum:     3845236589,
			}, nil},
		},
		{
			"unmarshalResponseHeaderNilData",
			nil,
			expect{&header.ResponseHeader{},
				header.ErrUnmarshal},
		},
		{
			"unmarshalResponseHeaderTooShortData",
			[]byte{0x0},
			expect{&header.ResponseHeader{},
				header.ErrUnmarshal},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &header.ResponseHeader{}
			err := h.Unmarshal(tt.data)
			assert.Equal(t, true, reflect.DeepEqual(tt.expect.header, h))
			assert.Equal(t, tt.expect.err, err)
		})
	}
}

// TestResponseHeader_ResetHeader .
func TestResponseHeader_ResetHeader(t *testing.T) {
	tests := []struct {
		name string
		h    *header.ResponseHeader
	}{
		{
			"resetResponseHeader",
			&header.ResponseHeader{
				CompressType: 0,
				Error:        "error",
				ID:           12455,
				ResponseLen:  266,
				Checksum:     3845236589,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.ResetHeader()
			assert.Equal(t, true, reflect.DeepEqual(tt.h, &header.ResponseHeader{}))
		})
	}
}
