// LiteLLM proxy client.
//
// PR-B grows this from PR-A's passthrough into a metadata-injecting
// forwarder. The gateway needs to attach tenant_id and a
// galileo-generated request_id to every LLM call so LiteLLM's
// generic_api callback (PR-B's cost_events.go webhook receiver) can
// associate the StandardLoggingPayload back to the right tenant and
// the right Drift-2 correlation key.
//
// Channel choice (verified at planning time): LiteLLM's OSS logging
// path only populates StandardLoggingMetadata.requester_metadata from
// the request body's `metadata` field. The headers→requester_custom_headers
// path exists in enterprise/litellm_enterprise/ but not in OSS. So we
// inject metadata into the request body, not as HTTP headers.
//
// Wire shape:
//
//	gateway parses request body JSON
//	→ injects metadata.galileo_tenant_id + metadata.galileo_request_id
//	→ re-serializes
//	→ POSTs to LiteLLM
//	→ LiteLLM passes metadata through to logging callback
//	→ generic_api POSTs StandardLoggingPayload to our /internal/cost-events
//	→ webhook reads payload.metadata.requester_metadata.galileo_request_id
//	→ INSERT cost_events ON CONFLICT (request_id) DO NOTHING

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// MetadataKeyTenantID is the JSON key set in the LiteLLM request body's
// `metadata` object that carries the gateway's tenant identifier. Read
// back from the webhook payload's requester_metadata.
const MetadataKeyTenantID = "galileo_tenant_id"

// MetadataKeyRequestID is the JSON key for the gateway-generated UUIDv7
// that becomes cost_events.request_id (PRIMARY KEY for idempotency).
const MetadataKeyRequestID = "galileo_request_id"

// LiteLLMClient forwards chat-completion requests to LiteLLM with
// gateway-injected metadata. The base URL targets the LiteLLM service
// (http://litellm:4000 in compose, http://localhost:4000 from host).
type LiteLLMClient struct {
	base   *url.URL
	client *http.Client
}

// NewLiteLLMClient parses baseURL and configures an http.Client with a
// 60s timeout — long enough for a slow completion, short enough that a
// wedged upstream surfaces as 502 within a minute.
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

// Forward parses body as JSON, injects gateway-side metadata, and POSTs
// to base + path. tenantID and requestID land in the body's `metadata`
// object so LiteLLM's OSS logging path surfaces them in
// requester_metadata for the generic_api callback. Caller closes the
// returned response body.
func (c *LiteLLMClient) Forward(ctx context.Context, path string, body []byte, tenantID, requestID string) (*http.Response, error) {
	injected, err := injectMetadata(body, tenantID, requestID)
	if err != nil {
		return nil, err
	}
	target := *c.base
	target.Path = path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(injected))
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

// injectMetadata parses body as a JSON object, merges
// {galileo_tenant_id, galileo_request_id} into a `metadata` field (or
// creates one), and re-serializes. If body is not a JSON object the
// gateway returns the parse error to the caller — 400 to the client —
// rather than silently dropping the request.
func injectMetadata(body []byte, tenantID, requestID string) ([]byte, error) {
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, fmt.Errorf("parse request body as JSON object: %w", err)
	}
	metadata, _ := obj["metadata"].(map[string]any)
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata[MetadataKeyTenantID] = tenantID
	metadata[MetadataKeyRequestID] = requestID
	obj["metadata"] = metadata
	out, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("re-serialize request body: %w", err)
	}
	return out, nil
}
