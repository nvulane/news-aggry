package models

import "time"

type Feed struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Created  time.Time `json:"created_at"`
	Updated  time.Time `json:"updated_at"`
	Category string    `json:"category"`
	URL      string    `json:"url"`
}

type Article struct {
	Title       string
	Description string
	Link        string
	Published   time.Time
}

