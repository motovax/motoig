// Package models defines data types for the Instagram API.
package models

import (
	"encoding/json"
	"time"
)

type MediaType int

const (
	MediaTypeImage MediaType = 1
	MediaTypeVideo MediaType = 2
	MediaTypeAlbum MediaType = 8
)

func (m MediaType) String() string {
	switch m {
	case MediaTypeImage:
		return "image"
	case MediaTypeVideo:
		return "video"
	case MediaTypeAlbum:
		return "album"
	default:
		return "unknown"
	}
}

type Timestamp struct {
	time.Time
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		tt, err := time.Parse(time.RFC3339, s)
		if err == nil {
			t.Time = tt
			return nil
		}
	}
	var ms int64
	if err := json.Unmarshal(data, &ms); err == nil {
		t.Time = time.Unix(ms/1000, (ms%1000)*int64(time.Millisecond))
		return nil
	}
	return nil
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.Format(time.RFC3339))
}
