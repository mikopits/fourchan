package fourchan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Thread struct {
	Board         *Board
	Id            int
	Topic         *Post
	Replies       []*Post
	OmittedPosts  int
	OmittedImages int
	WantUpdate    bool
	LastModified  time.Time
	LastReplyId   int
	Expired       bool
}

// Create a new Thread from an http.Response.
func NewThreadFromResponse(b *Board, resp *http.Response, id int) (*Thread, error) {
	switch resp.StatusCode {
	case 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var postsInfo PostsInfo
		err = json.Unmarshal(body, &postsInfo)
		if err != nil {
			return nil, err
		}
		return NewThreadFromPostsInfo(postsInfo, b, id, time.Time{})
	case 404:
		return nil, nil
	default:
		return nil, ErrUnexpectedResponse
	}
}

// Parses JSON from an http.Response.Body and returns a Thread.
func NewThreadFromPostsInfo(postsInfo PostsInfo, b *Board, id int, lastModified time.Time) (*Thread, error) {
	thread := &Thread{Board: b, Id: id, Expired: false}
	if !lastModified.Equal(time.Time{}) {
		thread.LastModified = lastModified
	}

	thread.Topic = NewPost(thread, &postsInfo.Posts[0])
	for _, post := range postsInfo.Posts[1:] {
		thread.Replies = append(thread.Replies, NewPost(thread, &post))
	}

	if id != 0 {
		if len(thread.Replies) == 0 {
			thread.LastReplyId = thread.Topic.Data.PostNumber
		} else {
			thread.LastReplyId = thread.Replies[len(thread.Replies)-1].Data.PostNumber
		}
	} else {
		thread.WantUpdate = true
		head := postsInfo.Posts[0]
		thread.Id = head.PostNumber
		thread.OmittedImages = head.OmittedImages
		thread.OmittedPosts = head.OmittedPosts
	}

	return thread, nil
}

// Returns the URLs of all the files in the thread.
func (t *Thread) Files() []string {
	var files []string
	for _, reply := range t.Replies {
		if reply.HasFile() {
			files = append(files, reply.FileURL())
		}
	}
	return files
}

// Returns the URLs of all the thumbnails in the thread.
func (t *Thread) Thumbs() []string {
	var thumbs []string
	for _, reply := range t.Replies {
		if reply.HasFile() {
			thumbs = append(thumbs, reply.ThumbnailURL())
		}
	}
	return thumbs
}

// Fetch new posts from the server. Returns an integer with the
// number of new posts. Set force = true to force the update.
func (t *Thread) Update(force bool) (int, error) {
	// Thread has died.
	if t.Expired && !force {
		return 0, nil
	}

	url := fmt.Sprintf("%s/%s", t.Board.BaseURL, fmt.Sprintf(ThreadURL, t.Board.Name, t.Id))

	resp, err := Request(url, t.LastModified)
	if err != nil {
		return 0, err
	}

	switch resp.StatusCode {
	case 304:
		return 0, nil
	case 404:
		t.Expired = true
		delete(t.Board.ThreadCache, t.Id)
		return 0, nil
	case 200:
		// If we somehow 404'd we should put the thread back in the cache.
		if t.Expired {
			t.Expired = false
			t.Board.ThreadCache[t.Id] = t
		}

		t.WantUpdate = true
		t.OmittedImages = 0
		t.OmittedPosts = 0

		lastModified, err := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))
		if err != nil {
			return 0, err
		}
		t.LastModified = lastModified

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		var posts PostsInfo
		err = json.Unmarshal(body, &posts)
		if err != nil {
			return 0, err
		}

		originalPostCount := len(t.Replies)

		t.Topic = NewPost(t, &posts.Posts[0])
		if t.LastReplyId != 0 && !force {
			for _, post := range posts.Posts[1:] {
				if post.PostNumber > t.LastReplyId {
					t.Replies = append(t.Replies, NewPost(t, &post))
				}
			}
		} else {
			t.Replies = nil
			for _, post := range posts.Posts[1:] {
				t.Replies = append(t.Replies, NewPost(t, &post))
			}
		}

		newPostCount := len(t.Replies)
		postCountDelta := newPostCount - originalPostCount
		if postCountDelta == 0 {
			return 0, nil
		}

		t.LastReplyId = t.Replies[len(t.Replies)-1].Data.PostNumber

		return postCountDelta, nil
	default:
		return 0, ErrUnexpectedResponse
	}
}

// Return the URL of the thread.
func (t *Thread) URL() string {
	return fmt.Sprintf("http://boards.4chan.org/%s/thread/%d", t.Board.Name, t.Id)
}
