package model

import "time"

type Transaction struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	Notes           string    `json:"notes"`
	TransactionDate time.Time `json:"transaction_date"`
	Type            string    `json:"type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
