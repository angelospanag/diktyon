package uk

import "strings"

// SearchResponse is returned by GET /search/companies
type SearchResponse struct {
	Items        []SearchItem `json:"items"`
	TotalResults int          `json:"total_results"`
}

type SearchItem struct {
	CompanyNumber  string `json:"company_number"`
	Title          string `json:"title"`
	CompanyStatus  string `json:"company_status"`
	CompanyType    string `json:"company_type"`
	AddressSnippet string `json:"address_snippet"`
	DateOfCreation string `json:"date_of_creation"`
}

// CompanyProfile is returned by GET /company/{company_number}
type CompanyProfile struct {
	CompanyNumber           string                 `json:"company_number"`
	CompanyName             string                 `json:"company_name"`
	CompanyStatus           string                 `json:"company_status"`
	CompanyType             string                 `json:"company_type"`
	DateOfCreation          string                 `json:"date_of_creation"`
	RegisteredOfficeAddress Address                `json:"registered_office_address"`
	SICCodes                []string               `json:"sic_codes"`
	Accounts                *Accounts              `json:"accounts,omitempty"`
	ConfirmationStatement   *ConfirmationStatement `json:"confirmation_statement,omitempty"`
	HasInsolvencyHistory    bool                   `json:"has_insolvency_history,omitempty"`
	HasCharges              bool                   `json:"has_charges,omitempty"`
	Links                   CompanyLinks           `json:"links"`
}

// Accounts carries just the "is something overdue" signal — investigators
// care about the flag, not the full filing schedule.
type Accounts struct {
	Overdue bool `json:"overdue"`
}

type ConfirmationStatement struct {
	Overdue bool `json:"overdue"`
}

type CompanyLinks struct {
	Officers                      string `json:"officers"`
	PersonsWithSignificantControl string `json:"persons_with_significant_control"`
}

// OfficersResponse is returned by GET /company/{company_number}/officers
type OfficersResponse struct {
	Items         []Officer `json:"items"`
	TotalResults  int       `json:"total_results"`
	ActiveCount   int       `json:"active_count"`
	ResignedCount int       `json:"resigned_count"`
}

type Officer struct {
	Name        string       `json:"name"`
	OfficerRole string       `json:"officer_role"`
	AppointedOn string       `json:"appointed_on"`
	ResignedOn  string       `json:"resigned_on,omitempty"`
	Nationality string       `json:"nationality,omitempty"`
	Occupation  string       `json:"occupation,omitempty"`
	Address     Address      `json:"address"`
	DateOfBirth *DateOfBirth `json:"date_of_birth,omitempty"`
	Links       OfficerLinks `json:"links"`
}

type OfficerLinks struct {
	Officer OfficerLinkItem `json:"officer"`
}

type OfficerLinkItem struct {
	Appointments string `json:"appointments"`
}

// PSCResponse is returned by GET /company/{company_number}/persons-with-significant-control
type PSCResponse struct {
	Items        []PSC `json:"items"`
	TotalResults int   `json:"total_results"`
}

type PSC struct {
	Name                string             `json:"name"`
	Kind                string             `json:"kind"` // individual-person-with-significant-control | corporate-entity-person-with-significant-control | ...
	NaturesOfControl    []string           `json:"natures_of_control"`
	NotifiedOn          string             `json:"notified_on"`
	Ceased              bool               `json:"ceased,omitempty"`
	CeasedToBeEffective string             `json:"ceased_to_be_effective,omitempty"`
	Nationality         string             `json:"nationality,omitempty"`
	Address             Address            `json:"address"`
	Identification      *PSCIdentification `json:"identification,omitempty"`
	DateOfBirth         *DateOfBirth       `json:"date_of_birth,omitempty"`
}

func (p *PSC) IsCorporate() bool {
	return strings.Contains(p.Kind, "corporate-entity")
}

func (p *PSC) UKRegistrationNumber() string {
	if p.Identification == nil || p.Identification.RegistrationNumber == "" {
		return ""
	}
	id := p.Identification
	// "Companies Act" in the legal authority is the most reliable UK signal.
	if strings.Contains(strings.ToLower(id.LegalAuthority), "companies act") {
		return id.RegistrationNumber
	}
	c := id.CountryRegistered
	switch {
	case c == "",
		strings.EqualFold(c, "United Kingdom"),
		strings.EqualFold(c, "England and Wales"),
		strings.EqualFold(c, "England"),
		strings.EqualFold(c, "Wales"),
		strings.EqualFold(c, "Scotland"),
		strings.EqualFold(c, "Northern Ireland"):
		return id.RegistrationNumber
	}
	return ""
}

type PSCIdentification struct {
	LegalAuthority     string `json:"legal_authority"`
	LegalForm          string `json:"legal_form"`
	RegistrationNumber string `json:"registration_number,omitempty"`
	CountryRegistered  string `json:"country_registered,omitempty"`
}

// AppointmentsResponse is returned by GET /officers/{officer_id}/appointments
type AppointmentsResponse struct {
	Items        []Appointment `json:"items"`
	TotalResults int           `json:"total_results"`
	Name         string        `json:"name"`
}

type Appointment struct {
	AppointedTo AppointedTo `json:"appointed_to"`
	Name        string      `json:"name"`
	Role        string      `json:"role"`
	AppointedOn string      `json:"appointed_on"`
	ResignedOn  string      `json:"resigned_on,omitempty"`
}

type AppointedTo struct {
	CompanyNumber string `json:"company_number"`
	CompanyName   string `json:"company_name"`
	CompanyStatus string `json:"company_status"`
}

type Address struct {
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	Locality     string `json:"locality,omitempty"`
	Region       string `json:"region,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	Country      string `json:"country,omitempty"`
}

func (a Address) Snippet() string {
	var parts []string
	if a.AddressLine1 != "" {
		parts = append(parts, a.AddressLine1)
	}
	if a.Locality != "" {
		parts = append(parts, a.Locality)
	}
	if a.PostalCode != "" {
		parts = append(parts, a.PostalCode)
	}
	return strings.Join(parts, ", ")
}

type DateOfBirth struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}
