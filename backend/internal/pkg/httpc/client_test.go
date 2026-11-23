package httpc

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClientBlocksPrivateDialTargets(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	port := server.Listener.Addr().(*net.TCPAddr).Port
	targets := []struct {
		name string
		url  string
	}{
		{name: "localhost", url: fmt.Sprintf("http://localhost:%d/feed.xml", port)},
		{name: "loopback ip", url: fmt.Sprintf("http://127.0.0.1:%d/feed.xml", port)},
	}

	client, err := NewClient(time.Second, "", false)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	for _, tt := range targets {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatalf("new request: %v", err)
			}

			if _, err := client.Do(req); err == nil {
				t.Fatal("expected private dial target to be blocked")
			}
		})
	}

	if got := hits.Load(); got != 0 {
		t.Fatalf("expected no requests to reach the server, got %d", got)
	}
}

// TestClientBlocksDNSRebinding guards the SSRF check against a time-of-check /
// time-of-use gap: a hostname may resolve to a public IP when the URL is
// validated, then to loopback when the connection is dialed. The check must
// therefore apply to the resolved dial target, not only to the hostname before
// dialing.
func TestClientBlocksDNSRebinding(t *testing.T) {
	dnsAddr, aQueries := startRebindingDNS(t, 2)

	// Swapping net.DefaultResolver is process-wide, which is safe here only
	// because this package has no t.Parallel tests. Both validatePublicHost and
	// net.Dialer resolve through net.DefaultResolver, so no production seam is
	// needed.
	originalResolver := net.DefaultResolver
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, dnsAddr)
		},
	}
	t.Cleanup(func() { net.DefaultResolver = originalResolver })

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits.Add(1)
	}))
	defer server.Close()

	// The client key is distinct from other tests because clients are pooled.
	client, err := NewClient(300*time.Millisecond, "", false)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	// A trailing dot keeps the name absolute so no search domain is appended.
	target := fmt.Sprintf("http://rebind.test.:%d/feed.xml", server.Listener.Addr().(*net.TCPAddr).Port)
	if err := ValidateRequestURL(context.Background(), target, false); err != nil {
		t.Fatalf("expected the first resolution to look public, got %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	if got := hits.Load(); got != 0 {
		t.Fatalf("rebound request reached the server %d times", got)
	}
	if err == nil {
		t.Fatal("expected the rebound request to fail")
	}
	if got := aQueries.Load(); got < 2 {
		t.Fatalf("expected validation and dial resolutions to use the fake DNS server, got %d A queries", got)
	}
}

// startRebindingDNS serves A queries from loopback: the first publicAnswers
// answers are a public TEST-NET-1 address, later ones are loopback, modelling
// DNS rebinding. AAAA and other queries get an empty NOERROR answer so the
// resolver neither falls back nor retries. It returns the server address and
// an A query counter.
func startRebindingDNS(t *testing.T, publicAnswers int) (string, *atomic.Int32) {
	t.Helper()

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	var aQueries atomic.Int32
	go func() {
		buf := make([]byte, 512)
		for {
			n, peer, err := conn.ReadFrom(buf)
			if err != nil {
				return // listener closed by cleanup
			}
			query := buf[:n]

			i := 12
			for i < len(query) && query[i] != 0 {
				i += int(query[i]) + 1
			}
			if i+5 > len(query) {
				continue
			}

			reply := make([]byte, i+5, i+21) // header, question, one A record
			copy(reply, query[:i+5])
			reply[2] |= 0x80                               // QR: this is a response
			reply[3] = 0x80                                // RA, RCODE=NOERROR
			if binary.BigEndian.Uint16(query[i+1:]) == 1 { // QTYPE A
				answer := net.IPv4(127, 0, 0, 1).To4()
				if aQueries.Add(1) <= int32(publicAnswers) {
					answer = net.IPv4(192, 0, 2, 10).To4() // TEST-NET-1
				}
				binary.BigEndian.PutUint16(reply[6:], 1)                        // ANCOUNT
				reply = append(reply, 0xC0, 0x0C, 0, 1, 0, 1, 0, 0, 0, 0, 0, 4) // name ptr, TYPE A, CLASS IN, TTL 0, RDLENGTH 4
				reply = append(reply, answer...)
			}
			_, _ = conn.WriteTo(reply, peer)
		}
	}()

	return conn.LocalAddr().String(), &aQueries
}
