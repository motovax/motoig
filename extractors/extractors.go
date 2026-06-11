// Package extractors transforms raw Instagram API JSON into typed models.
package extractors

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/motovax/motoig/models"
)

func ExtractUserShort(data map[string]any) models.UserShort {
	u := models.UserShort{
		PK:                getString(data, "id", "pk"),
		Username:          getString(data, "username"),
		FullName:          getString(data, "full_name"),
		ProfilePicURL:     getString(data, "profile_pic_url"),
		ProfilePicURLHD:   getString(data, "profile_pic_url_hd"),
	}
	if v, ok := data["is_private"].(bool); ok {
		u.IsPrivate = &v
	}
	if v, ok := data["is_verified"].(bool); ok {
		u.IsVerified = &v
	}
	if v, ok := data["latest_reel_media"].(float64); ok {
		n := int(v)
		u.LatestReelMedia = &n
	}
	return u
}

func ExtractUser(data map[string]any) models.User {
	u := models.User{
		PK:                getString(data, "id", "pk"),
		Username:          getString(data, "username"),
		FullName:          getString(data, "full_name"),
		IsPrivate:         getBool(data, "is_private"),
		ProfilePicURL:     getString(data, "profile_pic_url"),
		IsVerified:        getBool(data, "is_verified"),
		MediaCount:        getInt(data, "media_count"),
		FollowerCount:     getInt(data, "follower_count"),
		FollowingCount:    getInt(data, "following_count"),
		Biography:         getString(data, "biography"),
		ExternalURL:       getString(data, "external_url"),
		IsBusiness:        getBool(data, "is_business"),
		PublicEmail:       getString(data, "public_email", "business_email"),
		ContactPhoneNumber: getString(data, "contact_phone_number", "business_phone_number"),
		BusinessCategoryName: getString(data, "business_category_name"),
		CategoryName:      getString(data, "category_name"),
		Category:          getString(data, "category"),
		AddressStreet:     getString(data, "address_street"),
		CityID:            getString(data, "city_id"),
		CityName:          getString(data, "city_name"),
		Zip:               getString(data, "zip"),
		InteropMessagingUserFBID: getString(data, "interop_messaging_user_fbid"),
	}

	if v, ok := data["account_type"].(float64); ok {
		n := int(v)
		u.AccountType = &n
	}

	if hd, ok := data["hd_profile_pic_versions"].([]any); ok && len(hd) > 0 {
		if last, ok := hd[len(hd)-1].(map[string]any); ok {
			u.ProfilePicURLHD = getString(last, "url")
		}
	} else if info, ok := data["hd_profile_pic_url_info"].(map[string]any); ok {
		u.ProfilePicURLHD = getString(info, "url")
	}

	if channels, ok := data["pinned_channels_info"].(map[string]any); ok {
		if list, ok := channels["pinned_channels_list"].([]any); ok {
			for _, ch := range list {
				if chm, ok := ch.(map[string]any); ok {
					u.BroadcastChannel = append(u.BroadcastChannel, models.Broadcast{
						Title:       getString(chm, "title"),
						ThreadIGID:  getString(chm, "thread_igid"),
						Subtitle:    getString(chm, "subtitle"),
						InviteLink:  getString(chm, "invite_link"),
						IsMember:    getBool(chm, "is_member"),
					})
				}
			}
		}
	}

	return u
}

func ExtractMediaV1(data map[string]any) models.Media {
	m := models.Media{
		PK:           getString(data, "pk"),
		ID:           getString(data, "id"),
		Code:         getString(data, "code"),
		MediaType:    getMediaType(data),
		ProductType:  getString(data, "product_type"),
		CaptionText:  getNestedString(data, "caption", "text"),
		LikeCount:    getInt(data, "like_count"),
		CommentCount: getInt(data, "comment_count"),
	}

	if t, ok := data["taken_at"].(float64); ok {
		m.TakenAt.Time = time.Unix(int64(t), 0)
	} else if s, ok := data["taken_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			m.TakenAt.Time = t
		}
	}

	if uv, ok := data["user"].(map[string]any); ok {
		m.User = ExtractUserShort(uv)
	}

	if iv, ok := data["image_versions2"].(map[string]any); ok {
		m.ImageVersions2 = extractImageVersions(iv)
		if cands, ok := iv["candidates"].([]any); ok && len(cands) > 0 {
			best := cands[0]
			for _, c := range cands {
				if cm, ok := c.(map[string]any); ok {
					if bm, ok := best.(map[string]any); ok {
						if getInt(cm, "height")*getInt(cm, "width") > getInt(bm, "height")*getInt(bm, "width") {
							best = c
						}
					}
				}
			}
			if bm, ok := best.(map[string]any); ok {
				m.ThumbnailURL = getString(bm, "url")
			}
		}
	}

	if vids, ok := data["video_versions"].([]any); ok && len(vids) > 0 {
		best := vids[0]
		for _, v := range vids {
			if vm, ok := v.(map[string]any); ok {
				if bm, ok := best.(map[string]any); ok {
					if getInt(vm, "height")*getInt(vm, "width") > getInt(bm, "height")*getInt(bm, "width") {
						best = v
					}
				}
			}
		}
		if bm, ok := best.(map[string]any); ok {
			m.VideoURL = getString(bm, "url")
		}
	}

	if m.MediaType == models.MediaTypeVideo && m.ProductType == "" {
		m.ProductType = "feed"
	}

	if loc, ok := data["location"].(map[string]any); ok && loc != nil {
		locModel := ExtractLocation(loc)
		m.Location = &locModel
	}

	if tags, ok := data["usertags"].(map[string]any); ok {
		if in, ok := tags["in"].([]any); ok {
			for _, t := range in {
				if tm, ok := t.(map[string]any); ok {
					m.Usertags = append(m.Usertags, ExtractUsertag(tm))
				}
			}
		}
	}

	if sponsors, ok := data["sponsor_tags"].([]any); ok {
		for _, s := range sponsors {
			if sm, ok := s.(map[string]any); ok {
				if sponsor, ok := sm["sponsor"].(map[string]any); ok {
					m.SponsorTags = append(m.SponsorTags, ExtractUserShort(sponsor))
				}
			}
		}
	}

	if carousel, ok := data["carousel_media"].([]any); ok {
		for _, r := range carousel {
			if rm, ok := r.(map[string]any); ok {
				m.Resources = append(m.Resources, ExtractResourceV1(rm))
			}
		}
	}

	if pc, ok := data["play_count"].(float64); ok {
		n := int(pc)
		m.PlayCount = &n
	}
	if hl, ok := data["has_liked"].(bool); ok {
		m.HasLiked = &hl
	}
	if vc, ok := data["video_view_count"].(float64); ok {
		n := int(vc)
		m.ViewCount = &n
	}
	if vd, ok := data["video_duration"].(float64); ok {
		m.VideoDuration = &vd
	}

	return m
}

func ExtractResourceV1(data map[string]any) models.Resource {
	r := models.Resource{
		PK:        getString(data, "pk"),
		MediaType: getMediaType(data),
	}

	if vids, ok := data["video_versions"].([]any); ok && len(vids) > 0 {
		best := vids[0]
		for _, v := range vids {
			if vm, ok := v.(map[string]any); ok {
				if bm, ok := best.(map[string]any); ok {
					if getInt(vm, "height")*getInt(vm, "width") > getInt(bm, "height")*getInt(bm, "width") {
						best = v
					}
				}
			}
		}
		if bm, ok := best.(map[string]any); ok {
			r.VideoURL = getString(bm, "url")
		}
	}

	if iv, ok := data["image_versions2"].(map[string]any); ok {
		if cands, ok := iv["candidates"].([]any); ok && len(cands) > 0 {
			best := cands[0]
			for _, c := range cands {
				if cm, ok := c.(map[string]any); ok {
					if bm, ok := best.(map[string]any); ok {
						if getInt(cm, "height")*getInt(cm, "width") > getInt(bm, "height")*getInt(bm, "width") {
							best = c
						}
					}
				}
			}
			if bm, ok := best.(map[string]any); ok {
				r.ThumbnailURL = getString(bm, "url")
			}
		}
	}

	return r
}

func ExtractUsertag(data map[string]any) models.Usertag {
	t := models.Usertag{}
	if user, ok := data["user"].(map[string]any); ok {
		t.User = ExtractUserShort(user)
	}
	if pos, ok := data["position"].([]any); ok && len(pos) >= 2 {
		if x, ok := pos[0].(float64); ok {
			t.X = x
		}
		if y, ok := pos[1].(float64); ok {
			t.Y = y
		}
	}
	if x, ok := data["x"].(float64); ok {
		t.X = x
	}
	if y, ok := data["y"].(float64); ok {
		t.Y = y
	}
	return t
}

func ExtractLocation(data map[string]any) models.Location {
	if data == nil {
		return models.Location{}
	}
	if place, ok := data["place"].(map[string]any); ok {
		if loc, ok := place["location"].(map[string]any); ok {
			data = loc
		}
	}
	l := models.Location{
		Name:              getString(data, "name"),
		Phone:             getString(data, "phone"),
		Website:           getString(data, "website"),
		Category:          getString(data, "category"),
		Address:           getString(data, "address", "location_address"),
		City:              getString(data, "city", "location_city"),
		Zip:               getString(data, "zip", "location_zip"),
		ExternalIDSource:  getString(data, "external_id_source", "external_source"),
	}
	if pk, ok := data["id"].(float64); ok {
		n := int64(pk)
		l.PK = &n
	} else if pk, ok := data["pk"].(float64); ok {
		n := int64(pk)
		l.PK = &n
	}
	if eid, ok := data["external_id"].(float64); ok {
		n := int64(eid)
		l.ExternalID = &n
	}
	if lat, ok := data["lat"].(float64); ok {
		l.Lat = &lat
	}
	if lng, ok := data["lng"].(float64); ok {
		l.Lng = &lng
	}
	return l
}

func ExtractComment(data map[string]any) models.Comment {
	c := models.Comment{
		PK:         getString(data, "pk"),
		Text:       getString(data, "text"),
		ContentType: getString(data, "content_type"),
		Status:     getString(data, "status"),
	}
	if hl, ok := data["has_liked"].(bool); ok {
		c.HasLiked = &hl
	} else if hl, ok := data["has_liked_comment"].(bool); ok {
		c.HasLiked = &hl
	}
	if lc, ok := data["like_count"].(float64); ok {
		n := int(lc)
		c.LikeCount = &n
	} else if lc, ok := data["comment_like_count"].(float64); ok {
		n := int(lc)
		c.LikeCount = &n
	}
	if user, ok := data["user"].(map[string]any); ok {
		c.User = ExtractUserShort(user)
	}
	if t, ok := data["created_at_utc"].(float64); ok {
		c.CreatedAtUTC = time.Unix(int64(t), 0)
	} else if s, ok := data["created_at_utc"].(string); ok {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			c.CreatedAtUTC = t
		}
	}
	if rid, ok := data["replied_to_comment_id"].(string); ok {
		c.RepliedToCommentID = &rid
	}
	return c
}

func ExtractStoryV1(data map[string]any) models.Story {
	s := models.Story{
		PK:          getString(data, "pk"),
		ID:          getString(data, "id"),
		Code:        getString(data, "code"),
		ProductType: getString(data, "product_type"),
	}

	if t, ok := data["taken_at"].(float64); ok {
		s.TakenAt.Time = time.Unix(int64(t), 0)
	} else if t, ok := data["device_timestamp"].(float64); ok {
		s.TakenAt.Time = time.Unix(int64(t), 0)
	}

	if s.PK == "" {
		s.PK = fmt.Sprintf("%d", int64(getFloat(data, "pk")))
	}

	mt := getMediaType(data)
	s.MediaType = mt
	if mt == models.MediaTypeVideo && s.ProductType == "" {
		s.ProductType = "story"
	}

	if iv, ok := data["image_versions2"].(map[string]any); ok {
		if cands, ok := iv["candidates"].([]any); ok && len(cands) > 0 {
			best := cands[0]
			for _, c := range cands {
				if cm, ok := c.(map[string]any); ok {
					if bm, ok := best.(map[string]any); ok {
						if getInt(cm, "height")*getInt(cm, "width") > getInt(bm, "height")*getInt(bm, "width") {
							best = c
						}
					}
				}
			}
			if bm, ok := best.(map[string]any); ok {
				s.ThumbnailURL = getString(bm, "url")
			}
		}
	}

	if vids, ok := data["video_versions"].([]any); ok && len(vids) > 0 {
		best := vids[0]
		for _, v := range vids {
			if vm, ok := v.(map[string]any); ok {
				if bm, ok := best.(map[string]any); ok {
					if getInt(vm, "height")*getInt(vm, "width") > getInt(bm, "height")*getInt(bm, "width") {
						best = v
					}
				}
			}
		}
		if bm, ok := best.(map[string]any); ok {
			s.VideoURL = getString(bm, "url")
		}
	}

	if user, ok := data["user"].(map[string]any); ok {
		s.User = ExtractUserShort(user)
	}

	if mentions, ok := data["reel_mentions"].([]any); ok {
		for _, m := range mentions {
			if mm, ok := m.(map[string]any); ok {
				sm := models.StoryMention{}
				if user, ok := mm["user"].(map[string]any); ok {
					sm.User = ExtractUserShort(user)
				}
				s.Mentions = append(s.Mentions, sm)
			}
		}
	}

	if sponsors, ok := data["sponsor_tags"].([]any); ok {
		for _, st := range sponsors {
			if sm, ok := st.(map[string]any); ok {
				if sponsor, ok := sm["sponsor"].(map[string]any); ok {
					s.SponsorTags = append(s.SponsorTags, ExtractUserShort(sponsor))
				}
			}
		}
	}

	if s.Code == "" {
		s.Code = encodeShortcode(s.PK)
	}

	return s
}

func ExtractDirectThread(data map[string]any) models.DirectThread {
	dt := models.DirectThread{
		PK:         getString(data, "thread_v2_id"),
		ID:         getString(data, "thread_id"),
		Named:      getBool(data, "named"),
		Pending:    getBool(data, "pending"),
		Archived:   getBool(data, "archived"),
		ThreadType: getString(data, "thread_type"),
		Muted:      getBool(data, "muted"),
		IsGroup:    getBool(data, "is_group"),
	}

	if t, ok := data["last_activity_at"].(float64); ok {
		dt.LastActivityAt.Time = time.Unix(int64(t)/1e6, 0)
	}

	if users, ok := data["users"].([]any); ok {
		for _, u := range users {
			if um, ok := u.(map[string]any); ok {
				dt.Users = append(dt.Users, ExtractUserShort(um))
			}
		}
	}

	if inviter, ok := data["inviter"].(map[string]any); ok {
		u := ExtractUserShort(inviter)
		dt.Inviter = &u
	}

	if items, ok := data["items"].([]any); ok {
		for _, item := range items {
			if im, ok := item.(map[string]any); ok {
				im["thread_id"] = dt.ID
				dt.Messages = append(dt.Messages, ExtractDirectMessage(im))
			}
		}
	}

	return dt
}

func ExtractDirectMessage(data map[string]any) models.DirectMessage {
	dm := models.DirectMessage{
		ID:            getString(data, "item_id"),
		ItemType:      getString(data, "item_type"),
		Text:          getString(data, "text"),
		ClientContext: getString(data, "client_context"),
	}

	if t, ok := data["timestamp"].(float64); ok {
		dm.Timestamp = time.Unix(int64(t)/1e6, 0)
	} else if s, ok := data["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			dm.Timestamp = t
		}
	}

	if uid, ok := data["user_id"].(float64); ok {
		dm.UserID = fmt.Sprintf("%d", int64(uid))
	} else if uid, ok := data["user_id"].(string); ok {
		dm.UserID = uid
	}

	if isv, ok := data["is_sent_by_viewer"].(bool); ok {
		dm.IsSentByViewer = &isv
	}

	if media, ok := data["media"].(map[string]any); ok {
		dm.Media = extractDirectMedia(media)
	}

	if vm, ok := data["visual_media"].(map[string]any); ok {
		dm.VisualMedia = extractVisualMedia(vm)
	}

	if ms, ok := data["media_share"].(map[string]any); ok {
		m := ExtractMediaV1(ms)
		dm.MediaShare = &m
	}

	return dm
}

func extractDirectMedia(data map[string]any) *models.DirectMedia {
	dm := &models.DirectMedia{
		ID:        getString(data, "id"),
		MediaType: getInt(data, "media_type"),
	}
	if user, ok := data["user"].(map[string]any); ok {
		u := ExtractUserShort(user)
		dm.User = &u
	}
	if iv, ok := data["image_versions2"].(map[string]any); ok {
		if cands, ok := iv["candidates"].([]any); ok && len(cands) > 0 {
			best := cands[0]
			for _, c := range cands {
				if cm, ok := c.(map[string]any); ok {
					if bm, ok := best.(map[string]any); ok {
						if getInt(cm, "height")*getInt(cm, "width") > getInt(bm, "height")*getInt(bm, "width") {
							best = c
						}
					}
				}
			}
			if bm, ok := best.(map[string]any); ok {
				dm.ThumbnailURL = getString(bm, "url")
			}
		}
	}
	if vids, ok := data["video_versions"].([]any); ok && len(vids) > 0 {
		best := vids[0]
		for _, v := range vids {
			if vm, ok := v.(map[string]any); ok {
				if bm, ok := best.(map[string]any); ok {
					if getInt(vm, "height")*getInt(vm, "width") > getInt(bm, "height")*getInt(bm, "width") {
						best = v
					}
				}
			}
		}
		if bm, ok := best.(map[string]any); ok {
			dm.VideoURL = getString(bm, "url")
		}
	}
	return dm
}

func extractVisualMedia(data map[string]any) *models.VisualMedia {
	vm := &models.VisualMedia{
		ViewMode: getString(data, "view_mode"),
	}
	if media, ok := data["media"].(map[string]any); ok {
		vm.Media = models.VisualMediaContent{
			MediaType: getInt(media, "media_type"),
			ID:        getString(media, "id"),
		}
		if iv, ok := media["image_versions2"].(map[string]any); ok {
			vm.Media.ImageVersions2 = extractDMImageVersions(iv)
		}
		if vids, ok := media["video_versions"].([]any); ok {
			for _, v := range vids {
				if vm2, ok := v.(map[string]any); ok {
					vv := models.VideoVersion{
						Width:  getInt(vm2, "width"),
						Height: getInt(vm2, "height"),
						URL:    getString(vm2, "url"),
					}
					vm.Media.VideoVersions = append(vm.Media.VideoVersions, vv)
				}
			}
		}
	}
	if sc, ok := data["seen_count"].(float64); ok {
		vm.SeenCount = int(sc)
	}
	return vm
}

func extractDMImageVersions(data map[string]any) *models.DirectMessageImageVersions {
	iv := &models.DirectMessageImageVersions{}
	if cands, ok := data["candidates"].([]any); ok {
		for _, c := range cands {
			if cm, ok := c.(map[string]any); ok {
				iv.Candidates = append(iv.Candidates, models.DirectMessageImageCandidate{
					Width:  getInt(cm, "width"),
					Height: getInt(cm, "height"),
					URL:    getString(cm, "url"),
				})
			}
		}
	}
	return iv
}

func extractImageVersions(data map[string]any) *models.SharedMediaImageVersions {
	iv := &models.SharedMediaImageVersions{}
	if cands, ok := data["candidates"].([]any); ok {
		for _, c := range cands {
			if cm, ok := c.(map[string]any); ok {
				iv.Candidates = append(iv.Candidates, models.SharedMediaImageCandidate{
					Width:  getInt(cm, "width"),
					Height: getInt(cm, "height"),
					URL:    getString(cm, "url"),
				})
			}
		}
	}
	return iv
}

func ExtractCollection(data map[string]any) models.Collection {
	c := models.Collection{
		Name:       getString(data, "collection_name"),
		Type:       getString(data, "collection_type"),
		MediaCount: getInt(data, "collection_media_count"),
	}
	c.ID = getString(data, "collection_id", "id")
	return c
}

func ExtractHashtagV1(data map[string]any) models.Hashtag {
	h := models.Hashtag{
		ID:            getString(data, "id"),
		Name:          getString(data, "name"),
		ProfilePicURL: getString(data, "profile_pic_url"),
	}
	if mc, ok := data["media_count"].(float64); ok {
		n := int(mc)
		h.MediaCount = &n
	}
	return h
}

func ExtractHighlightV1(data map[string]any) models.Highlight {
	h := models.Highlight{
		ID:          getString(data, "id"),
		Title:       getString(data, "title"),
		MediaCount:  getInt(data, "media_count"),
	}
	if id, ok := data["id"].(string); ok {
		parts := splitHighlightID(id)
		h.PK = parts
	}
	if user, ok := data["user"].(map[string]any); ok {
		h.User = ExtractUserShort(user)
	}
	if items, ok := data["items"].([]any); ok {
		for _, item := range items {
			if im, ok := item.(map[string]any); ok {
				h.Items = append(h.Items, ExtractStoryV1(im))
			}
		}
	}
	return h
}

func splitHighlightID(id string) string {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == ':' {
			return id[i+1:]
		}
	}
	return id
}

func encodeShortcode(pk string) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var n uint64
	fmt.Sscanf(pk, "%d", &n)
	if n == 0 {
		return ""
	}
	result := ""
	for n > 0 {
		result = string(alphabet[n%64]) + result
		n /= 64
	}
	return result
}

func getMediaType(data map[string]any) models.MediaType {
	if v, ok := data["media_type"].(float64); ok {
		return models.MediaType(int(v))
	}
	return models.MediaTypeImage
}

func getString(data map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := data[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func getNestedString(data map[string]any, keys ...string) string {
	current := data
	for i, key := range keys {
		if i == len(keys)-1 {
			return getString(current, key)
		}
		if v, ok := current[key].(map[string]any); ok {
			current = v
		} else {
			return ""
		}
	}
	return ""
}

func getBool(data map[string]any, key string) bool {
	if v, ok := data[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getInt(data map[string]any, keys ...string) int {
	for _, key := range keys {
		if v, ok := data[key]; ok {
			if f, ok := v.(float64); ok {
				return int(f)
			}
			if i, ok := v.(int); ok {
				return i
			}
		}
	}
	return 0
}

func getFloat(data map[string]any, key string) float64 {
	if v, ok := data[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func ParseJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
