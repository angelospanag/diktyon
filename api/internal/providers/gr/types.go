package gr

// SearchResponse is the envelope returned by GET /companies.
type SearchResponse struct {
	SearchMetadata struct {
		TotalCount    int    `json:"totalCount"`
		ResultsOffset int    `json:"resultsOffset"`
		ResultsSize   string `json:"resultsSize"`
	} `json:"searchMetadata"`
	SearchResults []Company `json:"searchResults"`
}

// Company is the full company record returned by both list and detail endpoints.
type Company struct {
	ArGemi            int               `json:"arGemi"`
	Afm               string            `json:"afm"`
	CoNameEl          string            `json:"coNameEl"`
	CoNamesEn         []string          `json:"coNamesEn"`
	CoTitlesEl        []string          `json:"coTitlesEl"`
	City              string            `json:"city"`
	Street            string            `json:"street"`
	StreetNumber      string            `json:"streetNumber"`
	ZipCode           string            `json:"zipCode"`
	LegalType         LegalType         `json:"legalType"`
	Status            CompanyStatus     `json:"status"`
	IncorporationDate string            `json:"incorporationDate"`
	Activities        []CompanyActivity `json:"activities"`
	Persons           []CompanyPerson   `json:"persons"`
	IsBranch          bool              `json:"isBranch"`
	Objective         string            `json:"objective"`
}

type LegalType struct {
	ID      int    `json:"id"`
	Descr   string `json:"descr"`
	DescrEn string `json:"descrEn"`
}

type CompanyStatus struct {
	ID       int    `json:"id"`
	Descr    string `json:"descr"`
	DescrEn  string `json:"descrEn"`
	IsActive bool   `json:"isActive"`
}

type CompanyActivity struct {
	Activity struct {
		ID    string `json:"id"`
		Descr string `json:"descr"`
	} `json:"activity"`
	Type   string `json:"type"`
	DtFrom string `json:"dtFrom"`
	DtTo   string `json:"dtTo"`
}

// CompanyPerson is an individual or corporate entity associated with a company.
// If PersonName is set it is a natural person; if BusinessName is set it is a legal entity.
type CompanyPerson struct {
	PersonName               string `json:"personName"`
	BusinessName             string `json:"businessName"`
	Role                     string `json:"role"`
	DtFrom                   string `json:"dtFrom"`
	DtTo                     string `json:"dtTo"`
	IsRepresentativeAlone    bool   `json:"isRepresentativeAlone"`
	IsRepresentativeInCommon bool   `json:"isRepresentativeInCommon"`
	Percentage               string `json:"percentage"`
	Category                 string `json:"category"`
}
