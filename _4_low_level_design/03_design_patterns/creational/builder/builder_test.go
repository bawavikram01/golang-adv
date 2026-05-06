package builder

import "testing"

func TestRequestBuilder_FluentAPI(t *testing.T) {
	req, err := NewRequestBuilder("POST", "https://api.example.com/users").
		Header("Content-Type", "application/json").
		Header("Authorization", "Bearer token123").
		Body(`{"name":"Vikram"}`).
		Timeout(10).
		Retries(3).
		Build()

	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if req.Method != "POST" {
		t.Errorf("Method = %q", req.Method)
	}
	if req.URL != "https://api.example.com/users" {
		t.Errorf("URL = %q", req.URL)
	}
	if len(req.Headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(req.Headers))
	}
	if req.Timeout != 10 {
		t.Errorf("Timeout = %d, want 10", req.Timeout)
	}
	if req.Retries != 3 {
		t.Errorf("Retries = %d, want 3", req.Retries)
	}
}

func TestRequestBuilder_Defaults(t *testing.T) {
	req, err := NewGETRequest("https://example.com").Build()
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if req.Method != "GET" {
		t.Errorf("Method = %q, want GET", req.Method)
	}
	if req.Timeout != 30 {
		t.Errorf("default Timeout = %d, want 30", req.Timeout)
	}
}

func TestRequestBuilder_POST_Convenience(t *testing.T) {
	req, err := NewPOSTRequest("https://api.com/data", `{"key":"val"}`).Build()
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type header not set")
	}
	if req.Body != `{"key":"val"}` {
		t.Errorf("Body = %q", req.Body)
	}
}

func TestRequestBuilder_ValidationErrors(t *testing.T) {
	_, err := NewRequestBuilder("", "").
		Timeout(-1).
		Retries(-5).
		Build()

	if err == nil {
		t.Fatal("expected build error")
	}
}
