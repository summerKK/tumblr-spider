package module

import "encoding/json"

type Post struct {
	ID            json.Number `json:"id,Number"`
	Type          string
	PhotoURL      string      `json:"photo-url-1280"`
	Photos        []Post      `json:"photos,omitempty"`
	UnixTimestamp int64       `json:"unix-timestamp"`
	PhotoCaption  string      `json:"photo-caption"`

	RegularBody string `json:"regular-body"`

	Answer string

	Video json.RawMessage `json:"video-player"`
	// For links to outside sites.
	VideoCaption string `json:"video-caption"`
}

type TumblrLog struct {
	Posts      []Post `json:"posts"`
	TotalPosts int    `json:"posts-total"`
}
