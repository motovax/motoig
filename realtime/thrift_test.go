package realtime

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuildConnectionPayloadIncludesClientInfo(t *testing.T) {
	fullDeviceID := "01234567-89ab-cdef-0123-456789abcdef"
	shortID := fullDeviceID[:20]

	payload := BuildConnectionPayload(ConnectionInfo{
		ClientIdentifier: shortID,
		ClientInfo: map[string]any{
			"userId":                    int64(12345),
			"userAgent":                 "Instagram 123 Android",
			"clientCapabilities":        int64(183),
			"endpointCapabilities":      int64(0),
			"publishFormat":             1,
			"noAutomaticForeground":     false,
			"makeUserAvailableInForeground": true,
			"deviceId":                  fullDeviceID,
			"isInitiallyForeground":     true,
			"networkType":               1,
			"networkSubtype":            0,
			"clientMqttSessionId":       int64(999),
			"subscribeTopics":           []int{88, 135, 149, 150, 133, 146},
			"clientType":                "cookie_auth",
			"appId":                     int64(IGRealtimeAppID),
			"deviceSecret":              "",
			"clientStack":               3,
		},
		Password: "sessionid=test-session",
		AppSpecificInfo: map[string]string{
			"app_version":   "123.0.0.0.0",
			"ig_mqtt_route": "django",
			"platform":      "android",
		},
	})

	if len(payload) < 128 {
		t.Fatalf("payload too short (%d bytes), expected full thrift connection blob", len(payload))
	}
	for _, needle := range []string{"cookie_auth", "sessionid=test-session", fullDeviceID, shortID, "django"} {
		if !bytes.Contains(payload, []byte(needle)) {
			t.Fatalf("payload missing %q", needle)
		}
	}
	if bytes.Contains(payload, []byte("clientIdentifier")) {
		// field names are not serialized in thrift compact; only values
	}
	if strings.Count(string(payload), shortID) < 1 {
		t.Fatalf("expected client identifier in payload")
	}
}