/**
* Incomplete WebSocket protocol implementation in Golang: for demonstration purposes.
* winter2718@gmail.com @mrrrgn github.com/mrrrgn
**/

package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
)

const (
	handshakeGUID   = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	handshakeFormat = "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: %s\n\r\n\r"
	bufferSize = 2048
)

type WebSocketContext struct {
	conn        net.Conn
	buffer      []byte // our raw data in the buffer!
	offset      int    // points to the end of our buffer data
	isConnected bool   // set to true after a successful handshake
}

func (cxt *WebSocketContext) clearBuffer(bufSize int) {
	// The garbage collector will throw out our old buffer sometime
	// after this finishes.
	cxt.buffer = make([]byte, bufSize)
	cxt.offset = 0
}

func (cxt *WebSocketContext) growBuffer(size int) error {
	if size <= 0 {
		return errors.New("Must grow buffer by more than 0 bytes.")
	}
	newBuffer := make([]byte, len(cxt.buffer)+size)
	copy(newBuffer, cxt.buffer)
	cxt.buffer = newBuffer
	return nil
}

// Build a response to the WebSocket clients initial security challenge
func (cxt *WebSocketContext) generateHandShake() ([]byte, error) {
	// Split out the security key
	rawKeyExpr := regexp.MustCompile("Sec-WebSocket-Key: (.*)==").FindSubmatch(cxt.buffer)
	if len(rawKeyExpr) <= 1 {
		return nil, errors.New("Can't parse client handshake, malformed Sec-WebSocket-Key")
	}
	rawKey := rawKeyExpr[1]
	securityKey := make([]byte, len(rawKey)+len(handshakeGUID)+len("=="))
	// Remember, slices are weird. It's still pointing to our buffer,
	// so let's make a hard copy to make sure we don't lose it!
	copy(securityKey, rawKey)
	copy(securityKey[len(rawKey):], "=="+handshakeGUID)
	sh := sha1.New()
	sh.Write(securityKey)
	sha1Key := sh.Sum(nil)
	base64Key := base64.StdEncoding.EncodeToString(sha1Key)
	response := fmt.Sprintf(handshakeFormat, base64Key)
	return []byte(response), nil
}

func (cxt *WebSocketContext) SendHandShake() error {
	handshake, err := cxt.generateHandShake()
	if err != nil {
		return err
	}
	size, err := cxt.conn.Write(handshake)
	if size < len(handshake) {
		return errors.New("Checksum Failed")
	}
	return err
}

func handleConnection(conn net.Conn) {
	cxt := WebSocketContext{conn, make([]byte, bufferSize), 0, false}
	for {
		if cxt.offset >= len(cxt.buffer) {
			// Double our buffer size if we run out of space
			cxt.growBuffer(len(cxt.buffer))
		}
		size, err := bufio.NewReader(conn).Read(cxt.buffer[cxt.offset:])
		if err != nil {
			break
		}
		cxt.offset += size
		if cxt.isConnected == false && bytes.Contains(cxt.buffer, []byte("Upgrade: websocket")) {
			err := cxt.SendHandShake()
			if err != nil {
				fmt.Errorf("Handshake Error: %s", err)
				break
			}
			fmt.Printf("Successful Connection with: %s", conn.RemoteAddr().String())
			cxt.isConnected = true
			cxt.clearBuffer(len(cxt.buffer))
		}
		fmt.Println(string(cxt.buffer))
	}
	conn.Close()
}

func main() {
	address := os.Args[1]
	port := os.Args[2]
	server, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		fmt.Errorf("%s", err)
		os.Exit(1)
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Errorf("%s", err)
		}
		go handleConnection(conn)
	}
}
