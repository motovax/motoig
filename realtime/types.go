package realtime

type DirectMessageEvent struct {
	ThreadID string         `json:"thread_id"`
	Op       string         `json:"op"`
	Path     string         `json:"path"`
	Message  map[string]any `json:"message,omitempty"`
}

type TypingEvent struct {
	ThreadID string `json:"thread_id"`
	UserID   string `json:"user_id"`
	Active   bool   `json:"active"`
}

type PresenceEvent struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SeenEvent struct {
	ThreadID string `json:"thread_id"`
	UserID   string `json:"user_id"`
	ItemID   string `json:"item_id"`
}
