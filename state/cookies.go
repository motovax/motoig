package state

import (
	"net/http"
	"net/url"
)

// SetCookie sets a cookie on the Instagram domain.
func (s *State) SetCookie(name, value string) {
	if s.Jar == nil {
		return
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
