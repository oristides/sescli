package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetJSONSendsBrowserHeadersAndDecodes(t *testing.T) {
	var referer, accept, ua string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		referer = r.Header.Get("Referer")
		accept = r.Header.Get("Accept")
		ua = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]string{{"ok": "yes"}})
	}))
	defer server.Close()

	c := New(Options{Timeout: time.Second, Retries: 1})
	var out []map[string]string
	if err := c.GetJSON(server.URL, &out); err != nil {
		t.Fatal(err)
	}

	if referer != ProgramacaoReferer {
		t.Fatalf("missing referer: %q", referer)
	}
	if accept != "application/json" {
		t.Fatalf("unexpected accept: %q", accept)
	}
	if ua == "" {
		t.Fatalf("expected user-agent")
	}
	if out[0]["ok"] != "yes" {
		t.Fatalf("unexpected response %#v", out)
	}
}

func TestGetJSONRetriesTransientServerErrors(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := New(Options{Timeout: time.Second, Retries: 2, Backoff: time.Millisecond})
	var out map[string]bool
	if err := c.GetJSON(server.URL, &out); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatalf("expected one retry, got %d attempts", attempts)
	}
}
