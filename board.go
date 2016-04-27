package fourchan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	BoardBaseURL = "a.4cdn.org"
	BoardURL     = "%s/%s.json"
	ThreadURL    = "%s/res/%d.json"
)

type Board struct {
	Name                string
	Https               bool
	BaseURL             string
	ThreadCache         map[int]*Thread
	PageCache           map[int][]*Thread
	PageLastModified    map[int]time.Time
	CatalogCache        []*Thread
	CatalogLastModified time.Time
}

func NewBoard(name string, https bool) *Board {
	b := &Board{
		Name:             name,
		Https:            https,
		ThreadCache:      make(map[int]*Thread),
		PageCache:        make(map[int][]*Thread),
		PageLastModified: make(map[int]time.Time),
	}

	if https {
		b.BaseURL = fmt.Sprintf("%s%s", "https://", BoardBaseURL)
	} else {
		b.BaseURL = fmt.Sprintf("%s%s", "http://", BoardBaseURL)
	}

	return b
}

// Retrieve a thread from 4chan or update and return the thread if it is already cached.
func (b *Board) GetThread(id int) (*Thread, error) {
	if thread, cached := b.ThreadCache[id]; cached {
		thread.Update(false)
		return thread, nil
	}

	url := fmt.Sprintf("%s/%s", b.BaseURL, fmt.Sprintf(ThreadURL, b.Name, id))
	resp, err := Request(url, time.Time{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	thread, err := NewThreadFromResponse(b, resp, id)
	if err != nil {
		return nil, err
	}
	b.ThreadCache[id] = thread

	return thread, nil
}

// Check if a thread exists. Makes a request.
func (b *Board) ThreadExists(id int) (bool, error) {
	url := fmt.Sprintf("%s/%s", b.BaseURL, fmt.Sprintf(ThreadURL, b.Name, id))
	resp, err := Request(url, time.Time{})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// Get thread by pages. If the thread is already in the cache, return the
// cached entry.
//
// Sets Thread.WantsUpdate to true if the thread is being returned from
// the cache.
func (b *Board) GetThreadsByPage(page int) ([]*Thread, error) {
	uri := fmt.Sprintf("%s/%s", b.BaseURL, fmt.Sprintf(BoardURL, b.Name, strconv.Itoa(page)))
	var lastModified time.Time
	if t, ok := b.PageLastModified[page]; ok {
		lastModified = t
	}
	resp, err := Request(uri, lastModified)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		lastModified, err := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))
		if err != nil {
			return nil, err
		}
		b.PageLastModified[page] = lastModified
		delete(b.PageCache, page)

		var p Page
		err = json.Unmarshal(body, &p)
		if err != nil {
			return nil, err
		}

		for _, postsInfo := range p.Threads {
			var thread *Thread
			var cached bool
			id := postsInfo.Posts[0].PostNumber
			if thread, cached = b.ThreadCache[id]; cached {
				thread.Update(false)
			} else {
				thread, err = NewThreadFromPostsInfo(postsInfo, b, 0, time.Time{})
				if err != nil {
					return nil, err
				}
				b.ThreadCache[thread.Id] = thread
			}
			cached = false
			for _, cachedThread := range b.PageCache[page] {
				if cachedThread == thread {
					cached = true
				}
			}
			if !cached {
				b.PageCache[page] = append(b.PageCache[page], thread)
			}
		}
		return b.PageCache[page], nil
	case 304:
		return b.PageCache[page], nil
	case 404:
		panic(ErrPageNotFound)
	default:
		panic(ErrUnexpectedResponse)
	}
}

// Get all threads on a board with a single request.
func (b *Board) GetCatalog() ([]*Thread, error) {
	url := fmt.Sprintf("%s/%s", b.BaseURL, fmt.Sprintf(BoardURL, b.Name, "catalog"))
	resp, err := Request(url, b.CatalogLastModified)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		b.CatalogLastModified, err = time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))
		if err != nil {
			return nil, err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		pages := []PageInfo{}
		err = json.Unmarshal(body, &pages)
		if err != nil {
			return nil, err
		}

		fmt.Println(pages)

		b.CatalogCache = nil
		for _, page := range pages {
			for _, thread := range page.Threads {
				id := thread.PostNumber
				var thread *Thread
				var cached bool
				if thread, cached = b.ThreadCache[id]; cached {
					thread.Update(false)
				} else {
					thread, err = b.GetThread(id)
					if err != nil {
						return nil, err
					}
				}
				b.CatalogCache = append(b.CatalogCache, thread)
			}
		}
		return b.CatalogCache, nil
	case 304:
		return b.CatalogCache, nil
	case 404:
		return nil, ErrCatalogNotFound
	default:
		return nil, ErrUnexpectedResponse
	}
}

// Make an HTTP Get request.
func Request(url string, lastModified time.Time) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("go-fourchan/%s", Version))
	if lastModified.Equal(time.Time{}) {
		req.Header.Set("If-Modified-Since", lastModified.UTC().Format(http.TimeFormat))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
