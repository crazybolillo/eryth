package model

type BouncerResponse struct {
	Allow       bool   `json:"allow"`
	Destination string `json:"destination"`
	CallerID    string `json:"callerid"`
}
