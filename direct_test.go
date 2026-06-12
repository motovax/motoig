package motoig

import (
	"encoding/json"
	"testing"
)

func TestExtractIrisSyncStateTopLevel(t *testing.T) {
	raw := json.RawMessage(`{"seq_id":42,"snapshot_at_ms":1700000000000,"status":"ok"}`)
	seqID, snapshotAtMS, err := extractIrisSyncState(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seqID != 42 {
		t.Fatalf("seq_id = %d, want 42", seqID)
	}
	if snapshotAtMS != 1700000000000 {
		t.Fatalf("snapshot_at_ms = %d, want 1700000000000", snapshotAtMS)
	}
}

func TestExtractIrisSyncStateNestedInbox(t *testing.T) {
	raw := json.RawMessage(`{
		"inbox": {
			"threads": [],
			"seq_id": 99,
			"snapshot_at_ms": 1800000000000
		},
		"status": "ok"
	}`)
	seqID, snapshotAtMS, err := extractIrisSyncState(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seqID != 99 {
		t.Fatalf("seq_id = %d, want 99", seqID)
	}
	if snapshotAtMS != 1800000000000 {
		t.Fatalf("snapshot_at_ms = %d, want 1800000000000", snapshotAtMS)
	}
}

func TestExtractIrisSyncStateMissing(t *testing.T) {
	raw := json.RawMessage(`{"inbox":{"threads":[]},"status":"ok"}`)
	_, _, err := extractIrisSyncState(raw)
	if err == nil {
		t.Fatal("expected error for missing seq_id/snapshot_at_ms")
	}
}

func TestSetSessionIDPreservesCookieJar(t *testing.T) {
	c := New()
	sessionID := "1234567890%3Atestsessionidvaluewithmorethan30chars"
	if err := c.SetSessionID(t.Context(), sessionID); err != nil {
		t.Fatalf("SetSessionID: %v", err)
	}
	if c.state.Jar == nil {
		t.Fatal("expected cookie jar to be initialized")
	}
	if c.state.HTTP == nil || c.state.HTTP.Jar == nil {
		t.Fatal("expected HTTP client jar to be initialized")
	}
	if got := c.state.SessionID(); got == "" {
		t.Fatal("expected sessionid to be stored")
	}
	if c.state.UserID != "1234567890" {
		t.Fatalf("user id = %q, want 1234567890", c.state.UserID)
	}
}

func TestLoadSettingsRestoresUserIDFromEncodedSessionCookie(t *testing.T) {
	c := New()
	c.LoadSettings(map[string]any{
		"cookies": map[string]any{
			"sessionid": "1234567890%3Atestsessionidvaluewithmorethan30chars",
		},
	})
	if c.UserID() != "1234567890" {
		t.Fatalf("user id = %q, want 1234567890", c.UserID())
	}
	settings := c.GetSettings()
	if settings["user_id"] != "1234567890" {
		t.Fatalf("saved user_id = %v, want 1234567890", settings["user_id"])
	}
}

func TestLoadSettingsSyncsMIDFromCookies(t *testing.T) {
	c := New()
	c.LoadSettings(map[string]any{
		"cookies": map[string]any{
			"sessionid":   "1234567890%3Atestsessionidvaluewithmorethan30chars",
			"ds_user_id":  "1234567890",
			"mid":         "test-mid-value",
			"rur":         "\"CCO,123,999:abc\"",
		},
	})
	if c.state.MID != "test-mid-value" {
		t.Fatalf("mid = %q, want test-mid-value", c.state.MID)
	}
	if c.state.IgURUR != "CCO,123,999:abc" {
		t.Fatalf("ig_u_rur = %q", c.state.IgURUR)
	}
}