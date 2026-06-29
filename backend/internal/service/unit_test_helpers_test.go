//go:build unit

package service

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func newEventStreamResponse(status int, body string) *http.Response {
	stream := encodeEventStreamMessage("assistantResponseEvent", []byte(body))
	return &http.Response{
		StatusCode: status,
		Header: http.Header{
			"Content-Type":      []string{"application/vnd.amazon.eventstream"},
			"Cache-Control":     []string{"no-cache"},
			"X-Accel-Buffering": []string{"no"},
		},
		Body: io.NopCloser(bytes.NewReader(stream)),
	}
}

func rewriteHostTransport(t *testing.T, server *httptest.Server) http.RoundTripper {
	t.Helper()

	targetURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}

	base := server.Client().Transport
	if base == nil {
		base = http.DefaultTransport
	}

	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		clone := req.Clone(req.Context())
		clone.URL = new(url.URL)
		*clone.URL = *req.URL
		clone.URL.Scheme = targetURL.Scheme
		clone.URL.Host = targetURL.Host
		return base.RoundTrip(clone)
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func encodeEventStreamMessage(eventType string, payload []byte) []byte {
	headers := encodeEventTypeHeader(eventType)
	totalLength := 12 + len(headers) + len(payload) + 4
	frame := make([]byte, totalLength)
	binary.BigEndian.PutUint32(frame[0:4], uint32(totalLength))
	binary.BigEndian.PutUint32(frame[4:8], uint32(len(headers)))
	copy(frame[12:], headers)
	copy(frame[12+len(headers):], payload)
	return frame
}

func encodeEventTypeHeader(eventType string) []byte {
	name := []byte(":event-type")
	value := []byte(eventType)
	header := make([]byte, 1+len(name)+1+2+len(value))
	offset := 0
	header[offset] = byte(len(name))
	offset++
	copy(header[offset:], name)
	offset += len(name)
	header[offset] = 7
	offset++
	binary.BigEndian.PutUint16(header[offset:offset+2], uint16(len(value)))
	offset += 2
	copy(header[offset:], value)
	return header
}
