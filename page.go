package fourchan

// Struct to hold page information form a JSON Unmarshal.
type Page struct {
	Threads []PostsInfo `json:"threads"`
}
