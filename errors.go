package fourchan

import "errors"

var (
	ErrUnexpectedResponse = errors.New("fourchan: got unexpected response code")
	ErrCatalogNotFound    = errors.New("fourchan: failed to retrieve catalog")
	ErrPageNotFound       = errors.New("fourchan: failed to retrieve page")
)
