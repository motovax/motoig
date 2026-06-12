package instagrapi

import "testing"

func TestUnwrapMessagePayloadNested(t *testing.T) {
	in := map[string]any{
		"event": "message",
		"message": map[string]any{
			"path":      "/direct_v2/threads/123/items/",
			"op":        "add",
			"thread_id": "123",
			"text":      "hi",
		},
	}
	out := UnwrapMessagePayload(in)
	if out["path"] != "/direct_v2/threads/123/items/" {
		t.Fatalf("path = %v", out["path"])
	}
	if out["text"] != "hi" {
		t.Fatalf("text = %v", out["text"])
	}
	if out["event"] != "message" {
		t.Fatalf("event = %v", out["event"])
	}
}

func TestUnwrapMessagePayloadFlat(t *testing.T) {
	in := map[string]any{
		"path": "/direct_v2/threads/123/items/",
		"op":   "add",
	}
	if got := UnwrapMessagePayload(in); got["path"] != in["path"] {
		t.Fatalf("flat payload changed: %v", got)
	}
}

func TestRefKnownSymbol(t *testing.T) {
	url := Ref("Client.DirectThreads")
	if !stringsHasPrefix(url, Repo+"/blob/"+Branch+"/instagrapi/mixins/direct.py") {
		t.Fatalf("unexpected ref url: %s", url)
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}