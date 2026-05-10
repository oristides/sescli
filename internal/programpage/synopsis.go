package programpage

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	xhtml "golang.org/x/net/html"
)

const (
	defaultUserAgent = "sescli/1.0 (programação reader; +https://www.sescsp.org.br/programacao/)"
	maxHTMLBytes     = 2 << 20 // 2 MiB safety cap on HTML download
)

// FetchSynopsis downloads the public programação page and extracts main copy.
// It prefers the editorial block .principal--post--conteudo, then falls back
// to og:description. Layout is site-specific and may break if templates change.
func FetchSynopsis(ctx context.Context, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/html,*/*;q=0.8")
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Referer", "https://www.sescsp.org.br/programacao/")

	c := &http.Client{Timeout: 25 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch program page: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxHTMLBytes))
	if err != nil {
		return "", err
	}

	doc, err := xhtml.Parse(bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("parse html: %w", err)
	}

	text := extractPrincipalPostConteudo(doc)
	if text == "" {
		text = extractOGDescription(doc)
	}
	if text == "" {
		return "", fmt.Errorf("no program text found in HTML (unknown layout for %s)", pageURL)
	}

	text = html.UnescapeString(text)
	text = normalizeSpace(text)
	return text, nil
}

func extractPrincipalPostConteudo(doc *xhtml.Node) string {
	n := findFirstNodeWithClassFragment(doc, "principal--post--conteudo")
	if n == nil {
		return ""
	}
	return textUnder(n)
}

func extractOGDescription(doc *xhtml.Node) string {
	var found string
	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if found != "" {
			return
		}
		if n.Type == xhtml.ElementNode && n.Data == "meta" {
			var prop, content string
			for _, a := range n.Attr {
				switch a.Key {
				case "property":
					prop = a.Val
				case "content":
					content = a.Val
				}
			}
			if prop == "og:description" && strings.TrimSpace(content) != "" {
				found = content
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return found
}

func findFirstNodeWithClassFragment(root *xhtml.Node, classFragment string) *xhtml.Node {
	if root.Type == xhtml.ElementNode {
		for _, a := range root.Attr {
			if a.Key == "class" && strings.Contains(a.Val, classFragment) {
				return root
			}
		}
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if n := findFirstNodeWithClassFragment(c, classFragment); n != nil {
			return n
		}
	}
	return nil
}

func textUnder(root *xhtml.Node) string {
	var b strings.Builder
	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		switch n.Type {
		case xhtml.TextNode:
			b.WriteString(n.Data)
		case xhtml.ElementNode:
			if n.Data == "script" || n.Data == "style" {
				return
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}
	walk(root)
	return b.String()
}

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
