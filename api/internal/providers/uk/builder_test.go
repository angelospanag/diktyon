package uk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/angelospanag/diktyon/internal/graph"
)

// serve registers a fixed JSON payload at the given pattern on mux.
func serve(mux *http.ServeMux, pattern string, v any) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	})
}

func testServer() (*httptest.Server, *Client) {
	mux := http.NewServeMux()

	serve(mux, "/company/00000001", CompanyProfile{
		CompanyNumber: "00000001",
		CompanyName:   "Test Company Ltd",
		CompanyStatus: "active",
		CompanyType:   "ltd",
		RegisteredOfficeAddress: Address{
			AddressLine1: "1 Test Street",
			Locality:     "London",
			PostalCode:   "EC1A 1AA",
		},
	})
	serve(mux, "/company/00000001/officers", OfficersResponse{
		Items: []Officer{
			{
				Name:        "SMITH, John",
				OfficerRole: "director",
				AppointedOn: "2020-01-01",
				Links: OfficerLinks{
					Officer: OfficerLinkItem{
						Appointments: "/officers/abc123/appointments",
					},
				},
			},
		},
	})
	serve(mux, "/company/00000001/persons-with-significant-control", PSCResponse{
		Items: []PSC{
			{
				Name:             "Big Corp Ltd",
				Kind:             "corporate-entity-person-with-significant-control",
				NaturesOfControl: []string{"ownership-of-shares-75-to-100-percent"},
				NotifiedOn:       "2020-01-01",
				Identification: &PSCIdentification{
					LegalAuthority:     "Companies Act 2006",
					RegistrationNumber: "99999999",
				},
			},
		},
	})
	serve(mux, "/officers/abc123/appointments", AppointmentsResponse{
		Name: "SMITH, John",
		Items: []Appointment{
			{
				Name: "SMITH, John",
				AppointedTo: AppointedTo{
					CompanyNumber: "00000001",
					CompanyName:   "Test Company Ltd",
					CompanyStatus: "active",
				},
				AppointedOn: "2020-01-01",
			},
			{
				Name: "SMITH, John",
				AppointedTo: AppointedTo{
					CompanyNumber: "00000002",
					CompanyName:   "Another Company Ltd",
					CompanyStatus: "active",
				},
				AppointedOn: "2021-06-01",
			},
		},
	})

	srv := httptest.NewServer(mux)
	client := NewClient("test", 1000, 100).WithBaseURL(srv.URL)
	return srv, client
}

func nodeByID(nodes []graph.Node, id string) *graph.Node {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
	}
	return nil
}

func hasEdge(edges []graph.Edge, src, tgt string, kind graph.EdgeKind) bool {
	for _, e := range edges {
		if e.Source == src && e.Target == tgt && e.Kind == kind {
			return true
		}
	}
	return false
}

func TestForCompanyDepth1(t *testing.T) {
	srv, client := testServer()
	defer srv.Close()

	b := NewBuilder(client)
	resp, err := b.ForCompany(context.Background(), "00000001", 1)
	if err != nil {
		t.Fatalf("ForCompany: %v", err)
	}

	// Company node
	company := nodeByID(resp.Nodes, "company:00000001")
	if company == nil {
		t.Fatal("company:00000001 node missing")
	}
	if company.Label != "Test Company Ltd" {
		t.Errorf("company label = %q, want Test Company Ltd", company.Label)
	}
	if company.Meta.Status != "active" {
		t.Errorf("company status = %q, want active", company.Meta.Status)
	}

	// Officer node
	officer := nodeByID(resp.Nodes, "officer:abc123")
	if officer == nil {
		t.Fatal("officer:abc123 node missing")
	}
	if officer.Label != "John Smith" {
		t.Errorf("officer label = %q, want John Smith", officer.Label)
	}

	// UK corporate PSC merges into company namespace
	psc := nodeByID(resp.Nodes, "company:99999999")
	if psc == nil {
		t.Fatal("company:99999999 (PSC) node missing")
	}

	// Edges
	if !hasEdge(resp.Edges, "officer:abc123", "company:00000001", graph.EdgeKindOfficerOf) {
		t.Error("missing officer_of edge")
	}
	if !hasEdge(resp.Edges, "company:99999999", "company:00000001", graph.EdgeKindPSCOf) {
		t.Error("missing psc_of edge")
	}
}

func TestForCompanyNoDuplicateNodes(t *testing.T) {
	srv, client := testServer()
	defer srv.Close()

	b := NewBuilder(client)
	resp, err := b.ForCompany(context.Background(), "00000001", 1)
	if err != nil {
		t.Fatalf("ForCompany: %v", err)
	}

	seen := make(map[string]int)
	for _, n := range resp.Nodes {
		seen[n.ID]++
	}
	for id, count := range seen {
		if count > 1 {
			t.Errorf("node %q appears %d times", id, count)
		}
	}
}

func TestForOfficer(t *testing.T) {
	srv, client := testServer()
	defer srv.Close()

	b := NewBuilder(client)
	resp, err := b.ForOfficer(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("ForOfficer: %v", err)
	}

	officer := nodeByID(resp.Nodes, "officer:abc123")
	if officer == nil {
		t.Fatal("officer:abc123 node missing")
	}
	if officer.Label != "John Smith" {
		t.Errorf("officer label = %q, want John Smith", officer.Label)
	}

	if nodeByID(resp.Nodes, "company:00000001") == nil {
		t.Error("company:00000001 node missing")
	}
	if nodeByID(resp.Nodes, "company:00000002") == nil {
		t.Error("company:00000002 node missing")
	}

	if !hasEdge(resp.Edges, "officer:abc123", "company:00000001", graph.EdgeKindOfficerOf) {
		t.Error("missing edge to company:00000001")
	}
	if !hasEdge(resp.Edges, "officer:abc123", "company:00000002", graph.EdgeKindOfficerOf) {
		t.Error("missing edge to company:00000002")
	}
}
