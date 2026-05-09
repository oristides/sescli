package normalize

import "testing"

func TestEventFromRawPrefersUsefulSuccinctFields(t *testing.T) {
	raw := map[string]any{
		"ID":             float64(123),
		"post_title":     "Mostra de Cinema",
		"permalink":      "https://www.sescsp.org.br/programacao/mostra/",
		"unidade_nome":   "Sesc Consolacao",
		"data_inicio":    "2026-05-08",
		"hora_inicio":    "20h",
		"gratuito":       "1",
		"preco":          "",
		"categorias":     []any{"Cinema"},
		"tipo_atividade": "cinema",
		"post_content":   "<p>Long <strong>description</strong>.</p>",
	}

	event := EventFromRaw(raw, false)

	if event.ID != "123" {
		t.Fatalf("expected ID 123, got %q", event.ID)
	}
	if event.Title != "Mostra de Cinema" {
		t.Fatalf("unexpected title: %q", event.Title)
	}
	if event.URL == "" || event.Venue != "Sesc Consolacao" {
		t.Fatalf("expected url and venue, got %#v", event)
	}
	if !event.Free {
		t.Fatalf("expected free event")
	}
	if len(event.Categories) != 1 || event.Categories[0] != "Cinema" {
		t.Fatalf("unexpected categories: %#v", event.Categories)
	}
	if event.Raw != nil {
		t.Fatalf("raw payload should be omitted unless requested")
	}
}

func TestEventsFromRawSkipsInvalidItems(t *testing.T) {
	events := EventsFromRaw([]any{
		map[string]any{"ID": "1", "post_title": "Valid"},
		"not an object",
	}, false)

	if len(events) != 1 {
		t.Fatalf("expected 1 normalized event, got %d", len(events))
	}
}

func TestEventsFromRawAcceptsLiveObjectEnvelope(t *testing.T) {
	events := EventsFromRaw(map[string]any{
		"atividade": []any{
			map[string]any{
				"id":                 float64(1209154),
				"titulo":             "Boate Class",
				"link":               "/programacao/boate-class",
				"gratuito":           "Atividade gratuita",
				"dataProxSessao":     "2026-05-23T20:30",
				"dataPrimeiraSessao": "2026-05-23T19:00",
				"unidade": []any{
					map[string]any{"name": "Avenida Paulista"},
				},
				"tipos_linguagens": []any{
					map[string]any{"titulo": "Cursos e Oficinas"},
				},
			},
		},
	}, false)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	event := events[0]
	if event.Title != "Boate Class" || event.URL != "https://www.sescsp.org.br/programacao/boate-class" {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
	if event.Venue != "Avenida Paulista" || !event.Free {
		t.Fatalf("expected venue/free from live payload, got %#v", event)
	}
	if len(event.Categories) != 1 || event.Categories[0] != "Cursos e Oficinas" {
		t.Fatalf("expected language category, got %#v", event.Categories)
	}
}

func TestEventFromRawTurnsObjectCategoriesIntoTitlesAndShortPrices(t *testing.T) {
	event := EventFromRaw(map[string]any{
		"id":       "1",
		"titulo":   "Peca",
		"gratuito": "Atividade paga",
		"categorias": []any{
			map[string]any{"titulo": "Teatro"},
		},
	}, false)

	if event.PriceLabel != "Pago" {
		t.Fatalf("expected short paid label, got %q", event.PriceLabel)
	}
	if len(event.Categories) != 1 || event.Categories[0] != "Teatro" {
		t.Fatalf("expected category title, got %#v", event.Categories)
	}
}

func TestUnitsFromRawAcceptsGroupedLiveEnvelope(t *testing.T) {
	units := UnitsFromRaw(map[string]any{
		"unidades": map[string]any{
			"capital": []any{
				map[string]any{"groupID": "2", "groupName": "24 de Maio", "groupLink": "24-de-maio"},
			},
		},
	}, false)

	if len(units) != 1 {
		t.Fatalf("expected 1 unit, got %d", len(units))
	}
	if units[0].ID != "2" || units[0].Name != "24 de Maio" {
		t.Fatalf("unexpected unit: %#v", units[0])
	}
}

func TestUnitsFromRawUnidadesAtividadesRowShape(t *testing.T) {
	un := UnitFromRaw(map[string]any{
		"name":        "Bom Retiro",
		"group_id":    "48",
		"group_slug":  "bom-retiro",
		"description": "capital",
	}, false)

	if un.ID != "48" || un.Name != "Bom Retiro" || un.Slug != "bom-retiro" {
		t.Fatalf("unexpected normalization: %#v", un)
	}
	if un.APISegment != "capital" {
		t.Fatalf("expected api_segment capital, got %q", un.APISegment)
	}
	list := UnitsFromRaw([]any{
		map[string]any{"name": "Bom Retiro", "group_id": "48"},
		map[string]any{"oops": "ignored"},
	}, false)
	if len(list) != 1 {
		t.Fatalf("expected to skip malformed row, got %d", len(list))
	}
}
