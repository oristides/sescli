package normalize

import (
	"encoding/json"
	"testing"
)

func TestFilterReportedTotalPtr(t *testing.T) {
	rawJSON := `{"atividade":[],"total":{"value":42,"relation":"eq"}}`
	var raw any
	if err := json.Unmarshal([]byte(rawJSON), &raw); err != nil {
		t.Fatal(err)
	}
	p := FilterReportedTotalPtr(raw)
	if p == nil || *p != 42 {
		t.Fatalf("got %#v", p)
	}
	if FilterReportedTotalPtr(map[string]any{"atividade": []any{}}) != nil {
		t.Fatal("expected nil when total missing")
	}
	if FilterReportedTotalPtr(nil) != nil {
		t.Fatal("expected nil for nil raw")
	}
}
