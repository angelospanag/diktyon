package graph

// NodeType distinguishes the kind of entity a node represents.
type NodeType string

const (
	NodeTypeCompany NodeType = "company"
	NodeTypeOfficer NodeType = "officer"
	NodeTypePSC     NodeType = "psc" // non-UK or unregistered corporate PSC / individual PSC
)

// EdgeKind describes the relationship between two nodes.
type EdgeKind string

const (
	EdgeKindOfficerOf EdgeKind = "officer_of"
	EdgeKindPSCOf     EdgeKind = "psc_of"
)

// NodeMeta carries type-specific metadata and visual hints for the Polaroid card renderer.
type NodeMeta struct {
	// company
	CompanyNumber     string   `json:"company_number,omitempty"     doc:"Companies House number (UK) or GEMI number (GR) — company nodes only" example:"12345678"`
	Status            string   `json:"status,omitempty"             doc:"Company status, e.g. active, dissolved — company nodes only"          example:"active"`
	CompanyType       string   `json:"company_type,omitempty"       doc:"Legal form, e.g. ltd, plc — company nodes only"                       example:"ltd"`
	IncorporationDate string   `json:"incorporation_date,omitempty" doc:"Date of incorporation, ISO 8601 — company nodes only"                 example:"2015-03-17"`
	Address           string   `json:"address,omitempty"            doc:"Registered office address snippet — company nodes only"               example:"1 Example Street, London, EC1A 1BB"`
	SICCodes          []string `json:"sic_codes,omitempty"          doc:"SIC classification codes — company nodes only"                        example:"62020,64209"`

	// Distress flags — straight from the company profile we already fetch for
	// the root node, no extra API calls. Only populated for nodes that came
	// from a direct profile fetch (i.e. the root company), not for company
	// nodes discovered via officer/PSC fan-out.
	AccountsOverdue              bool `json:"accounts_overdue,omitempty"               doc:"True if statutory accounts are overdue — root company node only"        example:"false"`
	ConfirmationStatementOverdue bool `json:"confirmation_statement_overdue,omitempty" doc:"True if the confirmation statement is overdue — root company node only" example:"false"`
	HasInsolvencyHistory         bool `json:"has_insolvency_history,omitempty"         doc:"True if the company has insolvency history — root company node only"    example:"false"`
	HasCharges                   bool `json:"has_charges,omitempty"                    doc:"True if the company has registered charges — root company node only"    example:"true"`

	// officer
	OfficerID   string `json:"officer_id,omitempty"   doc:"Officer identifier — officer nodes only"                      example:"AbC1dEfGhIjKlMnOpQrStUvW"`
	Role        string `json:"role,omitempty"         doc:"Officer role, e.g. director, secretary — officer nodes only"  example:"director"`
	AppointedOn string `json:"appointed_on,omitempty" doc:"Appointment start date, ISO 8601 — officer and PSC nodes"     example:"2015-03-17"`
	ResignedOn  string `json:"resigned_on,omitempty"  doc:"Resignation date, ISO 8601, if resigned — officer nodes only" example:"2020-06-01"`
	Nationality string `json:"nationality,omitempty"  doc:"Officer nationality — officer nodes only"                     example:"British"`
	Occupation  string `json:"occupation,omitempty"   doc:"Officer occupation — officer nodes only"                      example:"Company Director"`

	// psc
	NaturesOfControl []string `json:"natures_of_control,omitempty" doc:"Nature(s) of control held — PSC nodes only"                example:"ownership-of-shares-25-to-50-percent"`
	NotifiedOn       string   `json:"notified_on,omitempty"        doc:"Date the PSC was notified, ISO 8601 — PSC nodes only"      example:"2016-04-06"`
	CeasedOn         string   `json:"ceased_on,omitempty"          doc:"Date the PSC ceased, ISO 8601, if ceased — PSC nodes only" example:"2021-01-01"`

	// RegistryURL is an optional deep-link to the entity on its source registry site.
	// Populated by providers that have a stable public URL (e.g. GEMI company page).
	RegistryURL string `json:"registry_url,omitempty" doc:"Deep-link to the entity on its source registry site, where available" example:"https://find-and-update.company-information.service.gov.uk/company/12345678"`

	Expanded bool `json:"expanded" doc:"True if this node's connections have already been fetched and rendered" example:"false"`
}

// Node is a single entity in the graph.
type Node struct {
	ID    string   `json:"id"    doc:"Stable node identifier, e.g. company:{number}, officer:{id}, psc:{name_slug}" example:"company:12345678"`
	Label string   `json:"label" doc:"Display name for the node"                                                    example:"EXAMPLE LTD"`
	Type  NodeType `json:"type"  doc:"Node type: company, officer, or psc"                                          example:"company"`
	Meta  NodeMeta `json:"meta"  doc:"Type-specific metadata and visual hints for the Polaroid card renderer"`
}

// Edge is a directed relationship between two nodes.
type Edge struct {
	Source string   `json:"source" doc:"Source node ID"                          example:"company:12345678"`
	Target string   `json:"target" doc:"Target node ID"                          example:"officer:AbC1dEfGhIjKlMnOpQrStUvW"`
	Kind   EdgeKind `json:"kind"   doc:"Relationship type: officer_of or psc_of" example:"officer_of"`
}

// Response is the Cytoscape-ready graph payload returned by the API.
// Each element carries enough metadata for Polaroid-card rendering without a
// second fetch. The frontend wraps nodes/edges in {data: ...} before passing
// to cy.add() — one line of adapter code.
type Response struct {
	Nodes []Node `json:"nodes" doc:"Graph nodes: companies, officers, and PSCs"`
	Edges []Edge `json:"edges" doc:"Directed relationships between nodes"`
}

// Collect converts dedup maps into a flat node/edge list.
func Collect(nodes map[string]Node, edges map[string]Edge) *Response {
	resp := &Response{
		Nodes: make([]Node, 0, len(nodes)),
		Edges: make([]Edge, 0, len(edges)),
	}
	for _, n := range nodes {
		resp.Nodes = append(resp.Nodes, n)
	}
	for _, e := range edges {
		resp.Edges = append(resp.Edges, e)
	}
	return resp
}
