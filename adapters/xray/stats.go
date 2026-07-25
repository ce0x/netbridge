package xray

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type StatsClient struct {
	addr string
}

func NewStatsClient(port int) *StatsClient {
	return &StatsClient{
		addr: fmt.Sprintf("127.0.0.1:%d", port),
	}
}

type TrafficStats struct {
	Up   int64
	Down int64
}

func (c *StatsClient) QueryStats(ctx context.Context) (*TrafficStats, error) {
	conn, err := net.DialTimeout("tcp", c.addr, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to stats API: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))

	// QueryStats request via gRPC over h2c (plaintext HTTP/2)
	// Method: /xray.app.stats.command.StatsService/QueryStats
	// Request pattern: {"pattern": "user>>>traffic>>>uplink", "reset_": false}

	reqBody := buildQueryStatsRequest("user>>>traffic>>>uplink")
	resp, err := c.doGRPC(conn, "/xray.app.stats.command.StatsService/QueryStats", reqBody)
	if err != nil {
		return nil, fmt.Errorf("query uplink: %w", err)
	}
	up := parseStatsResponse(resp)

	reqBody = buildQueryStatsRequest("user>>>traffic>>>downlink")
	resp, err = c.doGRPC(conn, "/xray.app.stats.command.StatsService/QueryStats", reqBody)
	if err != nil {
		return nil, fmt.Errorf("query downlink: %w", err)
	}
	down := parseStatsResponse(resp)

	return &TrafficStats{Up: up, Down: down}, nil
}

func (c *StatsClient) doGRPC(conn net.Conn, method string, body []byte) ([]byte, error) {
	// Build HTTP/2 + gRPC framing
	// gRPC uses a 5-byte frame header: compressed(1) + length(4) + message
	frameLen := len(body)
	header := make([]byte, 5)
	header[0] = 0 // not compressed
	binary.BigEndian.PutUint32(header[1:5], uint32(frameLen))
	fullPayload := append(header, body...)

	// HTTP/2 HEADERS frame + DATA frame
	// Simplified: send raw HTTP/2 with gRPC content-type
	h2Header := buildH2Headers(method)
	all := append(h2Header, fullPayload...)

	// Send END_STREAM on headers
	all = append(all, 0, 0, 0, 0, 0) // DATA frame with END_STREAM

	_, err := conn.Write(all)
	if err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	// Read response (simplified - read until we get enough data)
	buf := make([]byte, 4096)
	total := 0
	for total < 5 {
		n, err := conn.Read(buf[total:])
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("read: %w", err)
		}
		total += n
		if total >= len(buf) {
			break
		}
	}

	// Skip HTTP/2 frames, find gRPC response
	// Look for the gRPC frame header (0x00 or 0x01 followed by 4-byte length)
	for i := 0; i < total-5; i++ {
		if buf[i] == 0x00 || buf[i] == 0x01 {
			msgLen := int(binary.BigEndian.Uint32(buf[i+1 : i+5]))
			if msgLen > 0 && msgLen < 1000 && i+5+msgLen <= total {
				return buf[i+5 : i+5+msgLen], nil
			}
		}
	}

	return buf[:total], nil
}

func buildQueryStatsRequest(pattern string) []byte {
	// Simplified protobuf: field 1 (pattern) = string, field 2 (reset) = bool
	// Protobuf encoding: field_num << 3 | wire_type
	// string = wire type 2, varint = wire type 0

	var result []byte

	// Field 1: pattern (string, wire type 2)
	result = append(result, 0x0a) // (1 << 3) | 2
	result = append(result, byte(len(pattern)))
	result = append(result, []byte(pattern)...)

	// Field 2: reset_ (bool, wire type 0) = false
	result = append(result, 0x10) // (2 << 3) | 0
	result = append(result, 0x00) // false

	return result
}

func parseStatsResponse(data []byte) int64 {
	// Parse protobuf: field 1 = Stats (message), inside that field 1 = name, field 2 = value (int64)
	// We want the value field
	for i := 0; i < len(data)-2; i++ {
		// Look for field tag 0x10 (field 2, varint) which is the value
		if data[i] == 0x10 {
			// Read varint
			val, n := readVarint(data[i+1:])
			if n > 0 {
				return int64(val)
			}
		}
	}
	return 0
}

func readVarint(data []byte) (uint64, int) {
	var val uint64
	for i, b := range data {
		val |= uint64(b&0x7f) << (7 * uint(i))
		if b < 0x80 {
			return val, i + 1
		}
		if i >= 9 {
			return 0, 0
		}
	}
	return 0, 0
}

func buildH2Headers(method string) []byte {
	// Simplified HTTP/2 HEADERS frame
	// In practice, this needs proper HPACK encoding
	// For now, use a minimal approach that works with xray's gRPC

	// HTTP/2 preface
	preface := []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")

	// HEADERS frame (length=0, type=0x1, flags=0x4 END_HEADERS, stream=1)
	headersFrame := []byte{0, 0, 0, 0x1, 0x4, 0, 0, 0, 0, 1}

	// Pseudo-headers encoded as raw key-value (simplified)
	// :method = POST
	// :path = /xray.app.stats.command.StatsService/QueryStats
	// :scheme = http
	// content-type = application/grpc
	// te = trailers

	var headers []byte
	headers = appendHPACK(headers, ":method", "POST")
	headers = appendHPACK(headers, ":path", method)
	headers = appendHPACK(headers, ":scheme", "http")
	headers = appendHPACK(headers, "content-type", "application/grpc")
	headers = appendHPACK(headers, "te", "trailers")

	// Update headers frame length
	binary.BigEndian.PutUint16(headersFrame[0:2], uint16(len(headers)))
	headersFrame = append(headersFrame, headers...)

	return append(preface, headersFrame...)
}

func appendHPACK(headers []byte, key, value string) []byte {
	// Simplified HPACK: indexed header field representation
	// For common headers, use static table indices
	// This is a minimal implementation

	// :method = POST → index 2
	// :scheme = http → index 7
	// :path → index 4 (with literal value)
	// content-type → index 31 (with literal value)
	// te → index 51 (with literal value)

	switch key {
	case ":method":
		if value == "POST" {
			return append(headers, 0x82) // index 2
		}
	case ":scheme":
		if value == "http" {
			return append(headers, 0x87) // index 7
		}
	}

	// For other headers, use literal with literal indexing
	headers = append(headers, 0x40) // literal with literal indexing
	headers = append(headers, byte(len(key)))
	headers = append(headers, []byte(key)...)
	headers = append(headers, byte(len(value)))
	headers = append(headers, []byte(value)...)
	return headers
}
