package format

import (
	"encoding/json"
	"strings"
	"testing"

	"sescli/internal/normalize"
)

func TestWhatsAppKeepsOneShortUsefulLinePerEvent(t *testing.T) {
	out := WhatsApp([]normalize.Event{{
		Title:      "Mostra de Cinema",
		URL:        "https://sescsp.org.br/e",
		Venue:      "Sesc Consolacao",
		PriceLabel: "Gratis",
		Categories: []string{"Cinema"},
	}})

	if strings.Contains(out, "\n\n") {
		t.Fatalf("whatsapp output should not contain blank lines: %q", out)
	}
	wantParts := []string{"Mostra de Cinema", "Gratis", "Sesc Consolacao", "https://sescsp.org.br/e"}
	for _, part := range wantParts {
		if !strings.Contains(out, part) {
			t.Fatalf("expected %q in %q", part, out)
		}
	}
	if len(out) > 140 {
		t.Fatalf("single event line should stay terse, got %d chars: %q", len(out), out)
	}
}

func TestWhatsAppExplainsEmptyResults(t *testing.T) {
	out := WhatsApp(nil)
	if out != "Nenhum evento encontrado." {
		t.Fatalf("unexpected empty output: %q", out)
	}
}

func TestJSONCompactOmitsEmptyRawByDefault(t *testing.T) {
	payload := Response("events", []normalize.Event{{ID: "1", Title: "A"}}, Meta{Total: 1})

	out, err := JSON(payload, false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "\n") {
		t.Fatalf("compact json should not include newlines: %q", out)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if _, ok := decoded["_meta"]; !ok {
		t.Fatalf("expected _meta envelope in %v", decoded)
	}
}
