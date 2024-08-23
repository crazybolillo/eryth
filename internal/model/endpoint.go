package model

type Endpoint struct {
	Sid         int32    `json:"sid"`
	ID          string   `json:"id"`
	DisplayName string   `json:"displayName"`
	Transport   string   `json:"transport"`
	Context     string   `json:"context"`
	Codecs      []string `json:"codecs"`
	MaxContacts int32    `json:"maxContacts"`
	Extension   string   `json:"extension"`
}

type NewEndpoint struct {
	ID          string   `json:"id"`
	Password    string   `json:"password"`
	Transport   string   `json:"transport,omitempty"`
	Context     string   `json:"context"`
	Codecs      []string `json:"codecs"`
	MaxContacts int32    `json:"maxContacts,omitempty"`
	Extension   string   `json:"extension,omitempty"`
	DisplayName string   `json:"displayName"`
}

type PatchedEndpoint struct {
	Password    *string  `json:"password,omitempty"`
	DisplayName *string  `json:"displayName,omitempty"`
	Transport   *string  `json:"transport,omitempty"`
	Context     *string  `json:"context,omitempty"`
	Codecs      []string `json:"codecs,omitempty"`
	MaxContacts *int32   `json:"maxContacts,omitempty"`
	Extension   *string  `json:"extension,omitempty"`
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
