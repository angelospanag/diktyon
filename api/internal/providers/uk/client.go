package uk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

const defaultBaseURL = "https://api.company-information.service.gov.uk"

// ErrNotFound is returned when Companies House returns 404.
var ErrNotFound = errors.New("not found")

type Client struct {
	apiKey     string
	baseURL    string
	http       *http.Client
	limiter    *rate.Limiter
	maxRetries int
}

func NewClient(apiKey string, rps float64, burst int) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		limiter:    rate.NewLimiter(rate.Limit(rps), burst),
		maxRetries: 3,
	}
}

// WithBaseURL overrides the API base URL. Intended for tests only.
func (c *Client) WithBaseURL(u string) *Client {
	c.baseURL = u
	return c
}

func (c *Client) SearchCompanies(ctx context.Context, query string) ([]SearchItem, error) {
	path := "/search/companies?q=" + url.QueryEscape(query) + "&items_per_page=20"
	var resp SearchResponse
	if err := c.fetch(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetCompany(ctx context.Context, companyNumber string) (*CompanyProfile, error) {
	path := "/company/" + url.PathEscape(companyNumber)
	var resp CompanyProfile
	if err := c.fetch(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetOfficers(ctx context.Context, companyNumber string) ([]Officer, error) {
	path := "/company/" + url.PathEscape(
		companyNumber,
	) + "/officers?items_per_page=100&register_view=false"
	var resp OfficersResponse
	if err := c.fetch(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetPSCs(ctx context.Context, companyNumber string) ([]PSC, error) {
	path := "/company/" + url.PathEscape(
		companyNumber,
	) + "/persons-with-significant-control?items_per_page=100"
	var resp PSCResponse
	if err := c.fetch(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetOfficerAppointments(
	ctx context.Context,
	officerID string,
) ([]Appointment, error) {
	path := "/officers/" + url.PathEscape(officerID) + "/appointments?items_per_page=100"
	var resp AppointmentsResponse
	if err := c.fetch(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) fetch(ctx context.Context, path string, v any) error {
	u := c.baseURL + path
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryBackoff(attempt)):
			}
		}

		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(c.apiKey, "")
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("GET %s: %w", path, err)
			continue
		}

		switch {
		case resp.StatusCode == http.StatusNotFound:
			_ = resp.Body.Close()
			return fmt.Errorf("GET %s: %w", path, ErrNotFound)

		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("GET %s: status %d", path, resp.StatusCode)
			continue

		case resp.StatusCode != http.StatusOK:
			_ = resp.Body.Close()
			return fmt.Errorf("GET %s: unexpected status %d", path, resp.StatusCode)
		}

		err = json.NewDecoder(resp.Body).Decode(v)
		_ = resp.Body.Close()
		return err
	}

	return fmt.Errorf("GET %s after %d retries: %w", path, c.maxRetries, lastErr)
}

func retryBackoff(attempt int) time.Duration {
	base := 500 * time.Millisecond
	// Exponential: 500ms, 1s, 2s
	d := min(base*(1<<uint(attempt-1)), 30*time.Second)
	// ±20% jitter
	jitter := time.Duration(rand.Int64N(int64(d / 5)))
	return d + jitter
}
