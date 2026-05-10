package bilheteria

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sescli/internal/normalize"
)

const (
	portalBilheteriaPath = "https://portal.sescsp.org.br/bilheteria/atividade.action"
	defaultReferer       = "https://www.sescsp.org.br/programacao/"
	maxResponseBytes     = 2 << 20
)

type activityDTO struct {
	Sessoes []sessionDTO `json:"sessoes"`
}

type sessionDTO struct {
	Gratuito            bool   `json:"gratuito"`
	ValorInteiraFmt     string `json:"valorInteiraFmt"`
	ValorMeiaFmt        string `json:"valorMeiaFmt"`
	ValorComerciarioFmt string `json:"valorComerciarioFmt"`
	StatusIngresso      string `json:"statusIngresso"`
	QtdeIngressosWeb    int    `json:"qtdeIngressosWeb"`
	QtdeIngressosRede   int    `json:"qtdeIngressosRede"`
}

// FetchActivityPricing loads the first session’s ticket prices from the Java bilheteria API
// using the same WordPress admin-ajax proxy the site front-end uses (idAtividade = id_java from list rows).
func FetchActivityPricing(ctx context.Context, javaID string, referer string) (*normalize.EventPricing, error) {
	javaID = strings.TrimSpace(javaID)
	if javaID == "" {
		return nil, fmt.Errorf("bilheteria: empty id_java")
	}
	if referer == "" {
		referer = defaultReferer
	}
	route := portalBilheteriaPath + "?idAtividade=" + url.QueryEscape(javaID)
	q := url.Values{}
	q.Set("action", "sesc_requester_proxy")
	q.Set("method", "GET")
	q.Set("route", route)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.sescsp.org.br/wp-admin/admin-ajax.php?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json,*/*")
	req.Header.Set("User-Agent", "sescli/1.0 (+bilheteria pricing read)")
	req.Header.Set("Referer", referer)

	c := &http.Client{Timeout: 20 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bilheteria proxy: HTTP %d", resp.StatusCode)
	}
	var dto activityDTO
	if err := json.Unmarshal(body, &dto); err != nil {
		return nil, fmt.Errorf("bilheteria: %w", err)
	}
	if len(dto.Sessoes) == 0 {
		return nil, fmt.Errorf("bilheteria: no sessoes")
	}
	s := dto.Sessoes[0]
	return &normalize.EventPricing{
		Gratuito:          s.Gratuito,
		ValorInteira:      s.ValorInteiraFmt,
		ValorMeia:         s.ValorMeiaFmt,
		ValorComerciario:  s.ValorComerciarioFmt,
		StatusIngresso:    s.StatusIngresso,
		QtdeIngressosWeb:  s.QtdeIngressosWeb,
		QtdeIngressosRede: s.QtdeIngressosRede,
	}, nil
}
