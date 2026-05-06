// Package builder demonstrates the Builder pattern.
//
// INTENT: Separate the construction of a complex object from its representation.
// Allows step-by-step construction with a fluent API.
//
// WHEN TO USE:
//   - Object has many optional fields
//   - Constructor would need too many parameters
//   - You want to enforce valid construction
//   - You need to build different representations of the same type
package builder

import (
	"fmt"
	"strings"
)

// ──────────────────────────────────────────────
// The complex object
// ──────────────────────────────────────────────

type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
	Timeout int // seconds
	Retries int
}

func (r HTTPRequest) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s", r.Method, r.URL))
	if len(r.Headers) > 0 {
		sb.WriteString(fmt.Sprintf(" headers=%d", len(r.Headers)))
	}
	if r.Body != "" {
		sb.WriteString(fmt.Sprintf(" body=%d bytes", len(r.Body)))
	}
	return sb.String()
}

// ──────────────────────────────────────────────
// The Builder — fluent API
// ──────────────────────────────────────────────

type RequestBuilder struct {
	request HTTPRequest
	errors  []string
}

func NewRequestBuilder(method, url string) *RequestBuilder {
	return &RequestBuilder{
		request: HTTPRequest{
			Method:  method,
			URL:     url,
			Headers: make(map[string]string),
			Timeout: 30, // sensible default
			Retries: 0,
		},
	}
}

func (b *RequestBuilder) Header(key, value string) *RequestBuilder {
	b.request.Headers[key] = value
	return b
}

func (b *RequestBuilder) Body(body string) *RequestBuilder {
	b.request.Body = body
	return b
}

func (b *RequestBuilder) Timeout(seconds int) *RequestBuilder {
	if seconds <= 0 {
		b.errors = append(b.errors, "timeout must be positive")
		return b
	}
	b.request.Timeout = seconds
	return b
}

func (b *RequestBuilder) Retries(n int) *RequestBuilder {
	if n < 0 {
		b.errors = append(b.errors, "retries cannot be negative")
		return b
	}
	b.request.Retries = n
	return b
}

func (b *RequestBuilder) Build() (HTTPRequest, error) {
	if b.request.Method == "" {
		b.errors = append(b.errors, "method is required")
	}
	if b.request.URL == "" {
		b.errors = append(b.errors, "URL is required")
	}
	if len(b.errors) > 0 {
		return HTTPRequest{}, fmt.Errorf("build failed: %s", strings.Join(b.errors, "; "))
	}
	return b.request, nil
}

// ──────────────────────────────────────────────
// Convenience builders (Director pattern)
// ──────────────────────────────────────────────

func NewGETRequest(url string) *RequestBuilder {
	return NewRequestBuilder("GET", url)
}

func NewPOSTRequest(url string, body string) *RequestBuilder {
	return NewRequestBuilder("POST", url).
		Header("Content-Type", "application/json").
		Body(body)
}
