package sescapi

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	BaseURL        = "https://www.sescsp.org.br/wp-json/wp/v1"
	ModeUnidade    = "unidade"
	ModeAcesso     = "acesso"
	ModeLinguagens = "tipos_linguagens"
)

type EventsQuery struct {
	Units         []string
	Categories    []string
	Free          string
	Online        string
	Audience      string
	Activity      string
	Language      string
	ActivityTypes []string
	From          string
	To            string
	PerPage       int
	Page          int
}

type DinamicoQuery struct {
	Units         []string
	Categories    []string
	Free          string
	Online        string
	Audience      string
	ActivityTypes []string
	Languages     []string
	Subcategory   bool
	Mode          string
}

func EventsURL(q EventsQuery) (string, error) {
	if q.Audience == "" {
		q.Audience = "adulto"
	}
	if q.PerPage <= 0 {
		q.PerPage = 40
	}
	if q.Page <= 0 {
		q.Page = 1
	}

	values := url.Values{}
	values.Set("local", strings.Join(q.Units, ","))
	values.Set("categoria", strings.Join(q.Categories, ","))
	values.Set("gratuito", q.Free)
	values.Set("online", q.Online)
	values.Set("publico", q.Audience)
	values.Set("atividade", strings.Join(q.ActivityTypes, ","))
	values.Set("linguagem", q.Language)
	values.Set("data_inicial", q.From)
	values.Set("data_final", q.To)
	values.Set("tipo", "atividade")
	values.Set("dinamico", "true")
	values.Set("ppp", strconv.Itoa(q.PerPage))
	values.Set("page", strconv.Itoa(q.Page))
	return BaseURL + "/atividades/filter?" + values.Encode(), nil
}

func DinamicoURL(q DinamicoQuery) (string, error) {
	if q.Audience == "" {
		q.Audience = "adulto"
	}
	values := url.Values{}
	values.Set("unidades", strings.Join(q.Units, ","))
	values.Set("categorias", strings.Join(q.Categories, ","))
	values.Set("gratuito", q.Free)
	values.Set("online", q.Online)
	values.Set("publico_tag", q.Audience)
	values.Set("tipos_atividades", strings.Join(q.ActivityTypes, ","))
	values.Set("tipos_linguagens", strings.Join(q.Languages, ","))
	if q.Subcategory {
		values.Set("subcategoria", "true")
	}
	values.Set("modes", q.Mode)
	return BaseURL + "/dinamico?" + values.Encode(), nil
}

// UnidadesAtividadesURL is the canonical WordPress roster for venues (opening
// hours); use this instead of dinamico?modes=unidade when listing venues.
func UnidadesAtividadesURL() string {
	return BaseURL + "/unidades-atividades"
}
