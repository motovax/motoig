package models

import "time"

type DirectThread struct {
	PK                              string                    `json:"pk"`
	ID                              string                    `json:"id"`
	Messages                        []DirectMessage           `json:"messages"`
	Users                           []UserShort               `json:"users"`
	Inviter                         *UserShort                `json:"inviter,omitempty"`
	LeftUsers                       []UserShort               `json:"left_users"`
	AdminUserIDs                    []any                     `json:"admin_user_ids"`
	LastActivityAt                  Timestamp                 `json:"last_activity_at"`
	Muted                           bool                      `json:"muted"`
	IsPin                           *bool                     `json:"is_pin,omitempty"`
	Named                           bool                      `json:"named"`
	Canonical                       bool                      `json:"canonical"`
	Pending                         bool                      `json:"pending"`
	Archived                        bool                      `json:"archived"`
	ThreadType                      string                    `json:"thread_type"`
	ThreadTitle                     string                    `json:"thread_title"`
	Folder                          int                       `json:"folder"`
	VCMuted                         bool                      `json:"vc_muted"`
	IsGroup                         bool                      `json:"is_group"`
	MentionsMuted                   bool                      `json:"mentions_muted"`
	ApprovalRequiredForNewMembers   bool                      `json:"approval_required_for_new_members"`
	InputMode                       int                       `json:"input_mode"`
	BusinessThreadFolder            *int                      `json:"business_thread_folder,omitempty"`
	ReadState                       *int                      `json:"read_state,omitempty"`
	IsCloseFriendThread             bool                      `json:"is_close_friend_thread"`
	AssignedAdminID                 *int                      `json:"assigned_admin_id,omitempty"`
	SHHModeEnabled                  *bool                     `json:"shh_mode_enabled,omitempty"`
	LastSeenAt                      map[string]LastSeenInfo   `json:"last_seen_at"`
}

type LastSeenInfo struct {
	ItemID                          string                        `json:"item_id"`
	Timestamp                       Timestamp                     `json:"timestamp"`
	CreatedAt                       Timestamp                     `json:"created_at"`
	SHHSeenState                    map[string]any                `json:"shh_seen_state"`
	DisappearingMessagesSeenState   *DisappearingMessagesSeenState `json:"disappearing_messages_seen_state,omitempty"`
}

type DisappearingMessagesSeenState struct {
	ItemID    string    `json:"item_id"`
	Timestamp Timestamp `json:"timestamp"`
	CreatedAt Timestamp `json:"created_at"`
}

type DirectMessage struct {
	ID              string            `json:"id"`
	UserID          string            `json:"user_id,omitempty"`
	ThreadID        *int64            `json:"thread_id,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
	ItemType        string            `json:"item_type,omitempty"`
	IsSentByViewer  *bool             `json:"is_sent_by_viewer,omitempty"`
	IsSHHMode       *bool             `json:"is_shh_mode,omitempty"`
	Reactions       *MessageReactions  `json:"reactions,omitempty"`
	Text            string            `json:"text,omitempty"`
	Reply           *ReplyMessage     `json:"reply,omitempty"`
	Link            *MessageLink      `json:"link,omitempty"`
	AnimatedMedia   map[string]any    `json:"animated_media,omitempty"`
	Media           *DirectMedia      `json:"media,omitempty"`
	VisualMedia     *VisualMedia      `json:"visual_media,omitempty"`
	MediaShare      *Media            `json:"media_share,omitempty"`
	ReelShare       map[string]any    `json:"reel_share,omitempty"`
	StoryShare      map[string]any    `json:"story_share,omitempty"`
	FelixShare      map[string]any    `json:"felix_share,omitempty"`
	XMAShare        *MediaXma         `json:"xma_share,omitempty"`
	GenericXMA      []MediaXma        `json:"generic_xma,omitempty"`
	RawXMA          map[string]any    `json:"raw_xma,omitempty"`
	Clip            *Media            `json:"clip,omitempty"`
	Placeholder     map[string]any    `json:"placeholder,omitempty"`
	ClientContext   string            `json:"client_context,omitempty"`
}

type MessageReactions struct {
	Likes      []map[string]any  `json:"likes,omitempty"`
	LikesCount *int              `json:"likes_count,omitempty"`
	Emojis     []MessageReaction `json:"emojis,omitempty"`
}

type MessageReaction struct {
	Timestamp       time.Time `json:"timestamp"`
	ClientContext    string    `json:"client_context,omitempty"`
	SenderID         int       `json:"sender_id"`
	Emoji            string    `json:"emoji"`
	SuperReactType   string    `json:"super_react_type"`
}

type ReplyMessage struct {
	ID              string        `json:"id"`
	UserID          string        `json:"user_id,omitempty"`
	Timestamp       time.Time     `json:"timestamp"`
	ItemType        string        `json:"item_type,omitempty"`
	IsSentByViewer  *bool         `json:"is_sent_by_viewer,omitempty"`
	IsSHHMode       *bool         `json:"is_shh_mode,omitempty"`
	Text            string        `json:"text,omitempty"`
	Link            map[string]any `json:"link,omitempty"`
	AnimatedMedia   map[string]any `json:"animated_media,omitempty"`
	Media           *DirectMedia  `json:"media,omitempty"`
	VisualMedia     *VisualMedia  `json:"visual_media,omitempty"`
	MediaShare      *Media        `json:"media_share,omitempty"`
	ReelShare       map[string]any `json:"reel_share,omitempty"`
	StoryShare      map[string]any `json:"story_share,omitempty"`
	FelixShare      map[string]any `json:"felix_share,omitempty"`
	XMAShare        *MediaXma     `json:"xma_share,omitempty"`
	GenericXMA      []MediaXma    `json:"generic_xma,omitempty"`
	RawXMA          map[string]any `json:"raw_xma,omitempty"`
	Clip            *Media        `json:"clip,omitempty"`
	Placeholder     map[string]any `json:"placeholder,omitempty"`
}

type MessageLink struct {
	Text         string      `json:"text"`
	LinkContext  LinkContext `json:"link_context"`
	ClientContext string     `json:"client_context,omitempty"`
	MutationToken string     `json:"mutation_token,omitempty"`
}

type LinkContext struct {
	LinkURL       string `json:"link_url"`
	LinkTitle     string `json:"link_title,omitempty"`
	LinkSummary   string `json:"link_summary,omitempty"`
	LinkImageURL  string `json:"link_image_url,omitempty"`
}

type DirectMedia struct {
	ID           string     `json:"id"`
	MediaType    int        `json:"media_type"`
	User         *UserShort `json:"user,omitempty"`
	ThumbnailURL string     `json:"thumbnail_url,omitempty"`
	VideoURL     string     `json:"video_url,omitempty"`
	AudioURL     string     `json:"audio_url,omitempty"`
}

type VisualMedia struct {
	Media                       VisualMediaContent       `json:"media"`
	SeenUserIDs                 []string                 `json:"seen_user_ids"`
	SeenCount                   int                      `json:"seen_count"`
	ViewMode                    string                   `json:"view_mode"`
	ReplayExpiringAtUs          *int                     `json:"replay_expiring_at_us,omitempty"`
	ReplyType                   string                   `json:"reply_type,omitempty"`
	URLExpireAtSecs             *int                     `json:"url_expire_at_secs,omitempty"`
	PlaybackDurationSecs        *int                     `json:"playback_duration_secs,omitempty"`
	ExpiringMediaActionSummary  *ExpiringMediaActionSummary `json:"expiring_media_action_summary,omitempty"`
}

type VisualMediaContent struct {
	MediaType            int                        `json:"media_type"`
	ID                   string                     `json:"id,omitempty"`
	MediaID              *int                       `json:"media_id,omitempty"`
	ImageVersions2       *DirectMessageImageVersions `json:"image_versions2,omitempty"`
	VideoVersions        []VideoVersion              `json:"video_versions,omitempty"`
	OriginalWidth        *int                       `json:"original_width,omitempty"`
	OriginalHeight       *int                       `json:"original_height,omitempty"`
	User                 *VisualMediaUser           `json:"user,omitempty"`
	VideoDuration        *int                       `json:"video_duration,omitempty"`
}

type DirectMessageImageVersions struct {
	Candidates []DirectMessageImageCandidate `json:"candidates"`
}

type DirectMessageImageCandidate struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	URL          string `json:"url"`
	ScansProfile string `json:"scans_profile,omitempty"`
}

type VideoVersion struct {
	ID       string `json:"id,omitempty"`
	Type     *int   `json:"type,omitempty"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	URL      string `json:"url"`
	Bandwidth *int  `json:"bandwidth,omitempty"`
}

type VisualMediaUser struct {
	ID                         string          `json:"id"`
	PK                         int             `json:"pk"`
	PKID                       string          `json:"pk_id"`
	FullName                   string          `json:"full_name"`
	Username                   string          `json:"username"`
	ProfilePicURL              string          `json:"profile_pic_url"`
	IsVerified                 bool            `json:"is_verified"`
	IsPrivate                  bool            `json:"is_private"`
	FriendshipStatus           *FriendshipStatus `json:"friendship_status,omitempty"`
}

type FriendshipStatus struct {
	Blocking                       bool `json:"blocking"`
	IsMessagingOnlyBlocking        bool `json:"is_messaging_only_blocking"`
	IsMessagingPseudoBlocking      bool `json:"is_messaging_pseudo_blocking"`
	IsUnavailable                  bool `json:"is_unavailable"`
}

type ExpiringMediaActionSummary struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

type DirectResponse struct {
	UnseenCount   *int    `json:"unseen_count,omitempty"`
	UnseenCountTS *int    `json:"unseen_count_ts,omitempty"`
	Status        string  `json:"status,omitempty"`
}

type DirectShortThread struct {
	ID          string      `json:"id"`
	Users       []UserShort `json:"users"`
	Named       bool        `json:"named"`
	ThreadTitle string      `json:"thread_title"`
	Pending     bool        `json:"pending"`
	ThreadType  string      `json:"thread_type"`
	ViewerID    string      `json:"viewer_id"`
	IsGroup     bool        `json:"is_group"`
}

type MediaXma struct {
	VideoURL           string  `json:"video_url"`
	Title              string  `json:"title,omitempty"`
	PreviewURL         string  `json:"preview_url,omitempty"`
	PreviewURLMimeType string  `json:"preview_url_mime_type,omitempty"`
	HeaderIconURL      string  `json:"header_icon_url,omitempty"`
	HeaderIconWidth    *int    `json:"header_icon_width,omitempty"`
	HeaderIconHeight   *int    `json:"header_icon_height,omitempty"`
	HeaderTitleText    string  `json:"header_title_text,omitempty"`
	PreviewMediaFbid   string  `json:"preview_media_fbid,omitempty"`
}
