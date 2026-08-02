package uk

import (
	"context"
	"fmt"
	"strings"

	"github.com/angelospanag/diktyon/internal/graph"
	"golang.org/x/sync/errgroup"
)

type Builder struct {
	ch *Client
}

func NewBuilder(ch *Client) *Builder {
	return &Builder{ch: ch}
}

// ForCompany builds the graph centred on a Companies House company number.
// depth=1 returns the company + all its officers + all its PSCs (active and
// resigned/ceased — the frontend filters resigned/ceased nodes by default).
// depth=2 additionally expands each active officer's other current appointments (capped to avoid rate-limit spikes).
func (b *Builder) ForCompany(
	ctx context.Context,
	companyNumber string,
	depth int,
) (*graph.Response, error) {
	var (
		profile  *CompanyProfile
		officers []Officer
		pscs     []PSC
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		p, err := b.ch.GetCompany(gctx, companyNumber)
		if err != nil {
			return fmt.Errorf("company profile: %w", err)
		}
		profile = p
		return nil
	})

	g.Go(func() error {
		o, err := b.ch.GetOfficers(gctx, companyNumber)
		if err != nil {
			return fmt.Errorf("officers: %w", err)
		}
		officers = o
		return nil
	})

	g.Go(func() error {
		p, err := b.ch.GetPSCs(gctx, companyNumber)
		if err != nil {
			return fmt.Errorf("PSCs: %w", err)
		}
		pscs = p
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	nodes := make(map[string]graph.Node)
	edges := make(map[string]graph.Edge)

	// Root company node
	rootID := "company:" + profile.CompanyNumber
	nodes[rootID] = graph.Node{
		ID:    rootID,
		Label: profile.CompanyName,
		Type:  graph.NodeTypeCompany,
		Meta: graph.NodeMeta{
			CompanyNumber:     profile.CompanyNumber,
			Status:            profile.CompanyStatus,
			CompanyType:       profile.CompanyType,
			IncorporationDate: profile.DateOfCreation,
			Address:           profile.RegisteredOfficeAddress.Snippet(),
			SICCodes:          profile.SICCodes,
			AccountsOverdue:   profile.Accounts != nil && profile.Accounts.Overdue,
			ConfirmationStatementOverdue: profile.ConfirmationStatement != nil &&
				profile.ConfirmationStatement.Overdue,
			HasInsolvencyHistory: profile.HasInsolvencyHistory,
			HasCharges:           profile.HasCharges,
		},
	}

	// Officer nodes — both active and resigned are included; the frontend
	// decides whether to render resigned officers based on a user toggle.
	// Only active officers are eligible for depth-2 fan-out, below.
	var activeOfficerIDs []string
	for _, o := range officers {
		officerID := extractOfficerID(o.Links.Officer.Appointments)
		if officerID == "" {
			continue
		}
		nodeID := "officer:" + officerID
		nodes[nodeID] = graph.Node{
			ID:    nodeID,
			Label: formatName(o.Name),
			Type:  graph.NodeTypeOfficer,
			Meta: graph.NodeMeta{
				OfficerID:   officerID,
				Role:        o.OfficerRole,
				AppointedOn: o.AppointedOn,
				ResignedOn:  o.ResignedOn,
				Nationality: o.Nationality,
				Occupation:  o.Occupation,
			},
		}
		edgeKey := nodeID + "→" + rootID + ":officer_of"
		edges[edgeKey] = graph.Edge{Source: nodeID, Target: rootID, Kind: graph.EdgeKindOfficerOf}
		if o.ResignedOn == "" {
			activeOfficerIDs = append(activeOfficerIDs, officerID)
		}
	}

	// PSC nodes — both current and ceased; same toggle convention as officers.
	for _, p := range pscs {
		nodeID, nodeType := pscNodeID(&p)
		if _, exists := nodes[nodeID]; !exists {
			meta := graph.NodeMeta{
				NaturesOfControl: p.NaturesOfControl,
				NotifiedOn:       p.NotifiedOn,
				CeasedOn:         ceasedOn(&p),
				Nationality:      p.Nationality,
			}
			if nodeType == graph.NodeTypeCompany {
				meta.CompanyNumber = p.UKRegistrationNumber()
			}
			nodes[nodeID] = graph.Node{
				ID:    nodeID,
				Label: p.Name,
				Type:  nodeType,
				Meta:  meta,
			}
		}
		edgeKey := nodeID + "→" + rootID + ":psc_of"
		edges[edgeKey] = graph.Edge{Source: nodeID, Target: rootID, Kind: graph.EdgeKindPSCOf}
	}

	// Depth-2: fan out to each active officer's other appointments.
	// Capped at 5 officers to stay well within the rate limit.
	if depth >= 2 {
		cap := min(5, len(activeOfficerIDs))
		g2, g2ctx := errgroup.WithContext(ctx)
		type result struct {
			officerID    string
			appointments []Appointment
		}
		results := make([]result, cap)
		for i := range cap {
			oid := activeOfficerIDs[i]
			g2.Go(func() error {
				appts, err := b.ch.GetOfficerAppointments(g2ctx, oid)
				if err != nil {
					return nil // non-fatal: skip this officer
				}
				results[i] = result{officerID: oid, appointments: appts}
				return nil
			})
		}
		if err := g2.Wait(); err != nil {
			return nil, err
		}
		for _, r := range results {
			officerNodeID := "officer:" + r.officerID
			for _, appt := range r.appointments {
				if appt.ResignedOn != "" || appt.AppointedTo.CompanyNumber == companyNumber {
					continue
				}
				cid := "company:" + appt.AppointedTo.CompanyNumber
				if _, exists := nodes[cid]; !exists {
					nodes[cid] = graph.Node{
						ID:    cid,
						Label: appt.AppointedTo.CompanyName,
						Type:  graph.NodeTypeCompany,
						Meta: graph.NodeMeta{
							CompanyNumber: appt.AppointedTo.CompanyNumber,
							Status:        appt.AppointedTo.CompanyStatus,
						},
					}
				}
				ek := officerNodeID + "→" + cid + ":officer_of"
				edges[ek] = graph.Edge{
					Source: officerNodeID,
					Target: cid,
					Kind:   graph.EdgeKindOfficerOf,
				}
			}
		}
	}

	return graph.Collect(nodes, edges), nil
}

// ForOfficer builds the graph centred on an officer ID showing all their current appointments.
func (b *Builder) ForOfficer(ctx context.Context, officerID string) (*graph.Response, error) {
	appointments, err := b.ch.GetOfficerAppointments(ctx, officerID)
	if err != nil {
		return nil, err
	}

	nodes := make(map[string]graph.Node)
	edges := make(map[string]graph.Edge)

	officerNodeID := "officer:" + officerID
	label := "Officer"
	if len(appointments) > 0 && appointments[0].Name != "" {
		label = formatName(appointments[0].Name)
	}
	var role string
	if len(appointments) > 0 {
		role = appointments[0].Role
	}
	nodes[officerNodeID] = graph.Node{
		ID:    officerNodeID,
		Label: label,
		Type:  graph.NodeTypeOfficer,
		Meta: graph.NodeMeta{
			OfficerID: officerID,
			Role:      role,
			Expanded:  true,
		},
	}

	for _, a := range appointments {
		cid := "company:" + a.AppointedTo.CompanyNumber
		if _, exists := nodes[cid]; !exists {
			nodes[cid] = graph.Node{
				ID:    cid,
				Label: a.AppointedTo.CompanyName,
				Type:  graph.NodeTypeCompany,
				Meta: graph.NodeMeta{
					CompanyNumber: a.AppointedTo.CompanyNumber,
					Status:        a.AppointedTo.CompanyStatus,
					AppointedOn:   a.AppointedOn,
					ResignedOn:    a.ResignedOn,
				},
			}
		}
		ek := officerNodeID + "→" + cid + ":officer_of"
		edges[ek] = graph.Edge{Source: officerNodeID, Target: cid, Kind: graph.EdgeKindOfficerOf}
	}

	return graph.Collect(nodes, edges), nil
}

// extractOfficerID parses "/officers/{id}/appointments" → "{id}".
func extractOfficerID(appointmentsURL string) string {
	parts := strings.Split(appointmentsURL, "/")
	for i, part := range parts {
		if part == "officers" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// formatName converts CH's "SURNAME, Firstname" to "Firstname Surname".
func formatName(raw string) string {
	if surname, given, ok := strings.Cut(raw, ", "); ok {
		return given + " " + toTitleCase(surname)
	}
	return raw
}

func toTitleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// pscNodeID returns the canonical node ID and type for a PSC.
// UK-registered corporate PSCs reuse the company: namespace so they merge
// naturally with any company node already in the graph.
func pscNodeID(p *PSC) (string, graph.NodeType) {
	if p.IsCorporate() {
		if reg := p.UKRegistrationNumber(); reg != "" {
			return "company:" + reg, graph.NodeTypeCompany
		}
	}
	return "psc:" + sanitizeID(p.Name), graph.NodeTypePSC
}

func sanitizeID(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' {
			return '_'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return -1
	}, strings.ToLower(s))
}

// ceasedOn returns the ceased date for a PSC, or "ceased" when the CH API
// marks the PSC as ceased via the boolean flag but omits the effective date.
func ceasedOn(p *PSC) string {
	if p.CeasedToBeEffective != "" {
		return p.CeasedToBeEffective
	}
	if p.Ceased {
		return "ceased"
	}
	return ""
}
