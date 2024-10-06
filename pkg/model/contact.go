package model

type Contact struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type ContactPage struct {
	Total     int64     `json:"total"`
	Retrieved int       `json:"retrieved"`
	Contacts  []Contact `json:"contacts"`
}
