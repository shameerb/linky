package models

import "time"

type Link struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
}

type Subject struct {
	Subject string `json:"subject"`
	Links   []Link `json:"links"`
}

type Response struct {
	Data []Subject `json:"data"`
}

type DeleteLink struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type DeleteLinkRequest struct {
	Filename string       `json:"filename"`
	Links    []DeleteLink `json:"links"`
}

type BulkLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type BulkLinksRequest struct {
	Filename string     `json:"filename"`
	Subject  string     `json:"subject"`
	Links    []BulkLink `json:"links"`
}
