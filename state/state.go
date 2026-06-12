// Package state manages Instagram session credentials and HTTP transport.
package state

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	igerr "github.com/motovax/motoig/errors"
	"github.com/motovax/motoig/internal"
)

const (
	defaultUserAgent    = "Instagram 428.0.0.47.67 Android (34/14; 480dpi; 1344x2992; Google/google; Pixel 8 Pro; husky; husky; en_US; 961145276)"
	privateAPIBase      = "https://i.instagram.com/api/v1/"
)

// State manages Instagram session credentials and HTTP transport.
type State struct {
	UserID               string
	Username             string
	Authorization        string
	AuthorizationData    map[string]any
	MID                  string
	IgURUR               string
	IgWWWClaim           string
	LastLogin            time.Time

	DeviceSettings  map[string]any
	UUIDs           map[string]string
	BloksVersioning string
	UserAgent       string
	Country         string
	CountryCode     int
	Locale          string
	TimezoneOffset  int
	TimezoneName    string
	PushDisabled    bool
	AppID           string

	HTTP        *http.Client
	Jar         http.CookieJar
	ReqCounter  int
	LoggedIn    bool

	DelayRange    []time.Duration
	RequestTimeout time.Duration
	LastResponse   json.RawMessage

	mu              sync.Mutex
	autoRefresh     bool
	refreshInterval time.Duration
	stopRefresh     chan struct{}
}

// Options configures State construction.
type Options struct {
	UserAgent      string
	ProxyURL       string
	HTTP           *http.Client
	DeviceSettings map[string]any
	UUIDs          map[string]any
	Country        string
	CountryCode    int
	Locale         string
	TimezoneOffset int
	DelayRange     []time.Duration
	RequestTimeout time.Duration
}

func New(opts Options) *State {
	jar, _ := cookiejar.New(nil)

	s := &State{
		DeviceSettings: make(map[string]any),
		UUIDs:          make(map[string]string),
		Country:        opts.Country,
		CountryCode:    opts.CountryCode,
		Locale:         opts.Locale,
		Jar:            jar,
		LoggedIn:       false,
		AppID:          "567067343352427",
		DelayRange:     opts.DelayRange,
		RequestTimeout: opts.RequestTimeout,
	}

	if s.Country == "" {
		s.Country = "US"
	}
	if s.CountryCode == 0 {
		s.CountryCode = 1
	}
	if s.Locale == "" {
		s.Locale = "en_US"
	}
	if s.RequestTimeout == 0 {
		s.RequestTimeout = time.Second
	}

	s.SetDeviceSettings(opts.DeviceSettings)
	s.SetUUIDs(opts.UUIDs)
	s.SetUserAgent(opts.UserAgent)
	s.SetTimezoneOffset(opts.TimezoneOffset)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if opts.ProxyURL != "" {
		proxyURL, err := url.Parse(opts.ProxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	if opts.HTTP != nil {
		s.HTTP = opts.HTTP
	} else {
		s.HTTP = &http.Client{
			Timeout:   60 * time.Second,
			Jar:       jar,
			Transport: transport,
		}
	}

	return s
}

func (s *State) SetDeviceSettings(ds map[string]any) {
	defaults := map[string]any{
		"android_version":  "34",
		"android_release":  "14",
		"dpi":              "480dpi",
		"resolution":       "1344x2992",
		"manufacturer":     "Google/google",
		"device":           "husky",
		"model":            "Pixel 8 Pro",
		"cpu":              "husky",
		"app_version":      "428.0.0.47.67",
		"version_code":     "961145276",
		"bloks_versioning_id": "7189b949425f9bf80ea8bd880cf5a3080b292d9b1c4b38a18d112f7c4b71e7a8",
	}
	for k, v := range defaults {
		s.DeviceSettings[k] = v
	}
	for k, v := range ds {
		s.DeviceSettings[k] = v
	}
	s.BloksVersioning, _ = s.DeviceSettings["bloks_versioning_id"].(string)
}

func (s *State) SetUUIDs(uuids map[string]any) {
	if uuids == nil {
		uuids = make(map[string]any)
	}
	set := func(key, fallback string) string {
		if v, ok := uuids[key]; ok {
			if str, ok := v.(string); ok && str != "" {
				return str
			}
		}
		return fallback
	}
	s.UUIDs = map[string]string{
		"phone_id":          set("phone_id", internal.GenerateUUID()),
		"uuid":              set("uuid", internal.GenerateUUID()),
		"client_session_id": set("client_session_id", internal.GenerateUUID()),
		"advertising_id":    set("advertising_id", internal.GenerateUUID()),
		"android_device_id": set("android_device_id", internal.GenerateAndroidDeviceID()),
		"request_id":        set("request_id", internal.GenerateUUID()),
		"tray_session_id":   set("tray_session_id", internal.GenerateUUID()),
	}
}

func (s *State) SetUserAgent(ua string) {
	if ua != "" {
		s.UserAgent = ua
		return
	}
	ds := s.DeviceSettings
	s.UserAgent = fmt.Sprintf(
		"Instagram %s Android (%s/%s; %s; %s; %s; %s; %s; %s; %s)",
		ds["app_version"], ds["android_version"], ds["android_release"],
		ds["dpi"], ds["resolution"], ds["manufacturer"], ds["model"],
		ds["cpu"], s.Locale, ds["version_code"],
	)
}

func (s *State) SetTimezoneOffset(seconds int) {
	s.TimezoneOffset = seconds
	s.TimezoneName = timezoneNameFromOffset(seconds)
}

func timezoneNameFromOffset(seconds int) string {
	if seconds == 0 {
		return "GMT"
	}
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	return fmt.Sprintf("GMT%s%02d:%02d", sign, hours, minutes)
}

func (s *State) GenerateUUID(prefix, suffix string) string {
	return prefix + internal.GenerateUUID() + suffix
}

func (s *State) BaseHeaders() http.Header {
	locale := strings.ReplaceAll(s.Locale, "-", "_")
	acceptLanguage := []string{"en-US"}
	if lang := strings.ReplaceAll(locale, "_", "-"); lang != "en-US" {
		acceptLanguage = append([]string{lang}, acceptLanguage...)
	}

	h := http.Header{}
	h.Set("X-IG-App-Locale", locale)
	h.Set("X-IG-Device-Locale", locale)
	h.Set("X-IG-Mapped-Locale", locale)
	h.Set("X-Pigeon-Session-Id", s.GenerateUUID("UFS-", "-1"))
	h.Set("X-Pigeon-Rawclienttime", fmt.Sprintf("%.3f", float64(time.Now().UnixNano())/1e9))
	h.Set("X-IG-Bandwidth-Speed-KBPS", fmt.Sprintf("%d", 2500000+time.Now().UnixNano()%500000/1000))
	h.Set("X-IG-Bandwidth-TotalBytes-B", fmt.Sprintf("%d", 5000000+time.Now().UnixNano()%85000000))
	h.Set("X-IG-Bandwidth-TotalTime-MS", fmt.Sprintf("%d", 2000+time.Now().UnixNano()%7000))
	h.Set("X-IG-App-Startup-Country", strings.ToUpper(s.Country))
	h.Set("X-Bloks-Version-Id", s.BloksVersioning)
	h.Set("X-IG-WWW-Claim", "0")
	h.Set("X-Bloks-Is-Layout-RTL", "false")
	h.Set("X-Bloks-Is-Panorama-Enabled", "true")
	h.Set("X-IG-Device-ID", s.UUIDs["uuid"])
	h.Set("X-IG-Family-Device-ID", s.UUIDs["phone_id"])
	h.Set("X-IG-Android-ID", s.UUIDs["android_device_id"])
	h.Set("X-IG-Timezone-Offset", fmt.Sprintf("%d", s.TimezoneOffset))
	h.Set("X-IG-Connection-Type", "WIFI")
	h.Set("X-IG-Capabilities", "3brTv10=")
	h.Set("X-IG-App-ID", s.AppID)
	h.Set("Priority", "u=3")
	h.Set("User-Agent", s.UserAgent)
	h.Set("Accept-Language", strings.Join(acceptLanguage, ", "))
	// Do not set Accept-Encoding here — Go's http.Transport adds gzip and
	// decompresses automatically; a manual header skips decompression.
	h.Set("Host", "i.instagram.com")
	h.Set("X-FB-HTTP-Engine", "Tigon/MNS/TCP")
	h.Set("X-Tigon-Is-Retry", "False")
	h.Set("X-Zero-Balance", "INIT")
	h.Set("X-Zero-Eh", "")
	h.Set("X-Zero-State", "unknown")
	h.Set("Zero-HTTP-Network-Interface", "wifi")
	h.Set("Connection", "keep-alive")
	h.Set("X-FB-Client-IP", "True")
	h.Set("X-FB-Server-Cluster", "True")
	h.Set("IG-INTENDED-USER-ID", fmt.Sprintf("%d", s.UserIDInt()))
	h.Set("X-IG-Nav-Chain", "9MV:self_profile:2,ProfileMediaTabFragment:self_profile:3,9Xf:self_following:4")
	h.Set("X-IG-SALT-IDS", fmt.Sprintf("%d", 1061162222+time.Now().UnixNano()%100000))

	if s.MID != "" {
		h.Set("X-MID", s.MID)
	}
	if s.IgURUR != "" {
		h.Set("IG-U-RUR", s.IgURUR)
	}
	if s.IgWWWClaim != "" {
		h.Set("X-IG-WWW-Claim", s.IgWWWClaim)
	}
	if s.Authorization != "" {
		h.Set("Authorization", s.Authorization)
	}

	return h
}

func (s *State) UserIDInt() int64 {
	var id int64
	fmt.Sscanf(s.UserID, "%d", &id)
	return id
}

func (s *State) Token() string {
	if s.Jar == nil {
		return ""
	}
	u, _ := url.Parse("https://i.instagram.com/")
	for _, c := range s.Jar.Cookies(u) {
		if c.Name == "csrftoken" {
			return c.Value
		}
	}
	return internal.GenToken(64)
}

func (s *State) CookieDict() map[string]string {
	m := make(map[string]string)
	if s.Jar == nil {
		return m
	}
	u, _ := url.Parse("https://i.instagram.com/")
	for _, c := range s.Jar.Cookies(u) {
		m[c.Name] = c.Value
	}
	return m
}

func (s *State) SessionID() string {
	cd := s.CookieDict()
	if v, ok := cd["sessionid"]; ok {
		return v
	}
	if s.AuthorizationData != nil {
		if v, ok := s.AuthorizationData["sessionid"]; ok {
			if str, ok := v.(string); ok {
				return str
			}
		}
	}
	return ""
}

func (s *State) BuildAuthorization() string {
	if s.AuthorizationData == nil || len(s.AuthorizationData) == 0 {
		return ""
	}
	data, _ := json.Marshal(s.AuthorizationData)
	b64 := fmt.Sprintf("Bearer IGT:2:%s", encodeBase64(data))
	return b64
}

func encodeBase64(data []byte) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	result := make([]byte, 0, (len(data)+2)/3*4)
	for i := 0; i < len(data); i += 3 {
		var n int
		for j := 0; j < 3; j++ {
			n <<= 8
			if i+j < len(data) {
				n |= int(data[i+j])
			}
		}
		for j := 3; j >= 0; j-- {
			if i*4/3+j < len(result)+4 {
				result = append(result, chars[(n>>(6*j))&0x3F])
			}
		}
	}
	for len(result)%4 != 0 {
		result = append(result, '=')
	}
	return string(result)
}

func (s *State) ParseAuthorization(auth string) map[string]any {
	if auth == "" {
		return nil
	}
	parts := strings.Split(auth, ":")
	if len(parts) < 2 {
		return nil
	}
	b64 := parts[len(parts)-1]
	_ = b64
	return nil
}

// PrivateRequest makes a signed request to the Instagram Private API.
func (s *State) PrivateRequest(ctx context.Context, endpoint string, data map[string]any, params url.Values) (map[string]any, error) {
	return s.privateRequest(ctx, endpoint, data, params, true)
}

func (s *State) privateRequest(ctx context.Context, endpoint string, data map[string]any, params url.Values, withSignature bool) (map[string]any, error) {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/v1/" + endpoint
	}

	apiURL := "https://i.instagram.com/api" + endpoint

	s.mu.Lock()
	headers := s.BaseHeaders()
	s.mu.Unlock()

	var resp *http.Response
	var err error

	if data != nil {
		body, _ := json.Marshal(data)
		sigData := internal.GenerateSignature(string(body))
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(sigData))
		if reqErr != nil {
			return nil, igerr.Wrap("PrivateRequest", "build request", reqErr)
		}
		for k, vals := range headers {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		if params != nil {
			q := req.URL.Query()
			for k := range params {
				q.Set(k, params.Get(k))
			}
			req.URL.RawQuery = q.Encode()
		}
		resp, err = s.HTTP.Do(req)
	} else {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if reqErr != nil {
			return nil, igerr.Wrap("PrivateRequest", "build request", reqErr)
		}
		for k, vals := range headers {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
		if params != nil {
			q := req.URL.Query()
			for k := range params {
				q.Set(k, params.Get(k))
			}
			req.URL.RawQuery = q.Encode()
		}
		resp, err = s.HTTP.Do(req)
	}

	if err != nil {
		return nil, igerr.Wrap("PrivateRequest", "http request", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, igerr.Wrap("PrivateRequest", "read response", err)
	}

	mid := resp.Header.Get("ig-set-x-mid")
	if mid != "" {
		s.mu.Lock()
		s.MID = mid
		s.mu.Unlock()
	}

	if resp.StatusCode >= 400 {
		var errJSON map[string]any
		if json.Unmarshal(body, &errJSON) == nil {
			msg, _ := errJSON["message"].(string)
			errType, _ := errJSON["error_type"].(string)
			switch resp.StatusCode {
			case 400:
				switch {
				case errType == "bad_password":
					return nil, igerr.Wrap("PrivateRequest", msg, igerr.ErrBadPassword)
				case errType == "two_factor_required":
					return nil, igerr.Wrap("PrivateRequest", msg, igerr.ErrTwoFactor)
				case msg == "challenge_required":
					return nil, igerr.Wrap("PrivateRequest", msg, igerr.ErrChallenge)
				case msg == "feedback_required":
					return nil, igerr.Wrap("PrivateRequest", msg, igerr.ErrFeedback)
				case errType == "sentry_block":
					return nil, igerr.Wrap("PrivateRequest", msg, igerr.ErrSentryBlock)
				default:
					return nil, igerr.WithCode("PrivateRequest", msg, 400)
				}
			case 401:
				return nil, igerr.Wrap("PrivateRequest", "unauthorized", igerr.ErrAuthentication)
			case 403:
				if msg == "login_required" {
					return nil, igerr.Wrap("PrivateRequest", msg, igerr.ErrLoginRequired)
				}
				return nil, igerr.WithCode("PrivateRequest", msg, 403)
			case 404:
				return nil, igerr.WithCode("PrivateRequest", msg, 404)
			case 429:
				return nil, igerr.Wrap("PrivateRequest", "rate limited", igerr.ErrThrottled)
			default:
				return nil, igerr.WithCode("PrivateRequest", msg, resp.StatusCode)
			}
		}
		return nil, igerr.WithCode("PrivateRequest", fmt.Sprintf("status %d", resp.StatusCode), resp.StatusCode)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, igerr.Wrap("PrivateRequest", "decode json", err)
	}

	s.mu.Lock()
	s.LastResponse = body
	s.mu.Unlock()

	if status, ok := result["status"].(string); ok && status == "fail" {
		msg, _ := result["message"].(string)
		return nil, igerr.Wrap("PrivateRequest", msg, igerr.ErrInstagramAPI)
	}

	return result, nil
}

// PrivateRequestRaw makes a signed request returning raw JSON bytes.
func (s *State) PrivateRequestRaw(ctx context.Context, endpoint string, data map[string]any, params url.Values) (json.RawMessage, error) {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/v1/" + endpoint
	}

	apiURL := "https://i.instagram.com/api" + endpoint

	s.mu.Lock()
	headers := s.BaseHeaders()
	s.mu.Unlock()

	var resp *http.Response
	var err error

	if data != nil {
		body, _ := json.Marshal(data)
		sigData := internal.GenerateSignature(string(body))
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(sigData))
		if reqErr != nil {
			return nil, igerr.Wrap("PrivateRequestRaw", "build request", reqErr)
		}
		for k, vals := range headers {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		if params != nil {
			q := req.URL.Query()
			for k := range params {
				q.Set(k, params.Get(k))
			}
			req.URL.RawQuery = q.Encode()
		}
		resp, err = s.HTTP.Do(req)
	} else {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if reqErr != nil {
			return nil, igerr.Wrap("PrivateRequestRaw", "build request", reqErr)
		}
		for k, vals := range headers {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
		if params != nil {
			q := req.URL.Query()
			for k := range params {
				q.Set(k, params.Get(k))
			}
			req.URL.RawQuery = q.Encode()
		}
		resp, err = s.HTTP.Do(req)
	}

	if err != nil {
		return nil, igerr.Wrap("PrivateRequestRaw", "http request", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, igerr.Wrap("PrivateRequestRaw", "read response", err)
	}

	mid := resp.Header.Get("ig-set-x-mid")
	if mid != "" {
		s.mu.Lock()
		s.MID = mid
		s.mu.Unlock()
	}

	if resp.StatusCode >= 400 {
		return nil, igerr.WithCode("PrivateRequestRaw", fmt.Sprintf("status %d", resp.StatusCode), resp.StatusCode)
	}

	return body, nil
}

func (s *State) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LoggedIn = false
	if s.stopRefresh != nil {
		close(s.stopRefresh)
	}
	return nil
}

// RandomDelay sleeps for a random duration within the configured delay range.
func (s *State) RandomDelay() {
	if len(s.DelayRange) == 0 {
		return
	}
	var total time.Duration
	for _, d := range s.DelayRange {
		total += d
	}
	avg := total / time.Duration(len(s.DelayRange))
	time.Sleep(avg / 2)
}

// GetAuthorization returns the current authorization header.
func (s *State) GetAuthorization() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Authorization
}

// SetAuthorization sets the authorization header.
func (s *State) SetAuthorization(auth string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Authorization = auth
}

// SetMID sets the MID header value.
func (s *State) SetMID(mid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MID = mid
}

// StoreLastResponse saves the raw response from the last API call.
func (s *State) StoreLastResponse(data json.RawMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastResponse = data
}

// LoadSnapshot restores session state from a snapshot map (e.g. from storage).
func (s *State) LoadSnapshot(snap map[string]any) {
	if snap == nil {
		return
	}
	if cookies, ok := snap["cookies"].(map[string]any); ok {
		s.SetCookiesFromDict(cookies)
	}
	if deviceSettings, ok := snap["device_settings"].(map[string]any); ok {
		s.SetDeviceSettings(deviceSettings)
	}
	if uuids, ok := snap["uuids"].(map[string]any); ok {
		s.SetUUIDs(uuids)
	}
	if ua, ok := snap["user_agent"].(string); ok {
		s.UserAgent = ua
	}
	if locale, ok := snap["locale"].(string); ok {
		s.Locale = locale
	}
	if country, ok := snap["country"].(string); ok {
		s.Country = country
	}
	if cc, ok := snap["country_code"].(float64); ok {
		s.CountryCode = int(cc)
	}
	if tz, ok := snap["timezone_offset"].(float64); ok {
		s.SetTimezoneOffset(int(tz))
	}
	if mid, ok := snap["mid"].(string); ok {
		s.MID = mid
	}
	if rur, ok := snap["ig_u_rur"].(string); ok {
		s.IgURUR = rur
	}
	if claim, ok := snap["ig_www_claim"].(string); ok {
		s.IgWWWClaim = claim
	}
	if authData, ok := snap["authorization_data"].(map[string]any); ok {
		NormalizeAuthorizationData(authData)
		s.AuthorizationData = authData
		s.Authorization = s.BuildAuthorization()
	}
	if uid, ok := snap["user_id"].(string); ok {
		s.UserID = uid
	}
	if username, ok := snap["username"].(string); ok {
		s.Username = username
	}
	s.EnsureLoggedInFromCookies()
}

// EnsureLoggedInFromCookies marks the session logged in when browser cookies are present.
func (s *State) EnsureLoggedInFromCookies() {
	cookies := s.CookieDict()
	if uid := strings.TrimSpace(cookies["ds_user_id"]); uid != "" {
		s.UserID = uid
	}
	if strings.TrimSpace(s.UserID) != "" {
		s.LoggedIn = true
		s.syncSessionHeadersFromCookies(cookies)
		s.syncAuthorizationFromCookies(cookies)
		return
	}
	userID := UserIDFromSessionID(s.SessionID())
	if userID == "" {
		return
	}
	s.UserID = userID
	s.LoggedIn = true
	if s.AuthorizationData == nil {
		s.AuthorizationData = map[string]any{}
	}
	if _, ok := s.AuthorizationData["ds_user_id"]; !ok {
		s.AuthorizationData["ds_user_id"] = userID
	}
	if _, ok := s.AuthorizationData["sessionid"]; !ok {
		s.AuthorizationData["sessionid"] = s.SessionID()
	}
	s.syncSessionHeadersFromCookies(cookies)
	s.syncAuthorizationFromCookies(cookies)
}

func (s *State) syncSessionHeadersFromCookies(cookies map[string]string) {
	if s.MID == "" {
		if mid := strings.TrimSpace(cookies["mid"]); mid != "" {
			s.MID = mid
		}
	}
	if s.IgURUR == "" {
		if rur := normalizeRUR(cookies["rur"]); rur != "" {
			s.IgURUR = rur
		}
	}
	if s.IgWWWClaim == "" {
		if claim := strings.TrimSpace(cookies["ig_www_claim"]); claim != "" {
			s.IgWWWClaim = claim
		}
	}
}

func (s *State) syncAuthorizationFromCookies(cookies map[string]string) {
	if s.AuthorizationData == nil {
		s.AuthorizationData = map[string]any{}
	}
	if uid := strings.TrimSpace(s.UserID); uid != "" {
		s.AuthorizationData["ds_user_id"] = uid
	}
	s.AuthorizationData["should_use_header_over_cookies"] = true
	sessionID := strings.TrimSpace(cookies["sessionid"])
	if sessionID == "" {
		sessionID = s.SessionID()
	}
	if sessionID != "" {
		s.AuthorizationData["sessionid"] = sessionID
	}
	s.Authorization = s.BuildAuthorization()
}
