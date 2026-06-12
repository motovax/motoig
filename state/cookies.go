package state

import (
	"net/http"
	"net/url"
	"strings"
)

func normalizeSessionID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if decoded, err := url.QueryUnescape(value); err == nil && decoded != "" {
		return decoded
	}
	return value
}

// NormalizeAuthorizationData decodes URL-encoded sessionid values in stored auth data.
func NormalizeAuthorizationData(authData map[string]any) {
	if authData == nil {
		return
	}
	if v, ok := authData["sessionid"].(string); ok {
		authData["sessionid"] = normalizeSessionID(v)
	}
}

// SetCookie sets a cookie on the Instagram domain.
func (s *State) SetCookie(name, value string) {
	if s.Jar == nil {
		return
	}
	if name == "sessionid" {
		value = normalizeSessionID(value)
	}
	u, _ := url.Parse("https://i.instagram.com/")
	s.Jar.SetCookies(u, []*http.Cookie{
		{
			Name:     name,
			Value:    value,
			Domain:   "i.instagram.com",
			Path:     "/",
			HttpOnly: true,
		},
	})
}

// SetCookiesFromDict sets multiple cookies from a map.
func (s *State) SetCookiesFromDict(cookies map[string]any) {
	for name, val := range cookies {
		if str, ok := val.(string); ok {
			s.SetCookie(name, str)
		}
	}
}
