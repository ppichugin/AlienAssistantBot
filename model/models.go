package model

// ExchangeRates represents model for currency pair exchange rates.
type ExchangeRates struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
}
