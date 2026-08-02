package gr

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/angelospanag/diktyon/internal/graph"
	"github.com/angelospanag/diktyon/internal/registry"
)

const registryBaseURL = "https://publicity.businessportal.gr/company/"

// Provider implements registry.Provider using the GEMI Open Data API.
// Officer-centric expansion (OfficerGraph) is not supported — GEMI has no
// person-by-ID lookup equivalent to Companies House appointments.
type Provider struct {
	client *Client
}

func New(client *Client) *Provider {
	return &Provider{client: client}
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
	for _, c := range items {
		out = append(out, registry.SearchResult{
			ID:             strconv.Itoa(c.ArGemi),
			Name:           c.CoNameEl,
			AddressSnippet: addressSnippet(&c),
			Status:         c.Status.Descr,
			CompanyType:    c.LegalType.Descr,
			DateOfCreation: c.IncorporationDate,
		})
	}
	return out, nil
}

func (p *Provider) CompanyGraph(
	ctx context.Context,
	companyID string,
	_ int,
) (*graph.Response, error) {
	company, err := p.client.GetCompany(ctx, companyID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, registry.ErrNotFound
		}
		return nil, err
	}

	nodes := make(map[string]graph.Node)
	edges := make(map[string]graph.Edge)

	arGemiStr := strconv.Itoa(company.ArGemi)
	rootID := "company:gr:" + arGemiStr

	// Collect KAD codes for the root company node.
	var kadCodes []string
	for _, a := range company.Activities {
		if a.Activity.ID != "" {
			kadCodes = append(kadCodes, a.Activity.ID)
		}
	}

	nodes[rootID] = graph.Node{
		ID:    rootID,
		Label: company.CoNameEl,
		Type:  graph.NodeTypeCompany,
		Meta: graph.NodeMeta{
			CompanyNumber:     arGemiStr,
			Status:            company.Status.Descr,
			CompanyType:       company.LegalType.Descr,
			IncorporationDate: company.IncorporationDate,
			Address:           addressSnippet(company),
			SICCodes:          kadCodes,
			RegistryURL:       registryBaseURL + arGemiStr,
		},
	}

	for _, person := range company.Persons {
		label, nodeID, nodeType := personNode(&person)
		if label == "" {
			continue
		}

		if _, exists := nodes[nodeID]; !exists {
			meta := graph.NodeMeta{
				Role:        person.Role,
				AppointedOn: person.DtFrom,
				ResignedOn:  person.DtTo,
			}
			// Persons with a percentage share are treated as PSC-equivalent.
			if person.Percentage != "" {
				meta.NaturesOfControl = []string{fmt.Sprintf("%s%%", person.Percentage)}
			}
			nodes[nodeID] = graph.Node{
				ID:    nodeID,
				Label: label,
				Type:  nodeType,
				Meta:  meta,
			}
		}

		edgeKind := graph.EdgeKindOfficerOf
		if nodeType == graph.NodeTypePSC {
			edgeKind = graph.EdgeKindPSCOf
		}
		edgeKey := nodeID + "→" + rootID + ":" + string(edgeKind)
		edges[edgeKey] = graph.Edge{Source: nodeID, Target: rootID, Kind: edgeKind}
	}

	return graph.Collect(nodes, edges), nil
}

func (p *Provider) OfficerGraph(_ context.Context, _ string) (*graph.Response, error) {
	return nil, registry.ErrUnsupported
}

// personNode derives the label, stable node ID, and node type for a CompanyPerson.
// Natural persons → officer; legal entities (corporate) with a percentage → psc; without → officer.
func personNode(cp *CompanyPerson) (label, nodeID string, nodeType graph.NodeType) {
	if cp.PersonName != "" {
		label = cp.PersonName
		nodeID = "officer:gr:" + sanitizeID(label)
		nodeType = graph.NodeTypeOfficer
		return
	}
	if cp.BusinessName != "" {
		label = cp.BusinessName
		if cp.Percentage != "" {
			nodeID = "psc:gr:" + sanitizeID(label)
			nodeType = graph.NodeTypePSC
		} else {
			nodeID = "officer:gr:" + sanitizeID(label)
			nodeType = graph.NodeTypeOfficer
		}
		return
	}
	return "", "", ""
}

func addressSnippet(c *Company) string {
	var parts []string
	if c.Street != "" {
		street := c.Street
		if c.StreetNumber != "" {
			street += " " + c.StreetNumber
		}
		parts = append(parts, street)
	}
	if c.City != "" {
		city := c.City
		if c.ZipCode != "" {
			city += " " + c.ZipCode
		}
		parts = append(parts, city)
	}
	return strings.Join(parts, ", ")
}

func sanitizeID(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' {
			return '_'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r + 32 // to lower
		}
		return -1
	}, s)
}
