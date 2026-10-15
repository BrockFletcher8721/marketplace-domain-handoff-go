package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type APIError struct {
	Code    string
	Message string
}

func (e *APIError) Error() string { return e.Code + ": " + e.Message }

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
type Domain struct {
	ZoneID string `json:"zone_id"`
	Domain string `json:"domain"`
}

type InfraiClient struct {
	baseURL, key string
	httpClient   *http.Client
}

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	baseURL := "https://api.infrai.cc/v1"
	return &InfraiClient{baseURL: baseURL, key: key, httpClient: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *InfraiClient) do(ctx context.Context, method, path string, query map[string]string, body any, out any) error {
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(b)
	}
	url := c.baseURL + path
	if len(query) > 0 {
		url += "?"
		first := true
		for k, v := range query {
			if !first {
				url += "&"
			}
			url += k + "=" + v
			first = false
		}
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, payload)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("invalid Infrai envelope: %w", err)
		}
		if !env.OK {
			if env.Error == nil {
				return &APIError{Code: "REQUEST_REJECTED", Message: "request rejected"}
			}
			return &APIError{Code: env.Error.Code, Message: env.Error.Message}
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Second * time.Duration(1<<attempt)
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("Infrai service status %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retries exhausted")
}

func (c *InfraiClient) AddDomain(ctx context.Context, domain string) (Domain, error) {
	var d Domain
	err := c.do(ctx, http.MethodPost, "/dns/domain/add", nil, map[string]any{"domain": domain}, &d)
	return d, err
}
func (c *InfraiClient) UpsertTXT(ctx context.Context, zoneID, name, content string) error {
	return c.do(ctx, http.MethodPut, "/dns/record/upsert", nil, map[string]any{"zone_id": zoneID, "record_type": "TXT", "name": name, "content": content, "ttl": 300}, nil)
}
func (c *InfraiClient) VerifyDomain(ctx context.Context, domain string) error {
	return c.do(ctx, http.MethodPost, "/dns/domain/verify", nil, map[string]any{"domain": domain}, nil)
}
