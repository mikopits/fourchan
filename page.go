package fourchan

// Struct to hold page information from a JSON Unmarshal.
type Page struct {
	Threads []PostsInfo `json:"threads"`
}
