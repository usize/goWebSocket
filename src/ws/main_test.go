package main

import (
	"bytes"
	"fmt"
	"testing"
)

const (
	// Values taken from RFC 6455 section 1.2
	secWebSocketKey    = "dGhlIHNhbXBsZSBub25jZQ=="
	secWebSocketAccept = "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
)

func generateClientHandshake(secWebSocketKey string) []byte {
	return []byte(fmt.Sprintf("GET /chat HTTP/1.1\r\n"+
		"Host: server.example.com\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Key: %s\r\n"+
		"Origin: http://example.com\r\n"+
		"Sec-WebSocket-Protocol: chat, superchat\r\n"+
		"Sec-WebSocket-Version: 13\r\n", secWebSocketKey))

}

func generateServerHandshake(secWebSocketAccept string) []byte {
	return []byte(fmt.Sprintf(handshakeFormat, secWebSocketAccept))
}

func TestGenerateHandshake(t *testing.T) {
	cxt := WebSocketContext{nil, generateClientHandshake(secWebSocketKey), 0, false}
	handshake, err := cxt.GenerateHandshake()
	if err != nil {
		t.Error("%s", err)
	}
	expectedHandshake := generateServerHandshake(secWebSocketAccept)
	if !bytes.Equal(handshake, expectedHandshake) {
		t.Error("Handshake Test Failure: %s != %s", string(handshake), string(expectedHandshake))
	}

}
