package what

// CulturalBundleSlugs are the API activity slugs bundled when --what cultural
// (or when --what is omitted and defaults to cultural). "teatro" is mapped to
// ShowsPerformancesActivity before the request is sent.
var CulturalBundleSlugs = []string{
	"teatro",
	"cinema",
	"teatro-cursos-e-oficinas",
	"cinema-cursos-e-oficinas",
	"musica",
	"danca",
	"artes-visuais",
	"literatura-cursos-e-oficinas",
	"tecnologias-e-artes",
}

// ShowsPerformancesActivity is the SESC filter slug for the “shows /
// espetáculos / performances” bucket (teatro is a linguagem inside it, not
// atividade=teatro).
const ShowsPerformancesActivity = "shows-espetaculos-e-performances"

// ProfileTokens are whole-expression keywords for --what / profile (not
// combinable with commas).
var ProfileTokens = []string{
	"all", "any", "todos", "todas",
	"cultural",
	"sports", "esportes",
	"teatro",
}

// ExpressionSynonyms map friendly tokens to canonical API slugs before
// validation and resolution. Keys must be lowercased.
var ExpressionSynonyms = map[string]string{
	"oficina":  "cursos-e-oficinas",
	"sports":   "esporte-e-atividade-fisica",
	"esportes": "esporte-e-atividade-fisica",
}

// TypoHints suggests replacements for common mistakes (ASCII lower keys).
var TypoHints = map[string]string{
	"espetaculo":   ShowsPerformancesActivity,
	"espetaculos":  ShowsPerformancesActivity,
	"espataculo":   ShowsPerformancesActivity,
	"performances": ShowsPerformancesActivity,
}
