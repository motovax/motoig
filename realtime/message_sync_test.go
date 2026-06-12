package realtime

import (
	"encoding/json"
	"testing"

	igref "github.com/motovax/motoig/references/instagrapi"
)

func TestDispatchMessageSyncInstagrapiShape(t *testing.T) {
	c := &RealtimeClient{handlers: make(map[string][]RealtimeHandler)}
	var got map[string]any
	c.On("message", func(event RealtimeEvent) {
		got, _ = event.Payload.(map[string]any)
	})

	body := []byte(`[{
		"data": [{
			"path": "/direct_v2/threads/999/items/",
			"op": "add",
			"value": "{\"text\":\"hello\",\"user_id\":42}"
		}]
	}]`)
	c.dispatchMessageSync(body)

	if got == nil {
		t.Fatal("expected message event")
	}
	if _, ok := got["message"].(map[string]any); !ok {
		t.Fatalf("expected nested instagrapi message key, got %v", got)
	}

	flat := igref.UnwrapMessagePayload(got)
	if flat["thread_id"] != "999" {
		t.Fatalf("thread_id = %v", flat["thread_id"])
	}
	if flat["text"] != "hello" {
		t.Fatalf("text = %v", flat["text"])
	}
	if _, err := json.Marshal(got); err != nil {
		t.Fatalf("marshal emitted payload: %v", err)
	}
}