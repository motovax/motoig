package realtime

import "testing"

func TestThreadIDFromPath(t *testing.T) {
	c := &RealtimeClient{}
	tests := []struct {
		path string
		want string
	}{
		{"/direct_v2/threads/340282366841710300949128352912842607044/items/", "340282366841710300949128352912842607044"},
		{"/direct_v2/inbox/threads/12345/items/", "12345"},
		{"/direct_v2/inbox/", ""},
	}
	for _, tc := range tests {
		if got := c.threadIDFromPath(tc.path); got != tc.want {
			t.Fatalf("threadIDFromPath(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}