package extractors

import "testing"

func TestExtractDirectMessageXMAReelMention(t *testing.T) {
	msg := ExtractDirectMessage(map[string]any{
		"item_id":   "123",
		"item_type": "xma_reel_mention",
		"user_id":   float64(549277362),
		"timestamp": float64(1733392992000000),
		"generic_xma": []any{
			map[string]any{
				"target_url":         "https://www.instagram.com/reel/ABC123/",
				"header_title_text":  "Mentioned you in their reel",
				"title_text":         "Check this out",
			},
		},
	})

	if msg.ItemType != "xma_reel_mention" {
		t.Fatalf("item_type = %q", msg.ItemType)
	}
	if msg.XMAShare == nil {
		t.Fatal("expected xma_share")
	}
	if msg.XMAShare.HeaderTitleText != "Mentioned you in their reel" {
		t.Fatalf("header_title_text = %q", msg.XMAShare.HeaderTitleText)
	}
	if msg.XMAShare.Title != "Check this out" {
		t.Fatalf("title = %q", msg.XMAShare.Title)
	}
}