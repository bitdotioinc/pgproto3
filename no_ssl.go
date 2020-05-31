package pgproto3

import (
	"encoding/json"
)

type NoSSL struct{}

// Backend identifies this message as sendable by the PostgreSQL backend.
func (*NoSSL) Backend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 4 byte message length.
func (dst *NoSSL) Decode(src []byte) error {
	if len(src) != 0 {
		return &invalidMessageLenErr{messageType: "NoSSL", expectedLen: 0, actualLen: len(src)}
	}

	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *NoSSL) Encode(dst []byte) []byte {
	return append(dst, 'N', 0, 0, 0, 4)
}

// MarshalJSON implements encoding/json.Marshaler.
func (src NoSSL) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string
	}{
		Type: "NoSSL",
	})
}
