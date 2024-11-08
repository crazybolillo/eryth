package model

type Location struct {
	ID        string `json:"id"`
	Endpoint  string `json:"endpoint"`
	Address   string `json:"address"`
	UserAgent string `json:"user_agent"`
}

type LocationPage struct {
	Total     int64      `json:"total"`
	Retrieved int        `json:"retrieved"`
	Locations []Location `json:"locations"`
}
