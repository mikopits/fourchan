package fourchan

// Struct to hold catalog information from a JSON Unmarshal.
type PageInfo struct {
	Page    int        `json:"page"`
	Threads []PostData `json:"threads"`
}
