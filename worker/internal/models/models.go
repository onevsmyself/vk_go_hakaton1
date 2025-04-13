package models

type Data struct {
	ID        int64  `json:"id,omitempty"`
	ASIN      string `json:"asin,omitempty"`
	Time      string `json:"time,omitempty"`
	Title     string `json:"title,omitempty"`
	Group     string `json:"group,omitempty"`
	SalesRank int64  `json:"salesrank,omitempty"`
}

type SearchStreamResponse struct {
	Data []Data `json:"data"`
}

type SearchStatResponse struct {
	Query        string `json:"query"`
	WorkTime     string `json:"work_time"`
	ResultsCount int64  `json:"results_count"`
	Active       string `json:"active"`
}

type SearchRequest struct {
	Query      string   `json:"query"`
	TypeSearch string   `json:"type_search"`
	Fields     []string `json:"fields"`
	Table      string   `json:"table"`
	Group      string   `json:"group"`
}
