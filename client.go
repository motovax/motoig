// Package motoig provides an unofficial Instagram client library.
//
// motoig is a Go port of instagrapi (https://github.com/subzeroid/instagrapi).
// See github.com/motovax/motoig/references/instagrapi for symbol mappings.
package motoig

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/url"
	"strings"
	"time"

	igerr "github.com/motovax/motoig/errors"
	"github.com/motovax/motoig/extractors"
	"github.com/motovax/motoig/models"
	"github.com/motovax/motoig/realtime"
	"github.com/motovax/motoig/state"
)

// Client is the main entry point matching instagrapi.Client.
type Client struct {
	state    *state.State
	log      *slog.Logger
	realtime *realtime.RealtimeClient

	username string
	password string
	settings map[string]any
}

// Option configures Client construction.
type Option func(*clientConfig)

type clientConfig struct {
	userAgent      string
	proxyURL       string
	log            *slog.Logger
	delayRange     []time.Duration
	requestTimeout time.Duration
	tlsVerify      bool
}

// WithUserAgent sets a custom user agent.
func WithUserAgent(ua string) Option {
	return func(c *clientConfig) { c.userAgent = ua }
}

// WithProxy sets an HTTP/SOCKS proxy URL.
func WithProxy(url string) Option {
	return func(c *clientConfig) { c.proxyURL = url }
}

// WithLogger sets the structured logger.
func WithLogger(log *slog.Logger) Option {
	return func(c *clientConfig) { c.log = log }
}

// WithDelayRange sets the delay range between requests.
func WithDelayRange(min, max time.Duration) Option {
	return func(c *clientConfig) { c.delayRange = []time.Duration{min, max} }
}

// WithRequestTimeout sets the request timeout.
func WithRequestTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.requestTimeout = d }
}

// New creates a new Instagram client.
func New(opts ...Option) *Client {
	cfg := clientConfig{
		log: slog.Default(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	s := state.New(state.Options{
		UserAgent:      cfg.userAgent,
		ProxyURL:       cfg.proxyURL,
		DelayRange:     cfg.delayRange,
		RequestTimeout: cfg.requestTimeout,
	})

	return &Client{
		state:    s,
		log:      cfg.log,
		settings: make(map[string]any),
	}
}

func newClient(st *state.State, cfg clientConfig) *Client {
	return &Client{
		state:    st,
		log:      cfg.log,
		settings: make(map[string]any),
	}
}

// State returns the underlying session state.
func (c *Client) State() *state.State { return c.state }

// UserID returns the authenticated user's ID.
func (c *Client) UserID() string { return c.state.UserID }

// Username returns the authenticated user's username.
func (c *Client) Username() string { return c.state.Username }

// SetSessionID logs in using a session ID from browser cookies.
//
// Port of instagrapi session bootstrap via sessionid cookie.
// Reference: https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/auth.py
func (c *Client) SetSessionID(ctx context.Context, sessionID string) error {
	if decoded, err := url.QueryUnescape(strings.TrimSpace(sessionID)); err == nil && decoded != "" {
		sessionID = decoded
	}
	if len(sessionID) < 30 {
		return igerr.New("SetSessionID", "invalid sessionid")
	}
	ensureSessionJar(c)
	c.state.SetCookie("sessionid", sessionID)

	userID := ""
	for i := 0; i < len(sessionID); i++ {
		if sessionID[i] >= '0' && sessionID[i] <= '9' {
			userID += string(sessionID[i])
		} else {
			break
		}
	}
	if userID == "" {
		return igerr.New("SetSessionID", "invalid sessionid: no user id")
	}

	c.state.UserID = userID
	c.state.LoggedIn = true
	c.state.AuthorizationData = map[string]any{
		"ds_user_id":                      userID,
		"sessionid":                       sessionID,
		"should_use_header_over_cookies": true,
	}
	c.state.Authorization = c.state.BuildAuthorization()

	profile, err := c.UserInfoV1(ctx, userID)
	if err != nil {
		c.log.Warn("failed to fetch user info, continuing anyway", "error", err)
	} else {
		c.state.Username = profile.Username
	}

	return nil
}

// ImportBrowserCookies applies cookies exported from a browser profile (e.g. motosocial).
func (c *Client) ImportBrowserCookies(ctx context.Context, cookies map[string]string) error {
	if len(cookies) == 0 {
		return igerr.New("ImportBrowserCookies", "no cookies provided")
	}
	ensureSessionJar(c)
	for name, value := range cookies {
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" || value == "" {
			continue
		}
		c.state.SetCookie(name, value)
	}
	c.state.EnsureLoggedInFromCookies()
	if c.state.UserID == "" {
		return igerr.New("ImportBrowserCookies", "invalid sessionid: no user id")
	}
	c.state.LoggedIn = true

	profile, err := c.UserInfoV1(ctx, c.state.UserID)
	if err != nil {
		c.log.Warn("failed to fetch user info, continuing anyway", "error", err)
	} else {
		c.state.Username = profile.Username
	}
	return nil
}

func (c *Client) privateRequest(ctx context.Context, endpoint string, data map[string]any) (map[string]any, error) {
	c.state.RandomDelay()
	c.state.ReqCounter++

	authorization := c.state.GetAuthorization()

	if authorization != "" && data == nil {
		data = make(map[string]any)
	}

	return c.state.PrivateRequest(ctx, endpoint, data, nil)
}

func (c *Client) privateRequestNoDelay(ctx context.Context, endpoint string, data map[string]any) (map[string]any, error) {
	authorization := c.state.GetAuthorization()

	if authorization != "" && data == nil {
		data = make(map[string]any)
	}

	return c.state.PrivateRequest(ctx, endpoint, data, nil)
}

func (c *Client) withDefaultData(data map[string]any) map[string]any {
	if data == nil {
		data = make(map[string]any)
	}
	data["_uuid"] = c.state.UUIDs["uuid"]
	data["device_id"] = c.state.UUIDs["android_device_id"]
	return data
}

func (c *Client) withExtraData(data map[string]any) map[string]any {
	data = c.withDefaultData(data)
	data["phone_id"] = c.state.UUIDs["phone_id"]
	data["_uid"] = c.state.UserID
	data["guid"] = c.state.UUIDs["uuid"]
	return data
}

// UserInfoV1 fetches user info via private API.
func (c *Client) UserInfoV1(ctx context.Context, pk string) (models.User, error) {
	data := c.withDefaultData(nil)
	result, err := c.privateRequest(ctx, "users/"+pk+"/info/", data)
	if err != nil {
		return models.User{}, err
	}
	user, ok := result["user"].(map[string]any)
	if !ok {
		return models.User{}, igerr.New("UserInfoV1", "user not found in response")
	}
	return extractors.ExtractUser(user), nil
}

// UserInfo fetches user info by PK (tries v1 then GQL).
func (c *Client) UserInfo(ctx context.Context, pk any) (models.User, error) {
	id := fmt.Sprintf("%v", pk)
	return c.UserInfoV1(ctx, id)
}

// UserInfoByUsername fetches user info by username.
func (c *Client) UserInfoByUsername(ctx context.Context, username string) (models.User, error) {
	data := c.withDefaultData(nil)
	data["username"] = username
	result, err := c.privateRequest(ctx, "users/"+username+"/usernameinfo/", data)
	if err != nil {
		return models.User{}, err
	}
	user, ok := result["user"].(map[string]any)
	if !ok {
		return models.User{}, igerr.New("UserInfoByUsername", "user not found in response")
	}
	return extractors.ExtractUser(user), nil
}

// UserMedias returns media items for a user.
func (c *Client) UserMedias(ctx context.Context, pk any, amount int) ([]models.Media, error) {
	id := fmt.Sprintf("%v", pk)
	if amount <= 0 {
		amount = 12
	}

	var medias []models.Media
	maxID := ""
	for len(medias) < amount {
		data := c.withDefaultData(nil)
		data["user_id"] = id
		data["include_feed"] = "true"
		data["count"] = fmt.Sprintf("%d", min(12, amount-len(medias)))
		if maxID != "" {
			data["max_id"] = maxID
		}

		result, err := c.privateRequest(ctx, "feed/user/"+id+"/carousel_media/", data)
		if err != nil {
			break
		}

		items, ok := result["items"].([]any)
		if !ok || len(items) == 0 {
			break
		}

		for _, item := range items {
			if m, ok := item.(map[string]any); ok {
				medias = append(medias, extractors.ExtractMediaV1(m))
			}
		}

		if mid, ok := result["next_max_id"].(string); ok && mid != "" {
			maxID = mid
		} else {
			break
		}
	}

	if len(medias) > amount {
		medias = medias[:amount]
	}
	return medias, nil
}

// MediaInfo fetches media info by PK.
func (c *Client) MediaInfo(ctx context.Context, pk any) (models.Media, error) {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	result, err := c.privateRequest(ctx, "media/"+id+"/info/", data)
	if err != nil {
		return models.Media{}, err
	}
	return extractors.ExtractMediaV1(result), nil
}

// MediaLike likes a media item.
func (c *Client) MediaLike(ctx context.Context, pk any) error {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "media/"+id+"/like/", data)
	return err
}

// MediaUnlike unlikes a media item.
func (c *Client) MediaUnlike(ctx context.Context, pk any) error {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "media/"+id+"/unlike/", data)
	return err
}

// MediaDelete deletes a media item.
func (c *Client) MediaDelete(ctx context.Context, pk any) error {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "media/"+id+"/delete/", data)
	return err
}

// MediaComments returns comments for a media item.
func (c *Client) MediaComments(ctx context.Context, pk any) ([]models.Comment, error) {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	result, err := c.privateRequest(ctx, "media/"+id+"/comments/", data)
	if err != nil {
		return nil, err
	}
	var comments []models.Comment
	if items, ok := result["comments"].([]any); ok {
		for _, item := range items {
			if cm, ok := item.(map[string]any); ok {
				comments = append(comments, extractors.ExtractComment(cm))
			}
		}
	}
	return comments, nil
}

// MediaComment posts a comment on a media item.
func (c *Client) MediaComment(ctx context.Context, pk any, text string) error {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	data["comment_text"] = text
	data["idempotency_key"] = fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
	_, err := c.privateRequest(ctx, "media/"+id+"/comment/", data)
	return err
}

// CommentDelete deletes a comment.
func (c *Client) CommentDelete(ctx context.Context, mediaPK, commentID any) error {
	mid := fmt.Sprintf("%v", mediaPK)
	cid := fmt.Sprintf("%v", commentID)
	data := c.withDefaultData(nil)
	data["comment_id"] = cid
	_, err := c.privateRequest(ctx, "media/"+mid+"/comment/"+cid+"/delete/", data)
	return err
}

// CommentLike likes a comment.
func (c *Client) CommentLike(ctx context.Context, commentID any) error {
	id := fmt.Sprintf("%v", commentID)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "media/"+id+"/comment_like/", data)
	return err
}

// CommentUnlike unlikes a comment.
func (c *Client) CommentUnlike(ctx context.Context, commentID any) error {
	id := fmt.Sprintf("%v", commentID)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "media/"+id+"/comment_unlike/", data)
	return err
}

// UserFollowers returns the followers of a user.
func (c *Client) UserFollowers(ctx context.Context, pk any, amount int) ([]models.UserShort, error) {
	id := fmt.Sprintf("%v", pk)
	if amount <= 0 {
		amount = 20
	}

	var users []models.UserShort
	rankToken := c.state.UserID + "_" + c.state.UUIDs["uuid"]
	maxID := ""

	for len(users) < amount {
		data := c.withDefaultData(nil)
		data["rank_token"] = rankToken
		data["query"] = ""
		data["count"] = fmt.Sprintf("%d", min(20, amount-len(users)))
		if maxID != "" {
			data["max_id"] = maxID
		}

		endpoint := fmt.Sprintf("friendships/%s/followers/", id)
		result, err := c.privateRequest(ctx, endpoint, data)
		if err != nil {
			break
		}

		if items, ok := result["users"].([]any); ok {
			for _, item := range items {
				if um, ok := item.(map[string]any); ok {
					users = append(users, extractors.ExtractUserShort(um))
				}
			}
		}

		if mid, ok := result["next_max_id"].(string); ok && mid != "" {
			maxID = mid
		} else {
			break
		}
	}

	if len(users) > amount {
		users = users[:amount]
	}
	return users, nil
}

// UserFollowing returns the following of a user.
func (c *Client) UserFollowing(ctx context.Context, pk any, amount int) ([]models.UserShort, error) {
	id := fmt.Sprintf("%v", pk)
	if amount <= 0 {
		amount = 20
	}

	var users []models.UserShort
	rankToken := c.state.UserID + "_" + c.state.UUIDs["uuid"]
	maxID := ""

	for len(users) < amount {
		data := c.withDefaultData(nil)
		data["rank_token"] = rankToken
		data["query"] = ""
		data["count"] = fmt.Sprintf("%d", min(20, amount-len(users)))
		if maxID != "" {
			data["max_id"] = maxID
		}

		endpoint := fmt.Sprintf("friendships/%s/following/", id)
		result, err := c.privateRequest(ctx, endpoint, data)
		if err != nil {
			break
		}

		if items, ok := result["users"].([]any); ok {
			for _, item := range items {
				if um, ok := item.(map[string]any); ok {
					users = append(users, extractors.ExtractUserShort(um))
				}
			}
		}

		if mid, ok := result["next_max_id"].(string); ok && mid != "" {
			maxID = mid
		} else {
			break
		}
	}

	if len(users) > amount {
		users = users[:amount]
	}
	return users, nil
}

// UserFollow follows a user.
func (c *Client) UserFollow(ctx context.Context, pk any) error {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "friendships/"+id+"/follow/", data)
	return err
}

// UserUnfollow unfollows a user.
func (c *Client) UserUnfollow(ctx context.Context, pk any) error {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "friendships/"+id+"/unfollow/", data)
	return err
}

// UserBlock blocks a user.
func (c *Client) UserBlock(ctx context.Context, pk any) error {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "friendships/"+id+"/block/", data)
	return err
}

// UserUnblock unblocks a user.
func (c *Client) UserUnblock(ctx context.Context, pk any) error {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "friendships/"+id+"/unblock/", data)
	return err
}

// FriendshipStatus returns the friendship status with a user.
func (c *Client) FriendshipStatus(ctx context.Context, pk any) (models.Relationship, error) {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	result, err := c.privateRequest(ctx, "friendships/show/"+id+"/", data)
	if err != nil {
		return models.Relationship{}, err
	}
	r := models.Relationship{
		UserID: id,
	}
	if v, ok := result["blocking"].(bool); ok {
		r.Blocking = v
	}
	if v, ok := result["followed_by"].(bool); ok {
		r.FollowedBy = v
	}
	if v, ok := result["following"].(bool); ok {
		r.Following = v
	}
	if v, ok := result["incoming_request"].(bool); ok {
		r.IncomingRequest = v
	}
	if v, ok := result["outgoing_request"].(bool); ok {
		r.OutgoingRequest = v
	}
	if v, ok := result["is_bestie"].(bool); ok {
		r.IsBestie = v
	}
	if v, ok := result["is_private"].(bool); ok {
		r.IsPrivate = v
	}
	if v, ok := result["is_restricted"].(bool); ok {
		r.IsRestricted = v
	}
	return r, nil
}

// UserStories returns stories for a user.
func (c *Client) UserStories(ctx context.Context, pk any) ([]models.Story, error) {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	result, err := c.privateRequest(ctx, "feed/user/"+id+"/story/", data)
	if err != nil {
		return nil, err
	}
	var stories []models.Story
	if items, ok := result["items"].([]any); ok {
		for _, item := range items {
			if sm, ok := item.(map[string]any); ok {
				stories = append(stories, extractors.ExtractStoryV1(sm))
			}
		}
	}
	return stories, nil
}

// StoryDelete deletes a story.
func (c *Client) StoryDelete(ctx context.Context, pk any) error {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	_, err := c.privateRequest(ctx, "media/"+id+"/delete/", data)
	return err
}

// StoryViewers returns viewers of a story.
func (c *Client) StoryViewers(ctx context.Context, pk any) ([]models.Viewer, error) {
	id := fmt.Sprintf("%v", pk)
	data := c.withDefaultData(nil)
	result, err := c.privateRequest(ctx, "media/"+id+"/list_reel_media_viewer/", data)
	if err != nil {
		return nil, err
	}
	var viewers []models.Viewer
	if items, ok := result["users"].([]any); ok {
		for _, item := range items {
			if um, ok := item.(map[string]any); ok {
				u := extractors.ExtractUserShort(um)
				v := models.Viewer{
					UserShort: u,
				}
				viewers = append(viewers, v)
			}
		}
	}
	return viewers, nil
}

// LocationSearch searches for locations.
func (c *Client) LocationSearch(ctx context.Context, query string) ([]models.Location, error) {
	data := c.withDefaultData(nil)
	data["query"] = query
	data["rank_token"] = c.state.UserID + "_" + c.state.UUIDs["uuid"]
	result, err := c.privateRequest(ctx, "location/search/", data)
	if err != nil {
		return nil, err
	}
	var locations []models.Location
	if venues, ok := result["venues"].([]any); ok {
		for _, v := range venues {
			if vm, ok := v.(map[string]any); ok {
				locations = append(locations, extractors.ExtractLocation(vm))
			}
		}
	}
	return locations, nil
}

// HashtagInfo returns info about a hashtag.
func (c *Client) HashtagInfo(ctx context.Context, name string) (models.Hashtag, error) {
	data := c.withDefaultData(nil)
	result, err := c.privateRequest(ctx, "hashtags/"+name+"/info/", data)
	if err != nil {
		return models.Hashtag{}, err
	}
	return extractors.ExtractHashtagV1(result), nil
}

// CollectionList returns saved collections.
func (c *Client) CollectionList(ctx context.Context) ([]models.Collection, error) {
	data := c.withDefaultData(nil)
	result, err := c.privateRequest(ctx, "collections/list/", data)
	if err != nil {
		return nil, err
	}
	var collections []models.Collection
	if items, ok := result["collections"].([]any); ok {
		for _, item := range items {
			if cm, ok := item.(map[string]any); ok {
				collections = append(collections, extractors.ExtractCollection(cm))
			}
		}
	}
	return collections, nil
}

// AccountInfo returns the authenticated account info.
func (c *Client) AccountInfo(ctx context.Context) (models.User, error) {
	data := c.withDefaultData(nil)
	result, err := c.privateRequest(ctx, "accounts/current_user/?edit=true", data)
	if err != nil {
		return models.User{}, err
	}
	user, ok := result["user"].(map[string]any)
	if !ok {
		return models.User{}, igerr.New("AccountInfo", "user not found in response")
	}
	return extractors.ExtractUser(user), nil
}

// TimelineFeed returns the user's timeline feed.
func (c *Client) TimelineFeed(ctx context.Context) ([]models.Media, error) {
	data := c.withDefaultData(nil)
	data["reason"] = "pull_to_refresh"
	data["feed_view_info"] = "[]"
	data["request_id"] = c.state.UUIDs["request_id"]
	result, err := c.privateRequest(ctx, "feed/timeline/", data)
	if err != nil {
		return nil, err
	}
	var medias []models.Media
	if items, ok := result["feed_items"].([]any); ok {
		for _, item := range items {
			if im, ok := item.(map[string]any); ok {
				if media, ok := im["media_or_ad"].(map[string]any); ok {
					medias = append(medias, extractors.ExtractMediaV1(media))
				}
			}
		}
	}
	return medias, nil
}

// ReelsTrayFeed returns the reels tray feed.
func (c *Client) ReelsTrayFeed(ctx context.Context) ([]models.Media, error) {
	data := c.withDefaultData(nil)
	data["reason"] = "pull_to_refresh"
	data["tray_session_id"] = c.state.UUIDs["tray_session_id"]
	data["request_id"] = c.state.UUIDs["request_id"]
	data["supported_capabilities_new"] = mustJSON([]map[string]string{
		{"value": "119.0,120.0,121.0,122.0,123.0,124.0,125.0,126.0,127.0,128.0,129.0,130.0,131.0,132.0,133.0,134.0,135.0,136.0,137.0,138.0,139.0,140.0,141.0,142.0", "name": "SUPPORTED_SDK_VERSIONS"},
		{"value": "14", "name": "FACE_TRACKER_VERSION"},
	})
	result, err := c.privateRequest(ctx, "feed/reels_tray/", data)
	if err != nil {
		return nil, err
	}
	var medias []models.Media
	if tray, ok := result["tray"].([]any); ok {
		for _, item := range tray {
			if im, ok := item.(map[string]any); ok {
				medias = append(medias, extractors.ExtractMediaV1(im))
			}
		}
	}
	return medias, nil
}

// GetSettings returns the current session settings for persistence.
func (c *Client) GetSettings() map[string]any {
	s := c.state
	settings := map[string]any{
		"uuids": map[string]any{
			"phone_id":          s.UUIDs["phone_id"],
			"uuid":              s.UUIDs["uuid"],
			"client_session_id": s.UUIDs["client_session_id"],
			"advertising_id":    s.UUIDs["advertising_id"],
			"android_device_id": s.UUIDs["android_device_id"],
			"request_id":        s.UUIDs["request_id"],
			"tray_session_id":   s.UUIDs["tray_session_id"],
		},
		"device_settings":     s.DeviceSettings,
		"user_agent":          s.UserAgent,
		"locale":              s.Locale,
		"country":             s.Country,
		"country_code":        s.CountryCode,
		"timezone_offset":     s.TimezoneOffset,
		"timezone_name":       s.TimezoneName,
		"push_disabled":       s.PushDisabled,
		"cookies":             s.CookieDict(),
		"authorization_data":  s.AuthorizationData,
		"mid":                 s.MID,
		"ig_u_rur":            s.IgURUR,
		"ig_www_claim":        s.IgWWWClaim,
		"last_login":          s.LastLogin.Unix(),
		"request_timeout":     int(s.RequestTimeout.Seconds()),
	}
	if uid := strings.TrimSpace(s.UserID); uid != "" {
		settings["user_id"] = uid
	}
	if username := strings.TrimSpace(s.Username); username != "" {
		settings["username"] = username
	}
	return settings
}

// LoadSettings restores session from saved settings.
func (c *Client) LoadSettings(settings map[string]any) {
	c.settings = settings

	if deviceSettings, ok := settings["device_settings"].(map[string]any); ok {
		c.state.SetDeviceSettings(deviceSettings)
	}
	if uuids, ok := settings["uuids"].(map[string]any); ok {
		c.state.SetUUIDs(uuids)
	}
	if ua, ok := settings["user_agent"].(string); ok {
		c.state.UserAgent = ua
	}
	if locale, ok := settings["locale"].(string); ok {
		c.state.Locale = locale
	}
	if country, ok := settings["country"].(string); ok {
		c.state.Country = country
	}
	if cc, ok := settings["country_code"].(float64); ok {
		c.state.CountryCode = int(cc)
	}
	if tz, ok := settings["timezone_offset"].(float64); ok {
		c.state.SetTimezoneOffset(int(tz))
	}
	if mid, ok := settings["mid"].(string); ok {
		c.state.MID = mid
	}
	if rur, ok := settings["ig_u_rur"].(string); ok {
		c.state.IgURUR = rur
	}
	if claim, ok := settings["ig_www_claim"].(string); ok {
		c.state.IgWWWClaim = claim
	}
	if authData, ok := settings["authorization_data"].(map[string]any); ok {
		state.NormalizeAuthorizationData(authData)
		c.state.AuthorizationData = authData
		c.state.Authorization = c.state.BuildAuthorization()
	}
	if cookies, ok := settings["cookies"].(map[string]any); ok {
		for name, val := range cookies {
			if str, ok := val.(string); ok {
				c.state.SetCookie(name, str)
			}
		}
	}
	if uid, ok := settings["user_id"].(string); ok {
		c.state.UserID = strings.TrimSpace(uid)
	}
	if username, ok := settings["username"].(string); ok {
		c.state.Username = strings.TrimSpace(username)
	}
	c.state.EnsureLoggedInFromCookies()
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RealtimeConnect connects to Instagram's MQTToT server for real-time DMs.
//
// Port of instagrapi.RealtimeMixin.realtime_connect.
// Reference: https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/realtime.py
func (c *Client) RealtimeConnect() error {
	userID := c.state.UserIDInt()
	if userID == 0 {
		return igerr.New("RealtimeConnect", "not logged in")
	}

	phoneID := c.state.UUIDs["phone_id"]
	clientIdentifier := phoneID
	if len(clientIdentifier) > 20 {
		clientIdentifier = clientIdentifier[:20]
	}

	c.realtime = realtime.NewRealtimeClient(realtime.RealtimeConfig{
		SessionID:        c.state.SessionID(),
		UserID:           userID,
		ClientIdentifier: clientIdentifier,
		DeviceID:         phoneID,
		UserAgent:        c.state.UserAgent,
		AppVersion:       fmt.Sprintf("%v", c.state.DeviceSettings["app_version"]),
		Capabilities:     "3brTv10=",
		Locale:           c.state.Locale,
	})

	return c.realtime.Connect()
}

// RealtimeDisconnect disconnects from the MQTToT server.
func (c *Client) RealtimeDisconnect() error {
	if c.realtime == nil {
		return nil
	}
	err := c.realtime.Disconnect()
	c.realtime = nil
	return err
}

// RealtimeOn registers an event handler for real-time events.
// Events: "message", "direct", "typing", "presence", "seen", "thread_update", "error"
func (c *Client) RealtimeOn(event string, handler func(realtime.RealtimeEvent)) {
	if c.realtime == nil {
		c.realtime = c.newRealtimeClient()
	}
	c.realtime.On(event, handler)
}

func (c *Client) newRealtimeClient() *realtime.RealtimeClient {
	userID := c.state.UserIDInt()
	phoneID := c.state.UUIDs["phone_id"]
	clientIdentifier := phoneID
	if len(clientIdentifier) > 20 {
		clientIdentifier = clientIdentifier[:20]
	}
	return realtime.NewRealtimeClient(realtime.RealtimeConfig{
		SessionID:        c.state.SessionID(),
		UserID:           userID,
		ClientIdentifier: clientIdentifier,
		DeviceID:         phoneID,
		UserAgent:        c.state.UserAgent,
		AppVersion:       fmt.Sprintf("%v", c.state.DeviceSettings["app_version"]),
		Capabilities:     "3brTv10=",
		Locale:           c.state.Locale,
	})
}

// RealtimeSendDirect sends a DM via the real-time MQTT connection.
func (c *Client) RealtimeSendDirect(threadID, text, clientContext string) error {
	if c.realtime == nil || !c.realtime.IsConnected() {
		return igerr.New("RealtimeSendDirect", "not connected")
	}
	return c.realtime.SendDirectMessage(threadID, text, clientContext)
}
