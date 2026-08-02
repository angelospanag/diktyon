package uk

import (
	"testing"

	"github.com/angelospanag/diktyon/internal/graph"
)

func TestExtractOfficerID(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"/officers/abc123/appointments", "abc123"},
		{"/officers/abc-456xyz/appointments", "abc-456xyz"},
		{"", ""},
		{"/company/00000001/officers", ""},
	}
	for _, c := range cases {
		if got := extractOfficerID(c.input); got != c.want {
			t.Errorf("extractOfficerID(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestFormatName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"SMITH, John", "John Smith"},
		{"DOE, Jane Elizabeth", "Jane Elizabeth Doe"},
		{"O'BRIEN, Patrick", "Patrick O'brien"},
		{"SingleName", "SingleName"},
	}
	for _, c := range cases {
		if got := formatName(c.input); got != c.want {
			t.Errorf("formatName(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestSanitizeID(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Big Corp Ltd", "big_corp_ltd"},
		{"Alpha-Beta", "alpha-beta"},
		{"Café & Co", "caf__co"},
		{"UPPER CASE", "upper_case"},
	}
	for _, c := range cases {
		if got := sanitizeID(c.input); got != c.want {
			t.Errorf("sanitizeID(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestPSCNodeID(t *testing.T) {
	t.Run("UK corporate PSC merges into company namespace", func(t *testing.T) {
		p := &PSC{
			Name: "Big Corp Ltd",
			Kind: "corporate-entity-person-with-significant-control",
			Identification: &PSCIdentification{
				LegalAuthority:     "Companies Act 2006",
				RegistrationNumber: "99999999",
			},
		}
		id, typ := pscNodeID(p)
		if id != "company:99999999" {
			t.Errorf("id = %q, want company:99999999", id)
		}
		if typ != graph.NodeTypeCompany {
			t.Errorf("type = %q, want company", typ)
		}
	})

	t.Run("foreign corporate PSC uses psc namespace", func(t *testing.T) {
		p := &PSC{
			Name: "Foreign GmbH",
			Kind: "corporate-entity-person-with-significant-control",
			Identification: &PSCIdentification{
				CountryRegistered:  "Germany",
				RegistrationNumber: "HRB12345",
			},
		}
		id, typ := pscNodeID(p)
		if id != "psc:foreign_gmbh" {
			t.Errorf("id = %q, want psc:foreign_gmbh", id)
		}
		if typ != graph.NodeTypePSC {
			t.Errorf("type = %q, want psc", typ)
		}
	})

	t.Run("individual PSC uses psc namespace", func(t *testing.T) {
		p := &PSC{
			Name: "Jane Doe",
			Kind: "individual-person-with-significant-control",
		}
		id, typ := pscNodeID(p)
		if id != "psc:jane_doe" {
			t.Errorf("id = %q, want psc:jane_doe", id)
		}
		if typ != graph.NodeTypePSC {
			t.Errorf("type = %q, want psc", typ)
		}
	})
}
