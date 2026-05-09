package presets

import (
	"fmt"
	"sort"
	"strings"
)

type Unit struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Group string `json:"group,omitempty"`
}

func AllUnits() []Unit {
	return append([]Unit(nil), units...)
}

func SearchUnits(query string) []Unit {
	needle := key(query)
	if needle == "" {
		return AllUnits()
	}
	matches := []Unit{}
	for _, unit := range units {
		if unit.ID == strings.TrimSpace(query) || strings.Contains(key(unit.Name), needle) || strings.Contains(key(unit.Slug), needle) {
			matches = append(matches, unit)
		}
	}
	if len(matches) == 0 {
		for _, unit := range units {
			if levenshtein(needle, key(unit.Name)) <= 2 || levenshtein(needle, key(unit.Slug)) <= 2 {
				matches = append(matches, unit)
			}
		}
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].Name < matches[j].Name })
	return matches
}

func ResolveUnitIDs(queries []string) ([]string, error) {
	out := []string{}
	for _, query := range queries {
		for _, part := range strings.Split(query, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			matches := SearchUnits(part)
			if len(matches) == 0 {
				return nil, fmt.Errorf("unknown unit %q (try: sescli units search %q)", part, part)
			}
			out = append(out, matches[0].ID)
		}
	}
	return dedupe(out), nil
}

func key(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "ã", "a", "â", "a",
		"é", "e", "ê", "e",
		"í", "i",
		"ó", "o", "ô", "o", "õ", "o",
		"ú", "u",
		"ç", "c",
		" ", "",
		"-", "",
		"_", "",
	)
	return replacer.Replace(value)
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			cur[j] = min3(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

var units = []Unit{
	{ID: "948", Name: "Ponto de Encontro", Slug: "ponto-de-encontro"},
	{ID: "2", Name: "24 de Maio", Slug: "24-de-maio", Group: "centro"},
	{ID: "43", Name: "Avenida Paulista", Slug: "avenida-paulista", Group: "centro"},
	{ID: "47", Name: "Belenzinho", Slug: "belenzinho", Group: "capital"},
	{ID: "49", Name: "Campo Limpo", Slug: "campo-limpo", Group: "capital"},
	{ID: "51", Name: "Centro de Pesquisa e Formação", Slug: "centro-de-pesquisa-e-formacao", Group: "centro"},
	{ID: "52", Name: "CineSesc", Slug: "cinesesc", Group: "centro"},
	{ID: "53", Name: "Consolação", Slug: "consolacao", Group: "centro"},
	{ID: "55", Name: "Interlagos", Slug: "interlagos", Group: "capital"},
	{ID: "56", Name: "Ipiranga", Slug: "ipiranga", Group: "capital"},
	{ID: "57", Name: "Itaquera", Slug: "itaquera", Group: "capital"},
	{ID: "58", Name: "Osasco", Slug: "osasco", Group: "capital"},
	{ID: "60", Name: "Pinheiros", Slug: "pinheiros", Group: "centro"},
	{ID: "61", Name: "Pompeia", Slug: "pompeia", Group: "centro"},
	{ID: "62", Name: "Santana", Slug: "santana", Group: "capital"},
	{ID: "63", Name: "Santo Amaro", Slug: "santo-amaro", Group: "capital"},
	{ID: "64", Name: "Santo André", Slug: "santo-andre", Group: "capital"},
	{ID: "65", Name: "São Caetano", Slug: "sao-caetano", Group: "capital"},
	{ID: "66", Name: "Vila Mariana", Slug: "vila-mariana", Group: "centro"},
	{ID: "71", Name: "Guarulhos", Slug: "guarulhos", Group: "capital"},
	{ID: "80", Name: "Mogi das Cruzes", Slug: "mogi-das-cruzes", Group: "capital"},
	{ID: "730", Name: "Casa Verde", Slug: "casa-verde", Group: "capital"},
	{ID: "761", Name: "14 Bis", Slug: "14-bis", Group: "centro"},
	{ID: "25", Name: "Araraquara", Slug: "araraquara"},
	{ID: "26", Name: "Bauru", Slug: "bauru"},
	{ID: "28", Name: "Birigui", Slug: "birigui"},
	{ID: "29", Name: "Campinas", Slug: "campinas"},
	{ID: "31", Name: "Jundiaí", Slug: "jundiai"},
	{ID: "32", Name: "Piracicaba", Slug: "piracicaba"},
	{ID: "33", Name: "Presidente Prudente", Slug: "presidente-prudente"},
	{ID: "34", Name: "Registro", Slug: "registro"},
	{ID: "35", Name: "Ribeirão Preto", Slug: "ribeirao-preto"},
	{ID: "36", Name: "Rio Preto", Slug: "rio-preto"},
	{ID: "38", Name: "Sorocaba", Slug: "sorocaba"},
	{ID: "40", Name: "São José dos Campos", Slug: "sao-jose-dos-campos"},
	{ID: "41", Name: "Taubaté", Slug: "taubate"},
	{ID: "42", Name: "São Carlos", Slug: "sao-carlos"},
	{ID: "1005", Name: "Franca", Slug: "franca"},
	{ID: "37", Name: "Santos", Slug: "santos"},
}
