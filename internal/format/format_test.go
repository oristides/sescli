package format

import (
	"strings"
	"testing"

	"sescli/internal/normalize"
)

func TestWhatsAppIncludesSummarySecondLine(t *testing.T) {
	ev := normalize.Event{
		Title:   "Show",
		Summary: "com Fulano",
		URL:     "https://www.sescsp.org.br/programacao/x",
	}
	out := WhatsApp([]normalize.Event{ev})
	if !strings.Contains(out, "Show") || !strings.Contains(out, "com Fulano") {
		t.Fatalf("got %q", out)
	}
	lines := strings.Split(out, "\n")
	if len(lines) < 2 || !strings.HasPrefix(lines[1], "  ") {
		t.Fatalf("expected indented summary line: %q", out)
	}
}

func TestTableIncludesSummaryColumn(t *testing.T) {
	ev := normalize.Event{Title: "A", Summary: "sub", Synopsis: strings.Repeat("x", 200), Venue: "V"}
	row := strings.Split(Table([]normalize.Event{ev}), "\n")
	if len(row) < 2 {
		t.Fatal(row)
	}
	if !strings.Contains(row[0], "SUMMARY") || !strings.Contains(row[0], "SYNOPSIS") {
		t.Fatalf("header %q", row[0])
	}
	if !strings.Contains(row[1], "\tsub\t") {
		t.Fatalf("row %q", row[1])
	}
	if !strings.Contains(row[1], "...") {
		t.Fatalf("expected truncated synopsis: %q", row[1])
	}
}

func TestWhatsAppIncludesSynopsisBlock(t *testing.T) {
	ev := normalize.Event{Title: "T", Synopsis: "Long body here.", URL: "https://x"}
	out := WhatsApp([]normalize.Event{ev})
	if !strings.Contains(out, "Long body here.") {
		t.Fatalf("got %q", out)
	}
}
