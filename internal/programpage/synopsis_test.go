package programpage

import (
	"strings"
	"testing"

	xhtml "golang.org/x/net/html"
)

const sampleProgramHTML = `<!DOCTYPE html><html><head>
<meta property="og:description" content="Short fallback from OpenGraph." />
</head><body>
<div class="principal--post--row--programacao--release">
<div class="principal--post--conteudo">
<p>First paragraph about the <strong>show</strong>.</p>
<p><em>Note:</em> content continues.</p>
</div>
</div>
<script>alert(1)</script>
</body></html>`

func TestExtractPrincipalPostConteudo(t *testing.T) {
	doc, err := parseHTML(sampleProgramHTML)
	if err != nil {
		t.Fatal(err)
	}
	got := extractPrincipalPostConteudo(doc)
	if !strings.Contains(got, "First paragraph") || !strings.Contains(got, "show") {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, "alert") {
		t.Fatalf("script leaked: %q", got)
	}
}

func TestExtractOGDescriptionFallback(t *testing.T) {
	doc, err := parseHTML(`<html><head><meta property="og:description" content="Only OG here." /></head><body></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	if extractPrincipalPostConteudo(doc) != "" {
		t.Fatal("expected empty principal block")
	}
	if extractOGDescription(doc) != "Only OG here." {
		t.Fatal(extractOGDescription(doc))
	}
}

func parseHTML(s string) (*xhtml.Node, error) {
	return xhtml.Parse(strings.NewReader(s))
}
