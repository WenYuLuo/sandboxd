// Copyright (c) 2026 Ant Group Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package imageregistry

import (
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetOrCreateTransportSetsConnectionLimits(t *testing.T) {
	client, err := NewClient("")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	transport := client.getOrCreateTransport("")
	if transport.base.MaxConnsPerHost != maxRegistryConnsPerHost {
		t.Fatalf("MaxConnsPerHost = %d, want %d", transport.base.MaxConnsPerHost, maxRegistryConnsPerHost)
	}
	if transport.base.MaxIdleConnsPerHost != maxRegistryIdleConnsPerHost {
		t.Fatalf("MaxIdleConnsPerHost = %d, want %d", transport.base.MaxIdleConnsPerHost, maxRegistryIdleConnsPerHost)
	}
	if transport.base.MaxIdleConns != maxRegistryIdleConns {
		t.Fatalf("MaxIdleConns = %d, want %d", transport.base.MaxIdleConns, maxRegistryIdleConns)
	}
}

func TestRegistryTransportClosesConnectionAfterServerError(t *testing.T) {
	conn := &testConn{}
	transport := &registryTransport{
		base: newRegistryBaseTransport(""),
		roundTripper: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			trace := httptrace.ContextClientTrace(req.Context())
			if trace == nil || trace.GotConn == nil {
				t.Fatalf("missing got-conn trace on request")
			}
			trace.GotConn(httptrace.GotConnInfo{Conn: conn})

			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Body:       io.NopCloser(strings.NewReader("boom")),
				Request:    req,
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "https://registry.example/v2/test/blobs/sha256:deadbeef", nil)
	if err != nil {
		t.Fatalf("NewRequest() error: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error: %v", err)
	}
	if _, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	if got := conn.closeCount.Load(); got != 1 {
		t.Fatalf("close count = %d, want 1", got)
	}
}

func TestRegistryTransportKeepsConnectionForHTTP2ServerError(t *testing.T) {
	conn := &testConn{}
	transport := &registryTransport{
		base: newRegistryBaseTransport(""),
		roundTripper: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			trace := httptrace.ContextClientTrace(req.Context())
			if trace == nil || trace.GotConn == nil {
				t.Fatalf("missing got-conn trace on request")
			}
			trace.GotConn(httptrace.GotConnInfo{Conn: conn})

			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Proto:      "HTTP/2.0",
				ProtoMajor: 2,
				ProtoMinor: 0,
				Body:       io.NopCloser(strings.NewReader("boom")),
				Request:    req,
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "https://registry.example/v2/test/blobs/sha256:deadbeef", nil)
	if err != nil {
		t.Fatalf("NewRequest() error: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error: %v", err)
	}
	if _, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	if got := conn.closeCount.Load(); got != 0 {
		t.Fatalf("close count = %d, want 0", got)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type testConn struct {
	closeCount atomic.Int32
}

func (c *testConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (c *testConn) Write(p []byte) (int, error)      { return len(p), nil }
func (c *testConn) Close() error                     { c.closeCount.Add(1); return nil }
func (c *testConn) LocalAddr() net.Addr              { return testAddr("local") }
func (c *testConn) RemoteAddr() net.Addr             { return testAddr("remote") }
func (c *testConn) SetDeadline(time.Time) error      { return nil }
func (c *testConn) SetReadDeadline(time.Time) error  { return nil }
func (c *testConn) SetWriteDeadline(time.Time) error { return nil }

type testAddr string

func (a testAddr) Network() string { return string(a) }
func (a testAddr) String() string  { return string(a) }
