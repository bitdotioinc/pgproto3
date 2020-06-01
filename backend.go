package pgproto3

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Backend acts as a server for the PostgreSQL wire protocol version 3.
type Backend struct {
	r io.Reader
	w  io.Writer

	// Frontend message flyweights
	bind            Bind
	cancelRequest   CancelRequest
	_close          Close
	copyFail        CopyFail
	copyData	CopyData
	copyDone	CopyDone
	describe        Describe
	execute         Execute
	flush           Flush
	gssEncRequest   GSSEncRequest
	parse           Parse
	passwordMessage PasswordMessage
	query           Query
	sslRequest      SSLRequest
	startupMessage  StartupMessage
	sync            Sync
	terminate       Terminate

	bodyLen    int
	msgType    byte
	partialMsg bool
}

// NewBackend creates a new Backend.
func NewBackend(r io.Reader, w io.Writer) *Backend {
	return &Backend{r: r, w: w}
}

// Send sends a message to the frontend.
func (b *Backend) Send(msg BackendMessage) error {
	_, err := b.w.Write(msg.Encode(nil))
	return err
}

// ReceiveStartupMessage receives the initial connection message. This method is used instead of the normal Receive method
// because the initial connection message is "special" and does not include the message type as the first byte. This
// will return either a StartupMessage, SSLRequest, GSSEncRequest, or CancelRequest.
func (b *Backend) ReceiveStartupMessage() (FrontendMessage, error) {
	buf := make([]byte, 4)
	n, err := io.ReadFull(b.r, buf)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, fmt.Errorf("Did not read the full startup message header, read=%d, error=%v", n, err)
	}

	msgSize := int(binary.BigEndian.Uint32(buf) - 4)

	msgBody := make([]byte, msgSize)
	n, err = io.ReadFull(b.r, msgBody)
	if err != nil {
		return nil, err
	}
	if n != msgSize {
		return nil, fmt.Errorf("Did not read the full startup message, read=%d, error=%v", n, err)
	}

	code := binary.BigEndian.Uint32(msgBody)

	switch code {
	case ProtocolVersionNumber:
		err = b.startupMessage.Decode(msgBody)
		if err != nil {
			return nil, err
		}
		return &b.startupMessage, nil
	case sslRequestNumber:
		err = b.sslRequest.Decode(msgBody)
		if err != nil {
			return nil, err
		}
		return &b.sslRequest, nil
	case cancelRequestCode:
		err = b.cancelRequest.Decode(msgBody)
		if err != nil {
			return nil, err
		}
		return &b.cancelRequest, nil
	case gssEncReqNumber:
		err = b.gssEncRequest.Decode(buf)
		if err != nil {
			return nil, err
		}
		return &b.gssEncRequest, nil
	default:
		return nil, fmt.Errorf("unknown startup message code: %d", code)
	}
}

// Receive receives a message from the frontend.
func (b *Backend) Receive() (FrontendMessage, error) {
	if !b.partialMsg {
		header := make([]byte, 5)
		n, err := io.ReadFull(b.r, header)
		if err != nil {
			return nil, err
		}
		if n!=5 {
			return nil, fmt.Errorf("Did not read the full message header, read=%d, error=%v", n, err)
		}

		b.msgType = header[0]
		b.bodyLen = int(binary.BigEndian.Uint32(header[1:])) - 4
		b.partialMsg = true
	}

	msgBody := make([]byte, b.bodyLen)
	n, err := io.ReadFull(b.r, msgBody)
	if err != nil {
		return nil, err
	}		
	if n != b.bodyLen {
		return nil, fmt.Errorf("Message declared len is longer than the actual message")
	}

	b.partialMsg = false

	var msg FrontendMessage
	switch b.msgType {
	case 'B':
		msg = &b.bind
	case 'C':
		msg = &b._close
	case 'c':
		msg = &b.copyDone
	case 'D':
		msg = &b.describe
	case 'd':
		msg = &b.copyData
	case 'E':
		msg = &b.execute
	case 'f':
		msg = &b.copyFail
	case 'H':
		msg = &b.flush
	case 'P':
		msg = &b.parse
	case 'p':
		msg = &b.passwordMessage
	case 'Q':
		msg = &b.query
	case 'S':
		msg = &b.sync
	case 'X':
		msg = &b.terminate
	default:
		return nil, fmt.Errorf("unknown message type: %c", b.msgType)
	}

	err = msg.Decode(msgBody)
	return msg, err
}
