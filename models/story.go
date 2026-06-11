package models

type Story struct {
	PK                  string           `json:"pk"`
	ID                  string           `json:"id"`
	Code                string           `json:"code"`
	TakenAt             Timestamp        `json:"taken_at"`
	ImportedTakenAt     *Timestamp        `json:"imported_taken_at,omitempty"`
	MediaType           MediaType        `json:"media_type"`
	ProductType         string           `json:"product_type,omitempty"`
	ThumbnailURL        string           `json:"thumbnail_url,omitempty"`
	User                UserShort        `json:"user"`
	VideoURL            string           `json:"video_url,omitempty"`
	VideoDuration       *float64         `json:"video_duration,omitempty"`
	SponsorTags         []UserShort      `json:"sponsor_tags"`
	IsPaidPartnership   *bool            `json:"is_paid_partnership,omitempty"`
	Mentions            []StoryMention   `json:"mentions"`
	Links               []StoryLink      `json:"links"`
	Hashtags            []StoryHashtag   `json:"hashtags"`
	Locations           []StoryLocation  `json:"locations"`
	Stickers            []StorySticker   `json:"stickers"`
	Medias              []StoryMedia     `json:"medias"`
	Polls               []StoryPoll      `json:"polls"`
}

type StoryMention struct {
	User      UserShort `json:"user"`
	X         *float64  `json:"x,omitempty"`
	Y         *float64  `json:"y,omitempty"`
	Width     *float64  `json:"width,omitempty"`
	Height    *float64  `json:"height,omitempty"`
	Rotation  *float64  `json:"rotation,omitempty"`
}

type StoryHashtag struct {
	Hashtag  Hashtag   `json:"hashtag"`
	X        *float64  `json:"x,omitempty"`
	Y        *float64  `json:"y,omitempty"`
	Width    *float64  `json:"width,omitempty"`
	Height   *float64  `json:"height,omitempty"`
	Rotation *float64  `json:"rotation,omitempty"`
}

type StoryLocation struct {
	Location Location  `json:"location"`
	X        *float64  `json:"x,omitempty"`
	Y        *float64  `json:"y,omitempty"`
	Width    *float64  `json:"width,omitempty"`
	Height   *float64  `json:"height,omitempty"`
	Rotation *float64  `json:"rotation,omitempty"`
}

type StoryMedia struct {
	X           float64  `json:"x"`
	Y           float64  `json:"y"`
	Z           float64  `json:"z"`
	Width       float64  `json:"width"`
	Height      float64  `json:"height"`
	Rotation    float64  `json:"rotation"`
	IsPinned    *bool    `json:"is_pinned,omitempty"`
	IsHidden    *bool    `json:"is_hidden,omitempty"`
	IsSticker   *bool    `json:"is_sticker,omitempty"`
	IsFbSticker *bool    `json:"is_fb_sticker,omitempty"`
	MediaPK     int      `json:"media_pk"`
	UserID      *int     `json:"user_id,omitempty"`
	ProductType string   `json:"product_type,omitempty"`
	MediaCode   string   `json:"media_code,omitempty"`
}

type StorySticker struct {
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"`
	X        float64            `json:"x"`
	Y        float64            `json:"y"`
	Z        *int               `json:"z,omitempty"`
	Width    float64            `json:"width"`
	Height   float64            `json:"height"`
	Rotation *float64           `json:"rotation,omitempty"`
	StoryLink *StoryStickerLink `json:"story_link,omitempty"`
	Extra    map[string]any     `json:"extra,omitempty"`
}

type StoryStickerLink struct {
	URL         string `json:"url"`
	LinkTitle   string `json:"link_title,omitempty"`
	LinkType    string `json:"link_type,omitempty"`
	DisplayURL  string `json:"display_url,omitempty"`
}

type StoryLink struct {
	WebURI    string  `json:"webUri"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Z         float64 `json:"z"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Rotation  float64 `json:"rotation"`
}

type StoryPoll struct {
	ID               string         `json:"id,omitempty"`
	Type             string         `json:"type,omitempty"`
	X                float64        `json:"x"`
	Y                float64        `json:"y"`
	Z                *int           `json:"z,omitempty"`
	Width            float64        `json:"width"`
	Height           float64        `json:"height"`
	Rotation         *float64       `json:"rotation,omitempty"`
	IsMultiOption    *bool          `json:"is_multi_option,omitempty"`
	IsSharedResult   *bool          `json:"is_shared_result,omitempty"`
	ViewerCanVote    *bool          `json:"viewer_can_vote,omitempty"`
	Finished         *bool          `json:"finished,omitempty"`
	Color            string         `json:"color,omitempty"`
	PollType         string         `json:"poll_type,omitempty"`
	Question         string         `json:"question"`
	Options          []any          `json:"options"`
	Extra            map[string]any `json:"extra,omitempty"`
}

type StoryArchiveDay struct {
	ID                string    `json:"id"`
	Timestamp         Timestamp `json:"timestamp"`
	MediaCount        int       `json:"media_count"`
	ReelType          string    `json:"reel_type"`
	LatestReelMedia   *int      `json:"latest_reel_media,omitempty"`
}

type Highlight struct {
	PK               string      `json:"pk"`
	ID               string      `json:"id"`
	LatestReelMedia  int         `json:"latest_reel_media"`
	CoverMedia       map[string]any `json:"cover_media"`
	User             UserShort   `json:"user"`
	Title            string      `json:"title"`
	CreatedAt        Timestamp   `json:"created_at"`
	IsPinnedHighlight bool       `json:"is_pinned_highlight"`
	MediaCount       int         `json:"media_count"`
	MediaIDs         []int       `json:"media_ids"`
	Items            []Story     `json:"items"`
}
