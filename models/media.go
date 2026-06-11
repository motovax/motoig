package models

type Media struct {
	PK                           string                `json:"pk"`
	ID                           string                `json:"id"`
	Code                         string                `json:"code"`
	TakenAt                      Timestamp             `json:"taken_at"`
	MediaType                    MediaType             `json:"media_type"`
	ImageVersions2               *SharedMediaImageVersions `json:"image_versions2,omitempty"`
	ProductType                  string                `json:"product_type,omitempty"`
	ThumbnailURL                 string                `json:"thumbnail_url,omitempty"`
	Location                     *Location             `json:"location,omitempty"`
	User                         UserShort             `json:"user"`
	CommentCount                 int                   `json:"comment_count"`
	CommentsDisabled             bool                  `json:"comments_disabled"`
	CommentingDisabledForViewer  bool                  `json:"commenting_disabled_for_viewer"`
	LikeCount                    int                   `json:"like_count"`
	PlayCount                    *int                  `json:"play_count,omitempty"`
	HasLiked                     *bool                 `json:"has_liked,omitempty"`
	CaptionText                  string                `json:"caption_text"`
	AccessibilityCaption         string                `json:"accessibility_caption,omitempty"`
	Usertags                     []Usertag             `json:"usertags"`
	SponsorTags                  []UserShort           `json:"sponsor_tags"`
	VideoURL                     string                `json:"video_url,omitempty"`
	ViewCount                    *int                  `json:"view_count,omitempty"`
	VideoDuration                *float64              `json:"video_duration,omitempty"`
	Title                        string                `json:"title,omitempty"`
	Resources                    []Resource            `json:"resources"`
	ClipsMetadata                *ClipsMetadata        `json:"clips_metadata,omitempty"`
}

type Resource struct {
	PK           string  `json:"pk"`
	VideoURL     string  `json:"video_url,omitempty"`
	ThumbnailURL string  `json:"thumbnail_url,omitempty"`
	MediaType    MediaType `json:"media_type"`
	Usertags     []Usertag `json:"usertags,omitempty"`
}

type SharedMediaImageVersions struct {
	AdditionalCandidates          *AdditionalCandidates          `json:"additional_candidates,omitempty"`
	Candidates                    []SharedMediaImageCandidate    `json:"candidates"`
	ScrubberSpritesheetInfoCandidates *ScrubberSpritesheetInfoCandidates `json:"scrubber_spritesheet_info_candidates,omitempty"`
}

type SharedMediaImageCandidate struct {
	EstimatedScansSizes []int  `json:"estimated_scans_sizes,omitempty"`
	Height              int    `json:"height"`
	ScansProfile        string `json:"scans_profile,omitempty"`
	URL                 string `json:"url"`
	Width               int    `json:"width"`
	IsSpatialImage      *bool  `json:"is_spatial_image,omitempty"`
}

type AdditionalCandidates struct {
	FirstFrame       *SharedMediaImageCandidate `json:"first_frame,omitempty"`
	IgtvFirstFrame   *SharedMediaImageCandidate `json:"igtv_first_frame,omitempty"`
	SmartFrame       *SharedMediaImageCandidate `json:"smart_frame,omitempty"`
}

type ScrubberSpritesheetInfoCandidates struct {
	Default ScrubberSpritesheetInfo `json:"default"`
}

type ScrubberSpritesheetInfo struct {
	FileSizeKB                int      `json:"file_size_kb"`
	MaxThumbnailsPerSprite    int      `json:"max_thumbnails_per_sprite"`
	RenderedWidth             int      `json:"rendered_width"`
	SpriteHeight              int      `json:"sprite_height"`
	SpriteURLs                []string `json:"sprite_urls"`
	SpriteWidth               int      `json:"sprite_width"`
	ThumbnailDuration         float64  `json:"thumbnail_duration"`
	ThumbnailHeight           int      `json:"thumbnail_height"`
	ThumbnailWidth            int      `json:"thumbnail_width"`
	ThumbnailsPerRow          int      `json:"thumbnails_per_row"`
	TotalThumbnailNumPerSprite int     `json:"total_thumbnail_num_per_sprite"`
	VideoLength               float64  `json:"video_length"`
}

type ClipsMetadata struct {
	ClipsCreationEntryPoint     string                    `json:"clips_creation_entry_point"`
	FeaturedLabel               string                    `json:"featured_label,omitempty"`
	AchievementsInfo            *ClipsAchievementsInfo    `json:"achievements_info,omitempty"`
	AdditionalAudioInfo         *ClipsAdditionalAudioInfo `json:"additional_audio_info,omitempty"`
	AudioRankingInfo            *ClipsAudioRankingInfo    `json:"audio_ranking_info,omitempty"`
	AudioType                  string                    `json:"audio_type"`
	BrandedContentTagInfo       *ClipsBrandedContentTagInfo `json:"branded_content_tag_info,omitempty"`
	ContentAppreciationInfo     *ClipsContentAppreciationInfo `json:"content_appreciation_info,omitempty"`
	MashupInfo                 *ClipsMashupInfo          `json:"mashup_info,omitempty"`
	MusicCanonicalID           string                    `json:"music_canonical_id"`
	OriginalSoundInfo          *ClipsOriginalSoundInfo   `json:"original_sound_info,omitempty"`
}

type ClipsAchievementsInfo struct {
	NumEarnedAchievements *int  `json:"num_earned_achievements,omitempty"`
	ShowAchievements      bool  `json:"show_achievements"`
}

type ClipsAdditionalAudioInfo struct {
	AdditionalAudioUsername string `json:"additional_audio_username,omitempty"`
	AudioReattributionInfo  *AudioReattributionInfo `json:"audio_reattribution_info,omitempty"`
}

type AudioReattributionInfo struct {
	ShouldAllowRestore bool `json:"should_allow_restore"`
}

type ClipsAudioRankingInfo struct {
	BestAudioClusterID string `json:"best_audio_cluster_id"`
}

type ClipsBrandedContentTagInfo struct {
	CanAddTag bool `json:"can_add_tag"`
}

type ClipsContentAppreciationInfo struct {
	Enabled                 bool   `json:"enabled"`
	EntryPointContainer     string `json:"entry_point_container,omitempty"`
}

type ClipsMashupInfo struct {
	CanToggleMashupsAllowed              bool   `json:"can_toggle_mashups_allowed"`
	FormattedMashupsCount                string `json:"formatted_mashups_count,omitempty"`
	HasBeenMashedUp                      bool   `json:"has_been_mashed_up"`
	HasNonmimicableAdditionalAudio       bool   `json:"has_nonmimicable_additional_audio"`
	IsCreatorRequestingMashup            bool   `json:"is_creator_requesting_mashup"`
	IsLightWeightCheck                   bool   `json:"is_light_weight_check"`
	IsPivotPageAvailable                 bool   `json:"is_pivot_page_available"`
	IsReuseAllowed                       bool   `json:"is_reuse_allowed"`
	MashupType                           string `json:"mashup_type,omitempty"`
	MashupsAllowed                       bool   `json:"mashups_allowed"`
	NonPrivacyFilteredMashupsMediaCount  int    `json:"non_privacy_filtered_mashups_media_count"`
}

type ClipsOriginalSoundInfo struct {
	AllowCreatorToRename              bool                     `json:"allow_creator_to_rename"`
	AudioAssetID                      int                      `json:"audio_asset_id"`
	CanRemixBeSharedToFB              bool                     `json:"can_remix_be_shared_to_fb"`
	DashManifest                      string                   `json:"dash_manifest"`
	DurationInMs                      int                      `json:"duration_in_ms"`
	IsAudioAutomaticallyAttributed    bool                     `json:"is_audio_automatically_attributed"`
	IsOriginalAudioDownloadEligible   bool                     `json:"is_original_audio_download_eligible"`
	OriginalAudioTitle                string                   `json:"original_audio_title"`
	OriginalMediaID                   int                      `json:"original_media_id"`
	ProgressiveDownloadURL            string                   `json:"progressive_download_url"`
	TimeCreated                       int                      `json:"time_created"`
	IgArtist                          *ClipsIgArtist           `json:"ig_artist,omitempty"`
	ConsumptionInfo                   *ClipsConsumptionInfo    `json:"consumption_info,omitempty"`
}

type ClipsIgArtist struct {
	PK              int    `json:"pk"`
	PKID            string `json:"pk_id"`
	ID              string `json:"id"`
	Username        string `json:"username"`
	FullName        string `json:"full_name"`
	IsPrivate       bool   `json:"is_private"`
	IsVerified      bool   `json:"is_verified"`
	ProfilePicURL   string `json:"profile_pic_url"`
	StrongID        string `json:"strong_id__"`
}

type ClipsConsumptionInfo struct {
	DisplayMediaID            string `json:"display_media_id,omitempty"`
	IsBookmarked              bool   `json:"is_bookmarked"`
	IsTrendingInClips         bool   `json:"is_trending_in_clips"`
	ShouldMuteAudioReason     string `json:"should_mute_audio_reason"`
}

type MediaOembed struct {
	Title            string `json:"title"`
	AuthorName       string `json:"author_name"`
	AuthorURL        string `json:"author_url"`
	AuthorID         string `json:"author_id"`
	MediaID          string `json:"media_id"`
	ProviderName     string `json:"provider_name"`
	ProviderURL      string `json:"provider_url"`
	Type             string `json:"type"`
	Width            *int   `json:"width,omitempty"`
	Height           *int   `json:"height,omitempty"`
	HTML             string `json:"html"`
	ThumbnailURL     string `json:"thumbnail_url"`
	ThumbnailWidth   int    `json:"thumbnail_width"`
	ThumbnailHeight  int    `json:"thumbnail_height"`
	CanView          bool   `json:"can_view"`
}
