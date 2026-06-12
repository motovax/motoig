package instagrapi

// UnwrapMessagePayload normalizes realtime "message" event payloads.
//
// instagrapi emits message-sync events as {"message": {...fields...}, ...meta}.
// motoig preserves that shape for parity. Callers that expect a flat patch map
// (path, op, thread_id, text, user_id, ...) can pass the payload through here.
func UnwrapMessagePayload(payload map[string]any) map[string]any {
	if payload == nil {
		return nil
	}
	nested, ok := payload["message"].(map[string]any)
	if !ok {
		return payload
	}
	out := make(map[string]any, len(payload)+len(nested))
	for k, v := range payload {
		if k == "message" {
			continue
		}
		out[k] = v
	}
	for k, v := range nested {
		out[k] = v
	}
	return out
}