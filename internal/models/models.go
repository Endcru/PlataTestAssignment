package models

import "time"

type Quotation struct {
	Name string `json:"name"`
	Rate float64 `json:"rate"`
	UpdatedAt time.Time `json:"updated_at"`
}

type QuotationRequest struct {
	id int `json:"id"`
	Name string `json:"name"`
	RequestedAt time.Time `json:"requested_at"`
	CompletedAt time.Time `json:"completed_at"`
	Done bool `json:"done"`
}

type QuotationUpdate struct {
	id int `json:"id"`
	Name string `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
	PreviousRate float64 `json:"previous_rate"`
	NewRate float64 `json:"new_rate"`
	Source string `json:"source"`
}