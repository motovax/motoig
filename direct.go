package motoig

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"strconv"

	igerr "github.com/motovax/motoig/errors"
	"github.com/motovax/motoig/extractors"
	"github.com/motovax/motoig/internal"
	"github.com/motovax/motoig/models"
)

func (c *Client) privateRequestGET(ctx context.Context, endpoint string, params url.Values) (map[string]any, error) {
	return c.state.PrivateRequest(ctx, endpoint, nil, params)
}

func (c *Client) directInboxParams(amount int) url.Values {
	if amount <= 0 {
		amount = 20
	}
	pushDisabled := "true"
	if c != nil && c.state != nil && !c.state.PushDisabled {
		pushDisabled = "false"
	}
	params := url.Values{}
	params.Set("visual_message_return_type", "unseen")
	params.Set("thread_message_limit", "10")
	params.Set("persistentBadging", "true")
	params.Set("limit", fmt.Sprintf("%d", amount))
	params.Set("is_prefetching", "false")
	params.Set("fetch_reason", "initial_snapshot")
	params.Set("include_old_mrs", "false")
	params.Set("no_pending_badge", "true")
	params.Set("push_disabled", pushDisabled)
	params.Set("eb_device_id", "0")
	params.Set("igd_request_log_tracking_id", internal.GenerateUUID())
	return params
}

func directThreadParams(amount int) url.Values {
	if amount <= 0 {
		amount = 20
	}
	params := url.Values{}
	params.Set("visual_message_return_type", "unseen")
	params.Set("direction", "older")
	params.Set("limit", fmt.Sprintf("%d", amount))
	return params
}

func extractIrisSyncState(raw json.RawMessage) (seqID int, snapshotAtMS int64, err error) {
	if len(raw) == 0 {
		return 0, 0, fmt.Errorf("empty inbox response")
	}

	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return 0, 0, err
	}

	seqID = intFromJSON(top["seq_id"])
	snapshotAtMS = int64FromJSON(top["snapshot_at_ms"])
	if seqID != 0 && snapshotAtMS != 0 {
		return seqID, snapshotAtMS, nil
	}

	if inbox, ok := top["inbox"].(map[string]any); ok {
		if seqID == 0 {
			seqID = intFromJSON(inbox["seq_id"])
		}
		if snapshotAtMS == 0 {
			snapshotAtMS = int64FromJSON(inbox["snapshot_at_ms"])
		}
	}

	if seqID == 0 || snapshotAtMS == 0 {
		return 0, 0, fmt.Errorf("seq_id or snapshot_at_ms missing from inbox response")
	}
	return seqID, snapshotAtMS, nil
}

func intFromJSON(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	default:
		return 0
	}
}

func int64FromJSON(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	default:
		return 0
	}
}

func ensureSessionJar(c *Client) {
	if c.state.Jar != nil {
		return
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return
	}
	c.state.Jar = jar
	if c.state.HTTP != nil {
		c.state.HTTP.Jar = jar
	}
}

// DirectThreads returns direct message threads.
//
// Port of instagrapi.DirectMixin.direct_threads_chunk.
// Reference: https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/direct.py
func (c *Client) DirectThreads(ctx context.Context, amount int) ([]models.DirectThread, error) {
	result, err := c.privateRequestGET(ctx, "direct_v2/inbox/", c.directInboxParams(amount))
	if err != nil {
		return nil, err
	}
	var threads []models.DirectThread
	if items, ok := result["inbox"].(map[string]any); ok {
		if threadsRaw, ok := items["threads"].([]any); ok {
			for _, t := range threadsRaw {
				if tm, ok := t.(map[string]any); ok {
					threads = append(threads, extractors.ExtractDirectThread(tm))
				}
			}
		}
	}
	return threads, nil
}

// DirectMessages returns messages in a direct thread.
//
// Port of instagrapi.DirectMixin.direct_messages.
// Reference: https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/direct.py
func (c *Client) DirectMessages(ctx context.Context, threadID string, amount int) ([]models.DirectMessage, error) {
	result, err := c.privateRequestGET(ctx, "direct_v2/threads/"+threadID+"/", directThreadParams(amount))
	if err != nil {
		return nil, err
	}
	var messages []models.DirectMessage
	if thread, ok := result["thread"].(map[string]any); ok {
		if items, ok := thread["items"].([]any); ok {
			for _, item := range items {
				if im, ok := item.(map[string]any); ok {
					messages = append(messages, extractors.ExtractDirectMessage(im))
				}
			}
		}
	}
	return messages, nil
}

// DirectSend sends a text message to a thread.
//
// Port of instagrapi.DirectMixin.direct_send (text broadcast).
// Reference: https://github.com/subzeroid/instagrapi/blob/master/instagrapi/mixins/direct.py
func (c *Client) DirectSend(ctx context.Context, threadID, text string) error {
	token := internal.GenerateMutationToken()
	data := c.withDefaultData(nil)
	data["action"] = "send_item"
	data["is_x_transport_forward"] = "false"
	data["send_silently"] = "false"
	data["is_shh_mode"] = "0"
	data["send_attribution"] = "message_button"
	data["client_context"] = token
	data["mutation_token"] = token
	data["offline_threading_id"] = token
	data["btt_dual_send"] = "false"
	data["is_ae_dual_send"] = "false"
	data["text"] = text
	data["thread_ids"] = fmt.Sprintf("[%s]", threadID)
	data["device_id"] = c.state.UUIDs["android_device_id"]
	data["_csrftoken"] = c.state.Token()
	_, err := c.privateRequest(ctx, "direct_v2/threads/broadcast/text/", data)
	return err
}

// RealtimeDirectSubscribe subscribes to real-time DM updates via IRIS.
//
// Port of instagrapi.RealtimeClient.direct_subscribe.
// Reference: https://github.com/subzeroid/instagrapi/blob/master/instagrapi/realtime/client.py
func (c *Client) RealtimeDirectSubscribe(ctx context.Context, amount int) error {
	if c.realtime == nil || !c.realtime.IsConnected() {
		return igerr.New("RealtimeDirectSubscribe", "not connected")
	}

	if _, err := c.DirectThreads(ctx, amount); err != nil {
		return igerr.Wrap("RealtimeDirectSubscribe", "fetch threads", err)
	}

	seqID, snapshotAtMS, err := extractIrisSyncState(c.state.LastResponse)
	if err != nil {
		return igerr.Wrap("RealtimeDirectSubscribe", "extract iris sync state", err)
	}

	return c.realtime.DirectSubscribe(amount, seqID, snapshotAtMS)
}