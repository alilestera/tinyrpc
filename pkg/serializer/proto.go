package serializer

import (
	"errors"
	"google.golang.org/protobuf/proto"
)

// ErrNotImplementProtoMessage refers to param not implemented by proto.Message
var ErrNotImplementProtoMessage = errors.New("param does not implement proto.Message")

var Proto = ProtoSerializer{}

// ProtoSerializer implements the Serializer interface
type ProtoSerializer struct {
}

// Marshal .
func (p ProtoSerializer) Marshal(message interface{}) ([]byte, error) {
	if message == nil {
		return []byte{}, nil
	}

	var body proto.Message
	var ok bool
	if body, ok = message.(proto.Message); !ok {
		return nil, ErrNotImplementProtoMessage
	}
	return proto.Marshal(body)
}

// Unmarshal .
func (p ProtoSerializer) Unmarshal(data []byte, message interface{}) error {
	if message == nil {
		return nil
	}

	var body proto.Message
	var ok bool
	if body, ok = message.(proto.Message); !ok {
		return ErrNotImplementProtoMessage
	}

	return proto.Unmarshal(data, body)
}
