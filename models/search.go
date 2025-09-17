package models

type SearchRequest struct {
    From string `json:"from"`
    To   string `json:"to"`
    Date string `json:"date"`
}

type FlightOption struct {
    Carrier  string  `json:"carrier"`
    Price    float64 `json:"price"`
    Duration string  `json:"duration"`
}
