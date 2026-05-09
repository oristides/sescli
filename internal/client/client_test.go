package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetJSONSendsAcceptAndDecodes(t *testing.T) {
	var accept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]string{{"ok": "yes"}})
	}))
	defer server.Close()

	c := New(Options{Timeout: time.Second, Retries: 1})
	var out []map[string]string
	if err := c.GetJSON(server.URL, &out); err != nil {
		t.Fatal(err)
	}

	if accept != "application/json" {
		t.Fatalf("unexpected accept: %q", accept)
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
