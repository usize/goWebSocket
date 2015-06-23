package frame

/**
* Encode/Decode WebSocket frames
**/

import "errors"

type WebSocketFrame struct {
	// WebSocket Header Fields
	fin           bool
	opcode        int
	mask          bool
	payloadLength uint64
	maskingKey    []byte
	// State Tracking
	headerReady  bool
	FrameReady   bool
	headerLength int
}

func (wsf *WebSocketFrame) parseHeader(buffer []byte) error {
	if len(buffer) < 2 {
		errors.New("Buffer is less than the minimum header length")
	}
	wsf.headerLength = 2 // Minimum header length
	wsf.fin = (buffer[0] >> 7) > 0
	wsf.opcode = int(buffer[0] & 15 /* 0b1111 */)
	wsf.mask = (buffer[1] >> 7) > 0
	wsf.payloadLength = uint64(buffer[1] & 127 /* 0b1111111 */)
	if wsf.payloadLength == 126 {
		// 16 bit unsigned integer
		wsf.payloadLength = uint64(buffer[2] << 8)
		wsf.payloadLength += uint64(buffer[3])
		wsf.headerLength += 2
	} else if wsf.payloadLength == 127 {
		// It's a 64 bit unsigned integer, TODO: we'll need to test
		// using unit tests, since I don't feel like trying to write
		// 2^16 characters right now
		for i := 0; i <= 8; i++ {
			wsf.payloadLength += uint64(buffer[2+i] << uint(8*(7-i)))
		}
		wsf.headerLength += 8
	}
	if wsf.mask {
		wsf.maskingKey = buffer[wsf.headerLength : wsf.headerLength+4]
		wsf.headerLength += 4 // The masking-key adds 4 bytes
	}

	wsf.headerReady = true
	return nil
}

func (wsf *WebSocketFrame) Decode(buffer []byte) ([]byte, error) {
	if !wsf.headerReady {
		return nil, errors.New("No header has been seen, impossible to parse message")
	} else if !wsf.mask {
		return nil, errors.New("Masking is not enabled on the current frame, no need to decode")
	}
	message := make([]byte, uint64(wsf.headerLength)+wsf.payloadLength)
	for i, c := range buffer[wsf.headerLength : uint64(wsf.headerLength)+wsf.payloadLength] {
		message = append(message, c^wsf.maskingKey[i%4])
	}
	return message, nil
}

func (wsf *WebSocketFrame) Read(buffer []byte) error {
	if !wsf.headerReady {
		wsf.parseHeader(buffer)
	}
	if uint64(len(buffer)) >= uint64(wsf.headerLength)+wsf.payloadLength {
		wsf.FrameReady = true
	}
	return nil
}
