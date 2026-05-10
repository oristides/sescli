package normalize

// FilterReportedTotalPtr returns total.value from the root of an atividades/filter
// JSON payload (WordPress REST shape: { "total": { "value": N, "relation": "eq" } }).
func FilterReportedTotalPtr(raw any) *int {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	t, ok := m["total"].(map[string]any)
	if !ok {
		return nil
	}
	v, ok := t["value"]
	if !ok {
		return nil
	}
	n, ok := numberToInt(v)
	if !ok {
		return nil
	}
	return &n
}

func numberToInt(v any) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case int:
		return x, true
	case int64:
		return int(x), true
	default:
		return 0, false
	}
}
