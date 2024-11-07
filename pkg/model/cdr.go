package model

import "time"

type CallRecord struct {
	ID          int64      `json:"id"`
	From        string     `json:"from"`
	To          string     `json:"to"`
	Context     string     `json:"context"`
	Start       time.Time  `json:"start"`
	Answer      *time.Time `json:"answer"`
	End         time.Time  `json:"end"`
	Duration    int32      `json:"duration"`
	BillSeconds int32      `json:"billsec"`
}

type CallRecordPage struct {
	Total     int64        `json:"total"`
	Retrieved int          `json:"retrieved"`
	Records   []CallRecord `json:"records"`
}
