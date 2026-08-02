package uk

import (
	"context"
	"errors"

	"github.com/angelospanag/diktyon/internal/graph"
	"github.com/angelospanag/diktyon/internal/registry"
)

// Provider implements registry.Provider using the Companies House public API.
type Provider struct {
	client  *Client
	builder *Builder
}

func New(client *Client) *Provider {
	return &Provider{
		client:  client,
		builder: NewBuilder(client),
	}
}

func (p *Provider) SearchCompanies(
	ctx context.Context,
	query string,
) ([]registry.SearchResult, error) {
	items, err := p.client.SearchCompanies(ctx, query)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return []registry.SearchResult{}, nil
		}
		return nil, err
	}
	out := make([]registry.SearchResult, 0, len(items))
	for _, item := range items {
		out = append(out, registry.SearchResult{
			ID:             item.CompanyNumber,
			Name:           item.Title,
			AddressSnippet: item.AddressSnippet,
			Status:         item.CompanyStatus,
			CompanyType:    item.CompanyType,
			DateOfCreation: item.DateOfCreation,
		})
	}
	return out, nil
}

func (p *Provider) CompanyGraph(
	ctx context.Context,
	companyID string,
	depth int,
) (*graph.Response, error) {
	result, err := p.builder.ForCompany(ctx, companyID, depth)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, registry.ErrNotFound
		}
		return nil, err
	}
	return result, nil
}

func (p *Provider) OfficerGraph(ctx context.Context, officerID string) (*graph.Response, error) {
	result, err := p.builder.ForOfficer(ctx, officerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, registry.ErrNotFound
		}
		return nil, err
	}
	return result, nil
}
