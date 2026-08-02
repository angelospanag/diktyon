package gr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://opendata-api.businessportal.gr/api/opendata/v1"

// ErrNotFound is returned when GEMI returns 404.
var ErrNotFound = errors.New("not found")

type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// WithBaseURL overrides the API base URL. Intended for tests only.
func (c *Client) WithBaseURL(u string) *Client {
	c.baseURL = u
	return c
}

// SearchCompanies searches companies by name. Returns up to 20 results.
func (c *Client) SearchCompanies(ctx context.Context, name string) ([]Company, error) {
	path := "/companies?name=" + url.QueryEscape(name) + "&resultsSize=20&resultsOffset=0"
	var resp SearchResponse
	if err := c.fetch(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.SearchResults, nil
}

// GetCompany fetches the full company record for the given GEMI number.
func (c *Client) GetCompany(ctx context.Context, arGemi string) (*Company, error) {
	path := "/companies/" + url.PathEscape(arGemi)
	var company Company
	if err := c.fetch(ctx, path, &company); err != nil {
		return nil, err
	}
	return &company, nil
}

func (c *Client) fetch(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("api_key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("GET %s: %w", path, ErrNotFound)
	case http.StatusOK:
		return json.NewDecoder(resp.Body).Decode(v)
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("GET %s: status %d: %s", path, resp.StatusCode, body)
	}
}
