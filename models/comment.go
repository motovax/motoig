package models

import "time"

type Comment struct {
	PK                    string    `json:"pk"`
	Text                  string    `json:"text"`
	User                  UserShort `json:"user"`
	CreatedAtUTC          time.Time `json:"created_at_utc"`
	ContentType           string    `json:"content_type"`
	Status                string    `json:"status"`
	RepliedToCommentID    *string   `json:"replied_to_comment_id,omitempty"`
	HasLiked              *bool     `json:"has_liked,omitempty"`
	LikeCount             *int      `json:"like_count,omitempty"`
}

type Location struct {
	PK                 *int64  `json:"pk,omitempty"`
	Name               string  `json:"name"`
	Phone              string  `json:"phone,omitempty"`
	Website            string  `json:"website,omitempty"`
	Category           string  `json:"category,omitempty"`
	Hours              map[string]any `json:"hours,omitempty"`
	Address            string  `json:"address,omitempty"`
	City               string  `json:"city,omitempty"`
	Zip                string  `json:"zip,omitempty"`
	Lng                *float64 `json:"lng,omitempty"`
	Lat                *float64 `json:"lat,omitempty"`
	ExternalID         *int64  `json:"external_id,omitempty"`
	ExternalIDSource   string  `json:"external_id_source,omitempty"`
}

type Hashtag struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	MediaCount     *int    `json:"media_count,omitempty"`
	ProfilePicURL  string  `json:"profile_pic_url,omitempty"`
}

type Collection struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	MediaCount  int    `json:"media_count"`
}

type Track struct {
	ID                             string         `json:"id"`
	Title                          string         `json:"title"`
	Subtitle                       string         `json:"subtitle"`
	DisplayArtist                  string         `json:"display_artist"`
	AudioClusterID                 int            `json:"audio_cluster_id"`
	MusicCanonicalID               string         `json:"music_canonical_id,omitempty"`
	ArtistID                       *int           `json:"artist_id,omitempty"`
	CoverArtworkURI                string         `json:"cover_artwork_uri,omitempty"`
	CoverArtworkThumbnailURI       string         `json:"cover_artwork_thumbnail_uri,omitempty"`
	ProgressiveDownloadURL         string         `json:"progressive_download_url,omitempty"`
	FastStartProgressiveDownloadURL string        `json:"fast_start_progressive_download_url,omitempty"`
	ReactiveAudioDownloadURL       string         `json:"reactive_audio_download_url,omitempty"`
	HighlightStartTimesInMs        []int          `json:"highlight_start_times_in_ms"`
	IsExplicit                     bool           `json:"is_explicit"`
	DashManifest                   string         `json:"dash_manifest"`
	URI                            string         `json:"uri,omitempty"`
	HasLyrics                      bool           `json:"has_lyrics"`
	AudioAssetID                   int            `json:"audio_asset_id"`
	DurationInMs                   int            `json:"duration_in_ms"`
	DarkMessage                    string         `json:"dark_message,omitempty"`
	AllowsSaving                   bool           `json:"allows_saving"`
	TerritoryValidityPeriods       map[string]any `json:"territory_validity_periods"`
}

type Note struct {
	ID            string    `json:"id"`
	Text          string    `json:"text"`
	UserID        string    `json:"user_id"`
	User          UserShort `json:"user"`
	Audience      int       `json:"audience"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	IsEmojiOnly   bool      `json:"is_emoji_only"`
	HasTranslation bool     `json:"has_translation"`
	NoteStyle     int       `json:"note_style"`
}
