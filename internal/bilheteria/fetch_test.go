package bilheteria

import (
	"encoding/json"
	"testing"
)

func TestFirstSessionMapsToEventPricing(t *testing.T) {
	body := `{"sessoes":[{"gratuito":false,"valorInteiraFmt":"R$ 30,00","valorMeiaFmt":"R$ 15,00","valorComerciarioFmt":"R$ 10,00","statusIngresso":"Disponível","qtdeIngressosWeb":12,"qtdeIngressosRede":3}]}`
	var dto activityDTO
	if err := json.Unmarshal([]byte(body), &dto); err != nil {
		t.Fatal(err)
	}
	if len(dto.Sessoes) != 1 || dto.Sessoes[0].ValorInteiraFmt != "R$ 30,00" {
		t.Fatalf("%#v", dto.Sessoes)
	}
}
