package model

type Endpoint struct {
	Sid         int32    `json:"sid"`
	ID          string   `json:"id"`
	AccountCode string   `json:"accountCode"`
	DisplayName string   `json:"displayName"`
	Transport   string   `json:"transport"`
	Context     string   `json:"context"`
	Codecs      []string `json:"codecs"`
	MaxContacts int32    `json:"maxContacts"`
	Extension   string   `json:"extension"`
	Nat         bool     `json:"nat"`
	Encryption  string   `json:"encryption"`
}

type NewEndpoint struct {
	ID          string   `json:"id"`
	AccountCode string   `json:"accountCode"`
	Password    string   `json:"password"`
	Transport   string   `json:"transport"`
	Context     string   `json:"context"`
	Codecs      []string `json:"codecs"`
	MaxContacts int32    `json:"maxContacts"`
	Extension   string   `json:"extension"`
	DisplayName string   `json:"displayName"`
	Nat         bool     `json:"nat"`
	Encryption  string   `json:"encryption"`
}

type PatchedEndpoint struct {
	Password    *string  `json:"password,"`
	DisplayName *string  `json:"displayName,"`
	Transport   *string  `json:"transport,"`
	Context     *string  `json:"context,"`
	Codecs      []string `json:"codecs,"`
	MaxContacts *int32   `json:"maxContacts,"`
	Extension   *string  `json:"extension,"`
	Nat         *bool    `json:"nat,"`
	Encryption  *string  `json:"encryption,"`
}

type EndpointPageEntry struct {
	Sid         int32  `json:"sid"`
	ID          string `json:"id"`
	Extension   string `json:"extension"`
	Context     string `json:"context"`
	DisplayName string `json:"displayName"`
}

type EndpointPage struct {
	Total     int64               `json:"total"`
	Retrieved int                 `json:"retrieved"`
	Endpoints []EndpointPageEntry `json:"endpoints"`
}
