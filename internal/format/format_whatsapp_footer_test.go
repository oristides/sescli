package format

import (
	"strings"
	"testing"
)

func TestWhatsAppQueryFooterSingleDayExplainsNarrowWindow(t *testing.T) {
	m := Meta{
		DateFrom: "2026-05-16",
		DateTo:   "2026-05-16",
		Total:    4,
		PerPage:  50,
	}
	out := WhatsAppQueryFooter(m)
	if !strings.Contains(out, "Período") || !strings.Contains(out, "só este dia") {
		t.Fatalf("expected single-day hint: %q", out)
	}
	if !strings.Contains(out, "--what all") {
		t.Fatalf("expected broadening hint: %q", out)
	}
	if !strings.Contains(out, "Nesta resposta") || !strings.Contains(out, "4") {
		t.Fatalf("expected count/limit explanation: %q", out)
	}
}

func TestWhatsAppQueryFooterWithReportedTotal(t *testing.T) {
	r := 42
	m := Meta{
		DateFrom:      "2026-05-16",
		DateTo:        "2026-05-17",
		Total:         40,
		PerPage:       40,
		TotalReported: &r,
	}
	out := WhatsAppQueryFooter(m)
	if !strings.Contains(out, "Total no filtro") || !strings.Contains(out, "42") {
		t.Fatalf("expected API total in footer: %q", out)
	}
}
