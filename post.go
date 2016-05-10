package fourchan

import (
	"fmt"
)

const (
	BoardsURL = "a.4cdn.org"
	ImagesURL = "i.4cdn.org"
	ThumbsURL = "t.4cdn.org"
)

type Post struct {
	Thread *Thread
	Data   *PostData
}

type PostData struct {
	PostNumber      int        `json:"no"`
	Resto           int        `json:"resto"`
	Sticky          int        `json:"sticky"`
	Closed          int        `json:"closed"`
	Archived        int        `json:"archived"`
	Now             string     `json:"now"`
	Time            int        `json:"time"`
	Name            string     `json:"tripcode"`
	Tripcode        string     `json:"trip"`
	Id              string     `json:"id"`
	Capcode         string     `json:"capcode"`
	Country         string     `json:"country"`
	CountryName     string     `json:"country_name"`
	Subject         string     `json:"sub"`
	Comment         string     `json:"com"`
	Tim             int        `json:"tim"`
	Filename        string     `json:"filename"`
	Extension       string     `json:"ext"`
	Filesize        int        `json:"fsize"`
	Md5             string     `json:"md5"`
	ImageWidth      int        `json:"w"`
	ImageHeight     int        `json:"h"`
	ThumbnailWidth  int        `json:"tn_w"`
	ThumbnailHeight int        `json:"tn_h"`
	FileDeleted     int        `json:"filedeleted"`
	Spoiler         int        `json:"spoiler"`
	CustomSpoiler   int        `json:"custom_spoiler"`
	OmittedPosts    int        `json:"omitted_posts"`
	OmittedImages   int        `json:"omitted_images"`
	Replies         int        `json:"replies"`
	Images          int        `json:"images"`
	BumpLimit       int        `json:"bumplimit"`
	ImageLimit      int        `json:"imagelimit"`
	CapcodeReplies  []int      `json:"capcode_replies"`
	LastModified    int        `json:"last_modified"`
	Tag             string     `json:"tag"`
	SemanticURL     string     `json:"semantic_url"`
	LastReplies     []PostData `json:"last_replies"`
}

// Post struct has structure for JSON unmarshaller.
// See the 4chan API at https://github.com/4chan/4chan-API for me details.
type PostsInfo struct {
	Posts []PostData `json:"posts"`
}

func NewPost(t *Thread, data *PostData) *Post {
	return &Post{
		Thread: t,
		Data:   data,
	}
}

// Returns the URL of the attached file.
// Returns an empty string if there is no attached file.
func (p *Post) FileURL() string {
	if !p.HasFile() {
		return ""
	}

	board := p.Thread.Board
	var protocol string
	if board.Https {
		protocol = "https"
	} else {
		protocol = "http"
	}

	return fmt.Sprintf("%s://%s/%s/src/%d%s", protocol, ImagesURL, board.Name, p.Data.Tim, p.Data.Extension)
}

// Returns true if the Post has an attached file.
func (p *Post) HasFile() bool {
	return p.Data.Filename != ""
}

// Returns the URL of the attached thumbnail.
// Returns an empty string if there is no attached file.
func (p *Post) ThumbnailURL() string {
	if !p.HasFile() {
		return ""
	}

	board := p.Thread.Board
	var protocol string
	if board.Https {
		protocol = "https"
	} else {
		protocol = "http"
	}

	return fmt.Sprintf("%s://%s/%s/thumb/%ds.jpg", protocol, ThumbsURL, board.Name, p.Data.Tim)
}
