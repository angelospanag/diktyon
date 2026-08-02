package uk

import "testing"

func TestAddressSnippet(t *testing.T) {
	cases := []struct {
		addr Address
		want string
	}{
		{
			Address{AddressLine1: "1 Test St", Locality: "London", PostalCode: "EC1A 1AA"},
			"1 Test St, London, EC1A 1AA",
		},
		{Address{AddressLine1: "2 High St"}, "2 High St"},
		{Address{}, ""},
	}
	for _, c := range cases {
		if got := c.addr.Snippet(); got != c.want {
			t.Errorf("Snippet() = %q, want %q", got, c.want)
		}
	}
}

func TestPSCIsCorporate(t *testing.T) {
	if !(&PSC{Kind: "corporate-entity-person-with-significant-control"}).IsCorporate() {
		t.Error("expected corporate PSC to be corporate")
	}
	if (&PSC{Kind: "individual-person-with-significant-control"}).IsCorporate() {
		t.Error("expected individual PSC not to be corporate")
	}
}

func TestPSCUKRegistrationNumber(t *testing.T) {
	cases := []struct {
		name string
		psc  PSC
		want string
	}{
		{
			name: "companies act authority",
			psc: PSC{Identification: &PSCIdentification{
				LegalAuthority:     "Companies Act 2006",
				RegistrationNumber: "12345678",
			}},
			want: "12345678",
		},
		{
			name: "UK country registered",
			psc: PSC{Identification: &PSCIdentification{
				CountryRegistered:  "England and Wales",
				RegistrationNumber: "87654321",
			}},
			want: "87654321",
		},
		{
			name: "foreign registration",
			psc: PSC{Identification: &PSCIdentification{
				CountryRegistered:  "Germany",
				RegistrationNumber: "HRB12345",
			}},
			want: "",
		},
		{
			name: "no identification",
			psc:  PSC{},
			want: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.psc.UKRegistrationNumber(); got != c.want {
				t.Errorf("UKRegistrationNumber() = %q, want %q", got, c.want)
			}
		})
	}
}
