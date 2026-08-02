package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/angelospanag/diktyon/internal/api/middleware"
	"github.com/angelospanag/diktyon/internal/cache"
	"github.com/angelospanag/diktyon/internal/graph"
	"github.com/angelospanag/diktyon/internal/registry"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type SearchInput struct {
	Q       string `query:"q"       minLength:"2" maxLength:"100" doc:"Company name or number to search for" example:"acme"`
	Country string `query:"country"               maxLength:"10"  doc:"Country code: uk (default), gr"       example:"uk"`
}

type SearchResult struct {
	CompanyNumber  string `json:"company_number"             doc:"Company identifier (CH number for UK, GEMI number for GR)" example:"12345678"`
	Name           string `json:"name"                       doc:"Registered company name"                                   example:"EXAMPLE LTD"`
	AddressSnippet string `json:"address_snippet"            doc:"Short registered address snippet"                          example:"1 Example Street, London"`
	Status         string `json:"status"                     doc:"Company status, e.g. active, dissolved"                    example:"active"`
	Type           string `json:"type"                       doc:"Legal form, e.g. ltd, plc"                                 example:"ltd"`
	DateOfCreation string `json:"date_of_creation,omitempty" doc:"Date of incorporation, ISO 8601"                           example:"2015-03-17"`
}

type SearchOutput struct {
	Body []SearchResult
}

type CompanyGraphInput struct {
	CompanyNumber string `path:"company_number" doc:"Company identifier (CH number for UK, GEMI number for GR)"                           example:"12345678"`
	Depth         int    `                      doc:"Graph expansion depth: 1=direct connections (default), 2=connections of connections" example:"1"        query:"depth"`
	Country       string `                      doc:"Country code: uk (default), gr"                                                      example:"uk"       query:"country" maxLength:"10"`
}

type CompanyGraphOutput struct {
	Body *graph.Response
}

type OfficerGraphInput struct {
	OfficerID string `path:"officer_id" doc:"Officer identifier"             example:"AbC1dEfGhIjKlMnOpQrStUvW"`
	Country   string `                  doc:"Country code: uk (default), gr" example:"uk"                       query:"country" maxLength:"10"`
}

type OfficerGraphOutput struct {
	Body *graph.Response
}

// NewRouter wires up all routes and returns the root http.Handler and the Huma API
// (for OpenAPI schema generation — see the openapi subcommand in cmd/api/main.go).
func NewRouter(reg *registry.Registry, c cache.Cache) (http.Handler, huma.API) {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// Liveness probe — outside Huma so it never appears in the OpenAPI spec.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	cfg := huma.DefaultConfig("Diktyon API", "1.0.0")
	cfg.Info.Description = "Multi-jurisdiction company investigation API"

	api := humachi.New(r, cfg)

	huma.Register(api, huma.Operation{
		OperationID: "search-companies",
		Method:      http.MethodGet,
		Path:        "/api/search",
		Summary:     "Search companies by name",
		Description: "Searches the requested country's company registry by name or number and returns matching companies for display as search result cards.",
		Tags:        []string{"search"},
	}, func(ctx context.Context, input *SearchInput) (*SearchOutput, error) {
		country := countryOrDefault(input.Country)

		provider, err := reg.Get(country)
		if err != nil {
			return nil, huma.Error404NotFound("unknown country: " + country)
		}

		cacheKey := fmt.Sprintf("search:v1:%s:%s", country, input.Q)

		var cached []SearchResult
		if err := c.Get(ctx, cacheKey, &cached); err == nil {
			return &SearchOutput{Body: cached}, nil
		}

		items, err := provider.SearchCompanies(ctx, input.Q)
		if err != nil {
			slog.ErrorContext(ctx, "upstream search failed", "country", country, "error", err)
			return nil, huma.Error502BadGateway("upstream request failed")
		}

		results := make([]SearchResult, 0, len(items))
		for _, item := range items {
			results = append(results, SearchResult{
				CompanyNumber:  item.ID,
				Name:           item.Name,
				AddressSnippet: item.AddressSnippet,
				Status:         item.Status,
				Type:           item.CompanyType,
				DateOfCreation: item.DateOfCreation,
			})
		}

		// Search results are time-sensitive; cache for 1 hour rather than the default 24h.
		_ = c.Set(ctx, cacheKey, results, time.Hour)

		return &SearchOutput{Body: results}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-company-graph",
		Method:      http.MethodGet,
		Path:        "/api/company/{company_number}/graph",
		Summary:     "Corporate network graph for a company",
		Description: "Returns the company plus all its officers and PSCs (active and resigned/ceased) as a node/edge list. Add ?depth=2 to also expand each active officer's other current appointments.",
		Tags:        []string{"graph"},
	}, func(ctx context.Context, input *CompanyGraphInput) (*CompanyGraphOutput, error) {
		country := countryOrDefault(input.Country)

		provider, err := reg.Get(country)
		if err != nil {
			return nil, huma.Error404NotFound("unknown country: " + country)
		}

		depth := input.Depth
		if depth < 1 || depth > 2 {
			depth = 1
		}

		cacheKey := fmt.Sprintf(
			"graph:v5:%s:company:%s:depth:%d",
			country,
			input.CompanyNumber,
			depth,
		)

		var resp graph.Response
		if err := c.Get(ctx, cacheKey, &resp); err == nil {
			return &CompanyGraphOutput{Body: &resp}, nil
		}

		result, err := provider.CompanyGraph(ctx, input.CompanyNumber, depth)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				return nil, huma.Error404NotFound("company not found")
			}
			slog.ErrorContext(
				ctx,
				"upstream company graph failed",
				"country",
				country,
				"company",
				input.CompanyNumber,
				"error",
				err,
			)
			return nil, huma.Error502BadGateway("upstream request failed")
		}

		_ = c.Set(ctx, cacheKey, result, 0) // uses cache default TTL

		return &CompanyGraphOutput{Body: result}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-officer-graph",
		Method:      http.MethodGet,
		Path:        "/api/officer/{officer_id}/graph",
		Summary:     "Appointments graph for an officer",
		Description: "Returns the officer plus all companies they currently sit on the board of.",
		Tags:        []string{"graph"},
	}, func(ctx context.Context, input *OfficerGraphInput) (*OfficerGraphOutput, error) {
		country := countryOrDefault(input.Country)

		provider, err := reg.Get(country)
		if err != nil {
			return nil, huma.Error404NotFound("unknown country: " + country)
		}

		cacheKey := fmt.Sprintf("graph:v5:%s:officer:%s", country, input.OfficerID)

		var resp graph.Response
		if err := c.Get(ctx, cacheKey, &resp); err == nil {
			return &OfficerGraphOutput{Body: &resp}, nil
		}

		result, err := provider.OfficerGraph(ctx, input.OfficerID)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				return nil, huma.Error404NotFound("officer not found")
			}
			if errors.Is(err, registry.ErrUnsupported) {
				return nil, huma.Error422UnprocessableEntity(
					"officer graph not supported for this country",
				)
			}
			slog.ErrorContext(
				ctx,
				"upstream officer graph failed",
				"country",
				country,
				"officer",
				input.OfficerID,
				"error",
				err,
			)
			return nil, huma.Error502BadGateway("upstream request failed")
		}

		_ = c.Set(ctx, cacheKey, result, 0)

		return &OfficerGraphOutput{Body: result}, nil
	})

	return r, api
}

func countryOrDefault(country string) string {
	if country == "" {
		return "uk"
	}
	return country
}
