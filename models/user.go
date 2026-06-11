package models

type UserShort struct {
	PK                      string `json:"pk"`
	Username                string `json:"username,omitempty"`
	FullName                string `json:"full_name,omitempty"`
	ProfilePicURL           string `json:"profile_pic_url,omitempty"`
	ProfilePicURLHD         string `json:"profile_pic_url_hd,omitempty"`
	IsPrivate               *bool  `json:"is_private,omitempty"`
	IsVerified              *bool  `json:"is_verified,omitempty"`
	LatestReelMedia         *int   `json:"latest_reel_media,omitempty"`
	HasAnonymousProfilePic  *bool  `json:"has_anonymous_profile_picture,omitempty"`
}

type User struct {
	PK                      string     `json:"pk"`
	Username                string     `json:"username"`
	FullName                string     `json:"full_name"`
	IsPrivate               bool       `json:"is_private"`
	ProfilePicURL           string     `json:"profile_pic_url"`
	ProfilePicURLHD         string     `json:"profile_pic_url_hd,omitempty"`
	IsVerified              bool       `json:"is_verified"`
	MediaCount              int        `json:"media_count"`
	FollowerCount           int        `json:"follower_count"`
	FollowingCount          int        `json:"following_count"`
	Biography               string     `json:"biography,omitempty"`
	BioLinks                []BioLink  `json:"bio_links,omitempty"`
	ExternalURL             string     `json:"external_url,omitempty"`
	AccountType             *int       `json:"account_type,omitempty"`
	IsBusiness              bool       `json:"is_business"`
	BroadcastChannel        []Broadcast `json:"broadcast_channel,omitempty"`
	PublicEmail             string     `json:"public_email,omitempty"`
	ContactPhoneNumber      string     `json:"contact_phone_number,omitempty"`
	PublicPhoneCountryCode  string     `json:"public_phone_country_code,omitempty"`
	PublicPhoneNumber       string     `json:"public_phone_number,omitempty"`
	BusinessContactMethod   string     `json:"business_contact_method,omitempty"`
	BusinessCategoryName    string     `json:"business_category_name,omitempty"`
	CategoryName            string     `json:"category_name,omitempty"`
	Category                string     `json:"category,omitempty"`
	AddressStreet           string     `json:"address_street,omitempty"`
	CityID                  string     `json:"city_id,omitempty"`
	CityName                string     `json:"city_name,omitempty"`
	Latitude                *float64   `json:"latitude,omitempty"`
	Longitude               *float64   `json:"longitude,omitempty"`
	Zip                     string     `json:"zip,omitempty"`
	InstagramLocationID     string     `json:"instagram_location_id,omitempty"`
	InteropMessagingUserFBID string    `json:"interop_messaging_user_fbid,omitempty"`
}

type BioLink struct {
	LinkID     string `json:"link_id,omitempty"`
	URL        string `json:"url"`
	LynxURL    string `json:"lynx_url,omitempty"`
	LinkType   string `json:"link_type,omitempty"`
	Title      string `json:"title,omitempty"`
	IsPinned   *bool  `json:"is_pinned,omitempty"`
}

type Broadcast struct {
	Title                   string `json:"title"`
	ThreadIGID              string `json:"thread_igid"`
	Subtitle                string `json:"subtitle"`
	InviteLink              string `json:"invite_link"`
	IsMember                bool   `json:"is_member"`
	GroupImageURI           string `json:"group_image_uri"`
	GroupImageBackgroundURI string `json:"group_image_background_uri"`
	ThreadSubtype           int    `json:"thread_subtype"`
	NumberOfMembers         int    `json:"number_of_members"`
	CreatorIGID             string `json:"creator_igid,omitempty"`
	CreatorUsername         string `json:"creator_username"`
}

type Viewer struct {
	UserShort
	HasLiked      bool   `json:"has_liked"`
	ReplyText      string `json:"reply_text,omitempty"`
	IsSpamViewer   bool   `json:"is_spam_viewer"`
}

type Usertag struct {
	User UserShort `json:"user"`
	X    float64   `json:"x"`
	Y    float64   `json:"y"`
}

type Relationship struct {
	UserID            string `json:"user_id"`
	Blocking          bool   `json:"blocking"`
	FollowedBy        bool   `json:"followed_by"`
	Following         bool   `json:"following"`
	IncomingRequest   bool   `json:"incoming_request"`
	IsBestie          bool   `json:"is_bestie"`
	IsBlockingReel    bool   `json:"is_blocking_reel"`
	IsMutingReel      bool   `json:"is_muting_reel"`
	IsPrivate         bool   `json:"is_private"`
	IsRestricted      bool   `json:"is_restricted"`
	Muting            bool   `json:"muting"`
	OutgoingRequest   bool   `json:"outgoing_request"`
}

type RelationshipShort struct {
	UserID            string `json:"user_id"`
	Following         bool   `json:"following"`
	IncomingRequest   bool   `json:"incoming_request"`
	IsBestie          bool   `json:"is_bestie"`
	IsFeedFavorite    bool   `json:"is_feed_favorite"`
	IsPrivate         bool   `json:"is_private"`
	IsRestricted      bool   `json:"is_restricted"`
	OutgoingRequest   bool   `json:"outgoing_request"`
}
