package motoig

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	igerr "github.com/motovax/motoig/errors"
	"github.com/motovax/motoig/internal"
)

// Login authenticates with username and password.
func (c *Client) Login(ctx context.Context, username, password string) error {
	return c.LoginWithCode(ctx, username, password, "")
}

// LoginWithCode authenticates with username, password, and optional 2FA code.
func (c *Client) LoginWithCode(ctx context.Context, username, password, verificationCode string) error {
	c.username = username
	c.password = password

	c.state.Username = username

	encPassword := c.passwordEncrypt(password)

	data := c.withDefaultData(nil)
	data["jazoest"] = internal.GenerateJazoest(c.state.UUIDs["phone_id"])
	data["country_codes"] = fmt.Sprintf(`[{"country_code":"%d","source":["default"]}]`, c.state.CountryCode)
	data["phone_id"] = c.state.UUIDs["phone_id"]
	data["enc_password"] = encPassword
	data["username"] = username
	data["adid"] = c.state.UUIDs["advertising_id"]
	data["guid"] = c.state.UUIDs["uuid"]
	data["device_id"] = c.state.UUIDs["android_device_id"]
	data["google_tokens"] = "[]"
	data["login_attempt_count"] = "0"

	result, err := c.state.PrivateRequest(ctx, "accounts/login/", data, nil)
	if err != nil {
		if verificationCode != "" {
			return c.twoFactorLogin(ctx, verificationCode)
		}
		return err
	}

	_ = result

	c.state.AuthorizationData = c.parseAuthorization(ctx)
	c.state.Authorization = c.state.BuildAuthorization()
	c.state.LastLogin = time.Now()
	c.state.LoggedIn = true

	c.loginFlow(ctx)

	return nil
}

// LoginBySessionID logs in using a session ID string.
func (c *Client) LoginBySessionID(ctx context.Context, sessionID string) error {
	if len(sessionID) < 30 {
		return igerr.New("LoginBySessionID", "invalid sessionid")
	}

	userID := ""
	for i := 0; i < len(sessionID); i++ {
		if sessionID[i] >= '0' && sessionID[i] <= '9' {
			userID += string(sessionID[i])
		} else {
			break
		}
	}
	if userID == "" {
		return igerr.New("LoginBySessionID", "invalid sessionid: no user id")
	}

	c.state.SetCookie("sessionid", sessionID)
	c.state.UserID = userID
	c.state.LoggedIn = true
	c.state.AuthorizationData = map[string]any{
		"ds_user_id": userID,
		"sessionid":  sessionID,
	}
	c.state.Authorization = c.state.BuildAuthorization()

	profile, err := c.UserInfoV1(ctx, userID)
	if err != nil {
		c.log.Warn("failed to fetch user info via v1, trying stream", "error", err)
	} else {
		c.state.Username = profile.Username
	}

	return nil
}

func (c *Client) twoFactorLogin(ctx context.Context, code string) error {
	var twoFactorInfo map[string]any
	if c.state.LastResponse != nil {
		json.Unmarshal(c.state.LastResponse, &twoFactorInfo)
	}

	twoFactorIdentifier := ""
	if info, ok := twoFactorInfo["two_factor_info"].(map[string]any); ok {
		twoFactorIdentifier, _ = info["two_factor_identifier"].(string)
	}

	data := c.withDefaultData(nil)
	data["verification_code"] = code
	data["phone_id"] = c.state.UUIDs["phone_id"]
	data["_csrftoken"] = c.state.Token()
	data["two_factor_identifier"] = twoFactorIdentifier
	data["username"] = c.username
	data["trust_this_device"] = "0"
	data["guid"] = c.state.UUIDs["uuid"]
	data["device_id"] = c.state.UUIDs["android_device_id"]
	data["waterfall_id"] = internal.GenerateUUID()
	data["verification_method"] = "3"

	result, err := c.state.PrivateRequest(ctx, "accounts/two_factor_login/", data, nil)
	if err != nil {
		return err
	}

	_ = result

	c.state.AuthorizationData = c.parseAuthorization(ctx)
	c.state.Authorization = c.state.BuildAuthorization()
	c.state.LastLogin = time.Now()
	c.state.LoggedIn = true

	c.loginFlow(ctx)

	return nil
}

func (c *Client) loginFlow(ctx context.Context) {
	c.getReelsTrayFeed(ctx, "cold_start")
	c.getTimelineFeed(ctx, []string{"cold_start_fetch"})
}

func (c *Client) getReelsTrayFeed(ctx context.Context, reason string) {
	data := c.withDefaultData(nil)
	data["reason"] = reason
	data["tray_session_id"] = c.state.UUIDs["tray_session_id"]
	data["request_id"] = c.state.UUIDs["request_id"]
	data["supported_capabilities_new"] = mustJSON([]map[string]string{
		{"value": "119.0,120.0,121.0,122.0,123.0,124.0,125.0,126.0,127.0,128.0,129.0,130.0,131.0,132.0,133.0,134.0,135.0,136.0,137.0,138.0,139.0,140.0,141.0,142.0", "name": "SUPPORTED_SDK_VERSIONS"},
		{"value": "14", "name": "FACE_TRACKER_VERSION"},
	})
	data["page_size"] = "50"
	if reason == "cold_start" {
		data["reel_tray_impressions"] = "{}"
	}

	c.state.PrivateRequest(ctx, "feed/reels_tray/", data, nil)
}

func (c *Client) getTimelineFeed(ctx context.Context, reasons []string) {
	data := c.withDefaultData(nil)
	data["reason"] = "pull_to_refresh"
	data["feed_view_info"] = "[]"
	data["request_id"] = c.state.UUIDs["request_id"]
	data["device_timezone_name"] = c.state.TimezoneName
	data["timezone_offset"] = fmt.Sprintf("%d", c.state.TimezoneOffset)
	data["push_disabled"] = "true"

	c.state.PrivateRequest(ctx, "feed/timeline/", data, nil)
}

func (c *Client) passwordEncrypt(password string) string {
	return fmt.Sprintf("#PWD_INSTAGRAM_BROWSER:0:%s:%s", internal.GenToken(10), password)
}

func (c *Client) parseAuthorization(ctx context.Context) map[string]any {
	resp := c.state.LastResponse
	if resp == nil {
		return nil
	}

	var result map[string]any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil
	}

	return result
}

// Logout logs out the current session.
func (c *Client) Logout(ctx context.Context) error {
	data := c.withDefaultData(nil)
	data["one_tap_app_login"] = "true"
	result, err := c.state.PrivateRequest(ctx, "accounts/logout/", data, nil)
	if err != nil {
		return err
	}

	if status, ok := result["status"].(string); ok && status == "ok" {
		c.state.AuthorizationData = nil
		c.state.Authorization = ""
		c.state.LoggedIn = false
		return nil
	}

	return igerr.New("Logout", "logout failed")
}

// Relogin re-authenticates with stored credentials.
func (c *Client) Relogin(ctx context.Context) error {
	if c.username == "" || c.password == "" {
		return igerr.New("Relogin", "no stored credentials")
	}
	c.state.AuthorizationData = nil
	c.state.Authorization = ""
	return c.Login(ctx, c.username, c.password)
}

// Expose performs the qe/expose/ call.
func (c *Client) Expose(ctx context.Context) error {
	data := c.withDefaultData(nil)
	data["experiment"] = "ig_android_profile_contextual_feed"
	_, err := c.state.PrivateRequest(ctx, "qe/expose/", data, nil)
	return err
}
