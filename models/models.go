package models

import (
	"time"
)

type Item struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type ItemResponse struct {
	Item Item `json:"item"`
	Ok   bool `json:"ok"`
}

type ListResponse struct {
	Item []Item `json:"items"`
	Ok   bool   `json:"ok"`
}

type Transaction struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	TransactionType   string    `json:"transaction_type"`
	Category          string    `json:"category"`
	Date              time.Time `json:"date"`
	Description       string    `json:"description"`
	ConvertedAmount   float64   `json:"converted_amount,omitempty"`
	ConvertedCurrency string    `json:"converted_currency,omitempty"`
}
type Commission struct {
	TransactionID   int     `json:"transaction_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	TransactionType string  `json:"transaction_type"`
	Commission      float64 `json:"commission"`
	Date            string  `json:"date"`
	Description     string  `json:"description"`
}

type Rate struct {
	CurrencyCode string  `json:"currency_code"`
	Rate         float64 `json:"rate"`
}

type CurrencyRates struct {
	Rates map[string]float64 `json:"conversion_rates"`
}
type TransactionResponse struct {
	Transaction Transaction `json:"transaction"`
	Commission  *Commission `json:"commission,omitempty"`
}
