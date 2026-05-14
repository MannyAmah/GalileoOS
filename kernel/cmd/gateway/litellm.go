// LiteLLM passthrough client.
//
// PR-A scope: forward the request body to LiteLLM and return the
// response unchanged. Usage-event capture (parsing the response for
// token counts, writing cost_events rows) is PR-B's job — this client
// is deliberately minimal so PR-B can wrap it without rewriting.

package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// LiteLLMClient proxies requests to the LiteLLM container. The base URL
// points at the LiteLLM service (e.g., http://litellm:4000 in compose,
// http://localhost:4000 from the host).
type LiteLLMClient struct {
	base   *url.URL
	client *http.Client
}

// NewLiteLLMClient parses baseURL and configures an http.Client with a
// 60s timeout — long enough for a streamed-but-not-streaming completion,
// short enough that a wedged upstream surfaces as 502 within a minute.
func NewLiteLLMClient(baseURL string) (*LiteLLMClient, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse litellm base url: %w", err)
	}
	return &LiteLLMClient{
		base:   u,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Forward POSTs body to base + path on LiteLLM. The caller is responsible
// for closing the returned response body.
func (c *LiteLLMClient) Forward(ctx context.Context, path string, body []byte) (*http.Response, error) {
	target := *c.base
	target.Path = path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build litellm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("litellm request: %w", err)
	}
	return resp, nil
}
