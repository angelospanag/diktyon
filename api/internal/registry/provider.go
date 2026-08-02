package registry

import (
	"context"
	"errors"

	"github.com/angelospanag/diktyon/internal/graph"
)

// ErrNotFound is returned by a Provider when the requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrUnsupported is returned when a provider does not implement a given operation.
var ErrUnsupported = errors.New("operation not supported by this provider")

// SearchResult is a single item returned by Provider.SearchCompanies.
type SearchResult struct {
	ID             string
	Name           string
	AddressSnippet string
	Status         string
	CompanyType    string
	DateOfCreation string
}

// Provider is the abstraction each jurisdiction must implement.
// All methods return canonical graph types so the API layer stays country-agnostic.
type Provider interface {
	SearchCompanies(ctx context.Context, query string) ([]SearchResult, error)
	CompanyGraph(ctx context.Context, companyID string, depth int) (*graph.Response, error)
	// OfficerGraph returns the graph centred on an officer.
	// Providers that do not support officer-centric expansion should return ErrUnsupported.
	OfficerGraph(ctx context.Context, officerID string) (*graph.Response, error)
}
