// =============================================================================
// LESSON 13: ADVANCED NETWORKING — TCP, HTTP Internals, Custom Protocols
// =============================================================================
//
// Go's net and net/http packages are used by 90% of cloud infrastructure.
// This lesson covers what's under the hood and how to build custom protocols.
//
// TOPICS:
//   - Raw TCP server/client with custom protocol
//   - HTTP server internals (connection hijacking, SSE, websocket-style)
//   - Connection pooling and keep-alive
//   - Timeouts at every layer
//   - gRPC-style framing
// =============================================================================

package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// PART 1: Custom TCP Protocol — Length-prefixed binary framing
// =============================================================================
//
// Most production protocols (gRPC, Kafka, Redis) use length-prefixed framing:
//   [4 bytes: message length][N bytes: message payload]
//
// This avoids delimiter-based parsing issues (what if data contains the delimiter?)

// Frame format: [4-byte big-endian length][payload]

func writeFrame(conn net.Conn, data []byte) error {
	// Write length prefix
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(data)))

	// Use writev-style: write header + data
	if _, err := conn.Write(header); err != nil {
		return err
	}
	_, err := conn.Write(data)
	return err
}

func readFrame(conn net.Conn) ([]byte, error) {
	// Read length prefix
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header)

	// Sanity check to prevent OOM
	const maxFrameSize = 16 * 1024 * 1024 // 16MB max
	if length > maxFrameSize {
		return nil, fmt.Errorf("frame too large: %d bytes", length)
	}

	// Read payload
	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

// TCP server using our protocol
func startTCPServer(ctx context.Context, addr string) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	fmt.Printf("TCP server listening on %s\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				fmt.Printf("Accept error: %v\n", err)
				continue
			}
		}

		go handleTCPConnection(conn)
	}
}

func handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	// Set deadlines to prevent hanging connections
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	for {
		data, err := readFrame(conn)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			return
		}

		// Echo back with "ECHO: " prefix
		response := append([]byte("ECHO: "), data...)
		if err := writeFrame(conn, response); err != nil {
			fmt.Printf("Write error: %v\n", err)
			return
		}

		// Reset read deadline
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	}
}

// TCP client
func tcpClient(addr string, messages []string) error {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, msg := range messages {
		// Send
		if err := writeFrame(conn, []byte(msg)); err != nil {
			return fmt.Errorf("send: %w", err)
		}

		// Receive
		response, err := readFrame(conn)
		if err != nil {
			return fmt.Errorf("recv: %w", err)
		}

		fmt.Printf("  Sent: %q → Got: %q\n", msg, string(response))
	}
	return nil
}

// =============================================================================
// PART 2: HTTP Server Internals — Timeout Architecture
// =============================================================================
//
// Go's HTTP server has 5 timeout points:
//
//               ┌──────────────────────────────────────────────┐
//   Accept ──→  │  ReadTimeout  │  Handler  │  WriteTimeout    │ ──→ Close
//               └──────────────────────────────────────────────┘
//                    │                              │
//            ReadHeaderTimeout               IdleTimeout (keep-alive)
//
// CRITICAL: Always set timeouts. Without them, a slow client can hold
// connections forever, exhausting your server's file descriptors.

func createProductionHTTPServer() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	// Server-Sent Events (SSE) — streaming response
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		ctx := r.Context() // cancelled when client disconnects

		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return // client disconnected
			case <-time.After(500 * time.Millisecond):
				fmt.Fprintf(w, "data: Event %d at %s\n\n", i, time.Now().Format(time.RFC3339))
				flusher.Flush()
			}
		}
	})

	return &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       5 * time.Second,   // max time to read request
		ReadHeaderTimeout: 2 * time.Second,   // max time to read headers
		WriteTimeout:      10 * time.Second,  // max time to write response
		IdleTimeout:       120 * time.Second, // keep-alive connection timeout
		MaxHeaderBytes:    1 << 20,           // 1MB max header size
	}
}

// =============================================================================
// PART 3: HTTP Client — Production Configuration
// =============================================================================

func createProductionHTTPClient() *http.Client {
	transport := &http.Transport{
		// Connection pool settings
		MaxIdleConns:        100,              // max idle connections across all hosts
		MaxIdleConnsPerHost: 10,               // max idle connections per host
		MaxConnsPerHost:     100,              // max total connections per host (Go 1.11+)
		IdleConnTimeout:     90 * time.Second, // how long idle connections live

		// Timeouts
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		// Keep-alive
		DisableKeepAlives: false,

		// Buffer sizes
		WriteBufferSize: 64 * 1024, // 64KB write buffer
		ReadBufferSize:  64 * 1024, // 64KB read buffer
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // overall request timeout (includes redirects)
		// Note: per-request timeouts via context are usually better than global Timeout
	}
}

// Making requests with per-request context timeout (preferred)
func makeRequest(ctx context.Context, client *http.Client, url string) (string, error) {
	// Per-request timeout via context (preferred over client.Timeout)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Limit response body read to prevent OOM
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB max
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// =============================================================================
// PART 4: Connection Hijacking — Raw TCP from HTTP
// =============================================================================
// Upgrade HTTP connection to raw TCP (used by WebSocket, HTTP/2, CONNECT proxy)

func handleHijack(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, buf, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Now we have raw TCP — send custom protocol data
	buf.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	buf.WriteString("Connection: Upgrade\r\n")
	buf.WriteString("Upgrade: custom-protocol\r\n\r\n")
	buf.Flush()

	// Read/write raw frames on the upgraded connection
	_ = buf
}

// =============================================================================
// PART 5: TCP Keep-Alive & Connection Health
// =============================================================================

func demonstrateKeepAlive() {
	fmt.Println("\n=== TCP Keep-Alive ===")

	// When you dial, enable TCP keep-alive
	dialer := net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second, // send keep-alive every 30s
	}

	_ = dialer

	// For listeners, set keep-alive on accepted connections:
	// tcpConn := conn.(*net.TCPConn)
	// tcpConn.SetKeepAlive(true)
	// tcpConn.SetKeepAlivePeriod(30 * time.Second)

	fmt.Println("TCP keep-alive prevents silent connection drops in:")
	fmt.Println("  - Load balancers (AWS ALB default: 60s idle timeout)")
	fmt.Println("  - Firewalls (may drop idle connections after 5min)")
	fmt.Println("  - NAT gateways")
}

// =============================================================================
// PART 6: Line-based text protocol (like Redis RESP or SMTP)
// =============================================================================

func startLineServer(ctx context.Context, addr string, wg *sync.WaitGroup) {
	defer wg.Done()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Listen error: %v\n", err)
		return
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go handleLineConn(conn)
	}
}

func handleLineConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch strings.ToUpper(parts[0]) {
		case "PING":
			fmt.Fprintf(conn, "+PONG\r\n")
		case "ECHO":
			msg := strings.Join(parts[1:], " ")
			fmt.Fprintf(conn, "+%s\r\n", msg)
		case "QUIT":
			fmt.Fprintf(conn, "+BYE\r\n")
			return
		default:
			fmt.Fprintf(conn, "-ERR unknown command\r\n")
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start TCP server
	fmt.Println("=== Custom Binary Protocol (Length-Prefixed) ===")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		startTCPServer(ctx, "127.0.0.1:9000")
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Run client
	err := tcpClient("127.0.0.1:9000", []string{"Hello", "World", "Go Networking"})
	if err != nil {
		fmt.Printf("Client error: %v\n", err)
	}

	// Start line-based server
	fmt.Println("\n=== Line-Based Text Protocol ===")
	wg.Add(1)
	go startLineServer(ctx, "127.0.0.1:9001", &wg)
	time.Sleep(100 * time.Millisecond)

	// Line protocol client
	conn, err := net.DialTimeout("tcp", "127.0.0.1:9001", 2*time.Second)
	if err == nil {
		fmt.Fprintf(conn, "PING\r\n")
		reader := bufio.NewReader(conn)
		resp, _ := reader.ReadString('\n')
		fmt.Printf("  PING → %s", resp)

		fmt.Fprintf(conn, "ECHO hello world\r\n")
		resp, _ = reader.ReadString('\n')
		fmt.Printf("  ECHO → %s", resp)

		fmt.Fprintf(conn, "QUIT\r\n")
		resp, _ = reader.ReadString('\n')
		fmt.Printf("  QUIT → %s", resp)
		conn.Close()
	}

	// HTTP info
	fmt.Println("\n=== HTTP Production Settings ===")
	client := createProductionHTTPClient()
	_ = client
	srv := createProductionHTTPServer()
	_ = srv

	fmt.Println("HTTP Server timeouts:")
	fmt.Printf("  ReadTimeout:       %v\n", srv.ReadTimeout)
	fmt.Printf("  ReadHeaderTimeout: %v\n", srv.ReadHeaderTimeout)
	fmt.Printf("  WriteTimeout:      %v\n", srv.WriteTimeout)
	fmt.Printf("  IdleTimeout:       %v\n", srv.IdleTimeout)

	demonstrateKeepAlive()

	cancel()
	// Give servers time to shut down
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n=== NETWORKING KEY INSIGHTS ===")
	fmt.Println("1. Always use length-prefixed framing for binary protocols (not delimiters)")
	fmt.Println("2. ALWAYS set timeouts on server and client — defaults are unlimited")
	fmt.Println("3. Use per-request context.WithTimeout, not http.Client.Timeout")
	fmt.Println("4. Limit response body reads with io.LimitReader to prevent OOM")
	fmt.Println("5. Configure connection pools (MaxIdleConnsPerHost) for high-throughput")
	fmt.Println("6. Enable TCP keep-alive to detect dead connections through LBs/NATs")
	fmt.Println("7. Set MaxHeaderBytes to prevent header-based DoS")
}
