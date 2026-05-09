package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ProgramacaoReferer = "https://www.sescsp.org.br/programacao/"

type Options struct {
	Timeout time.Duration
	Retries int
	Backoff time.Duration
}

type Client struct {
	http    *http.Client
	retries int
	backoff time.Duration
}

func New(options Options) *Client {
	if options.Timeout <= 0 {
		options.Timeout = 10 * time.Second
	}
	if options.Retries <= 0 {
		options.Retries = 2
	}
	if options.Backoff <= 0 {
		options.Backoff = 250 * time.Millisecond
	}
	return &Client{
		http:    &http.Client{Timeout: options.Timeout},
		retries: options.Retries,
		backoff: options.Backoff,
	}
}

func (c *Client) GetJSON(rawURL string, target any) error {
	var last error
	for attempt := 1; attempt <= c.retries; attempt++ {
		err := c.getJSON(rawURL, target)
		if err == nil {
			return nil
		}
		last = err
		if attempt < c.retries {
			time.Sleep(c.backoff * time.Duration(attempt))
		}
	}
	return last
}

func (c *Client) getJSON(rawURL string, target any) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", ProgramacaoReferer)
	req.Header.Set("User-Agent", "sescli/0.1 (+https://www.sescsp.org.br/programacao/)")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("transient HTTP %d", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}
