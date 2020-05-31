package pgproto3

import (
	"encoding/json"
)

type NoSSL struct{}

// Backend identifies this message as sendable by the PostgreSQL backend.
func (*NoSSL) Backend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
func (dst *NoSSL) Decode(src []byte) error {
	if len(src) != 0 {
		return &invalidMessageLenErr{messageType: "NoSSL", expectedLen: 0, actualLen: len(src)}
	}

	return nil
}

// Encode encodes src into dst. dst should just be one byte
func (src *NoSSL) Encode(dst []byte) []byte {
	return append(dst, 'N')
}

// MarshalJSON implements encoding/json.Marshaler.
func (src NoSSL) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string
	}{
		Type: "NoSSL",
	})
}
