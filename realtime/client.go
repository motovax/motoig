// Package realtime implements Instagram's MQTToT protocol for real-time DMs.
//
// Port of instagrapi.realtime — https://github.com/subzeroid/instagrapi/tree/master/instagrapi/realtime
package realtime

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	RealtimeHost       = "edge-mqtt.facebook.com"
	IGRealtimeAppID    = 567067343352427
	SubscribeTopicsStr = "88,135,149,150,133,146"
)

type RealtimeEvent struct {
	Topic   string
	Payload any
}

type RealtimeHandler func(event RealtimeEvent)

type RealtimeClient struct {
	transport        *MQTToTTransport
	sessionID        string
	userID           int64
	clientIdentifier string
	deviceID         string
	userAgent        string
	appVersion       string
	capabilities     string
	locale           string
	connected        bool
	packetID         uint16
	mu               sync.RWMutex
	handlers         map[string][]RealtimeHandler
	stopCh           chan struct{}
}

type RealtimeConfig struct {
	SessionID          string
	UserID             int64
	ClientIdentifier   string
	DeviceID           string
	UserAgent          string
	AppVersion         string
	Capabilities       string
	Locale             string
}

func NewRealtimeClient(cfg RealtimeConfig) *RealtimeClient {
	clientIdentifier := cfg.ClientIdentifier
	if clientIdentifier == "" {
		clientIdentifier = cfg.DeviceID
		if len(clientIdentifier) > 20 {
			clientIdentifier = clientIdentifier[:20]
		}
	}
	return &RealtimeClient{
		transport:        NewMQTToTTransport(RealtimeHost),
		sessionID:        cfg.SessionID,
		userID:           cfg.UserID,
		clientIdentifier: clientIdentifier,
		deviceID:         cfg.DeviceID,
		userAgent:        cfg.UserAgent,
		appVersion:       cfg.AppVersion,
		capabilities:     cfg.Capabilities,
		locale:           cfg.Locale,
		handlers:         make(map[string][]RealtimeHandler),
		stopCh:           make(chan struct{}),
	}
}

func (c *RealtimeClient) On(event string, handler RealtimeHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[event] = append(c.handlers[event], handler)
}

func (c *RealtimeClient) Connect() error {
	c.mu.Lock()
	c.stopCh = make(chan struct{})
	c.mu.Unlock()

	if err := c.transport.Connect(); err != nil {
		return err
	}

	connPayload := c.buildConnection()
	connectPacket := WriteConnectPacket(connPayload, 20)

	if err := c.transport.Send(connectPacket); err != nil {
		return fmt.Errorf("send connect: %w", err)
	}

	resp, err := c.transport.RecvPacket()
	if err != nil {
		return fmt.Errorf("recv connack: %w", err)
	}

	pkt, err := DecodePacketData(resp)
	if err != nil {
		return fmt.Errorf("decode connack: %w", err)
	}
	if pkt.PacketType != PacketConnack || pkt.ReturnCode != 0 {
		return fmt.Errorf("connect rejected: type=%d code=%d", pkt.PacketType, pkt.ReturnCode)
	}

	c.connected = true
	go c.readLoop()
	return nil
}

func (c *RealtimeClient) Disconnect() error {
	c.mu.Lock()
	wasConnected := c.connected
	c.connected = false
	stopCh := c.stopCh
	c.mu.Unlock()

	if wasConnected {
		_ = c.transport.Send(WriteDisconnectPacket())
	}
	if stopCh != nil {
		select {
		case <-stopCh:
		default:
			close(stopCh)
		}
	}
	return c.transport.Close()
}

func (c *RealtimeClient) Ping() error {
	return c.transport.Send(WritePingreqPacket())
}

func (c *RealtimeClient) IrisSubscribe(seqID int, snapshotAtMS int64) error {
	payload := map[string]any{
		"seq_id":           seqID,
		"snapshot_at_ms":   snapshotAtMS,
		"snapshot_app_version": c.appVersion,
	}
	return c.PublishJSON(TopicIrisSub, payload)
}

func (c *RealtimeClient) DirectSubscribe(amount int, seqID int, snapshotAtMS int64) error {
	return c.IrisSubscribe(seqID, snapshotAtMS)
}

func (c *RealtimeClient) SendDirectMessage(threadID string, text string, clientContext string) error {
	payload := map[string]any{
		"action":    "send_item",
		"thread_id": threadID,
		"item_type": "text",
		"text":      text,
	}
	if clientContext != "" {
		payload["client_context"] = clientContext
	}
	return c.PublishJSON(TopicSendMessage, payload)
}

func (c *RealtimeClient) PublishJSON(topic string, data map[string]any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	compressed, err := CompressPayload(b)
	if err != nil {
		return err
	}
	return c.publishBytes(topic, compressed)
}

func (c *RealtimeClient) publishBytes(topic string, payload []byte) error {
	c.mu.Lock()
	c.packetID++
	pid := c.packetID
	c.mu.Unlock()

	packet := WritePublishPacket(topic, payload, 1, pid)
	return c.transport.Send(packet)
}

func (c *RealtimeClient) readLoop() {
	defer func() {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		data, err := c.transport.RecvPacket()
		if err != nil {
			c.emit("error", err)
			return
		}

		pkt, err := DecodePacketData(data)
		if err != nil {
			continue
		}

		switch pkt.PacketType {
		case PacketPublish:
			c.handlePublish(pkt)
			if pkt.QoS == 1 && pkt.PacketID > 0 {
				_ = c.transport.Send(WritePubackPacket(pkt.PacketID))
			}
		case PacketPingresp:
			c.emit("pingresp", nil)
		}
	}
}

func (c *RealtimeClient) handlePublish(pkt *MQTToTPacket) {
	body, err := DecompressPayload(pkt.Payload)
	if err != nil {
		body = pkt.Payload
	}

	c.emit("receive", RealtimeEvent{Topic: pkt.Topic, Payload: body})

	switch pkt.Topic {
	case TopicMessageSync:
		c.dispatchMessageSync(body)
	case TopicRealtimeSub:
		c.dispatchRealtimeSub(body)
	case TopicSendMessageResponse:
		var parsed any
		if json.Unmarshal(body, &parsed) == nil {
			c.emit("send_response", parsed)
		}
	case TopicIrisSubResponse:
		var parsed any
		if json.Unmarshal(body, &parsed) == nil {
			c.emit("iris_sub_response", parsed)
		}
	}
}

func (c *RealtimeClient) dispatchMessageSync(body []byte) {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return
	}

	items, ok := payload.([]any)
	if !ok {
		c.emit("message", payload)
		return
	}

	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		data, ok := m["data"].([]any)
		if !ok {
			c.emit("iris", item)
			continue
		}

		for _, patch := range data {
			p, ok := patch.(map[string]any)
			if !ok {
				continue
			}
			path, _ := p["path"].(string)
			rawValue := p["value"]

			var value any
			if s, ok := rawValue.(string); ok {
				json.Unmarshal([]byte(s), &value)
			} else {
				value = rawValue
			}

			threadID := c.threadIDFromPath(path)

			message := map[string]any{
				"path":      path,
				"op":        p["op"],
				"thread_id": threadID,
			}
			if vm, ok := value.(map[string]any); ok {
				for k, v := range vm {
					message[k] = v
				}
			} else {
				message["value"] = value
			}

			wrapper := make(map[string]any, len(m)+1)
			for k, v := range m {
				if k != "data" {
					wrapper[k] = v
				}
			}
			wrapper["message"] = message

			if strings.HasPrefix(path, "/direct_v2/threads/") {
				c.emit("message", wrapper)
			} else {
				c.emit("thread_update", wrapper)
			}
		}
	}
}

func (c *RealtimeClient) dispatchRealtimeSub(body []byte) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return
	}

	c.emit("realtime_sub", payload)

	message, ok := payload["message"]
	if !ok {
		return
	}

	var directPayload map[string]any
	switch m := message.(type) {
	case string:
		json.Unmarshal([]byte(m), &directPayload)
	case map[string]any:
		if topic, _ := m["topic"].(string); topic != "direct" {
			return
		}
		if dp, ok := m["json"].(string); ok {
			json.Unmarshal([]byte(dp), &directPayload)
		} else if dp, ok := m["payload"].(string); ok {
			json.Unmarshal([]byte(dp), &directPayload)
		} else {
			directPayload, _ = m["json"].(map[string]any)
			if directPayload == nil {
				directPayload, _ = m["payload"].(map[string]any)
			}
		}
	}

	if directPayload == nil {
		return
	}

	data, ok := directPayload["data"].([]any)
	if !ok {
		c.emit("direct", directPayload)
		return
	}

	for _, item := range data {
		if im, ok := item.(map[string]any); ok {
			c.emit("direct", im)
		}
	}
}

func (c *RealtimeClient) threadIDFromPath(path string) string {
	prefix := "/direct_v2/threads/"
	if strings.HasPrefix(path, prefix) {
		rest := path[len(prefix):]
		if idx := strings.Index(rest, "/"); idx > 0 {
			return rest[:idx]
		}
		return rest
	}
	prefix2 := "/direct_v2/inbox/threads/"
	if strings.HasPrefix(path, prefix2) {
		rest := path[len(prefix2):]
		if idx := strings.Index(rest, "/"); idx > 0 {
			return rest[:idx]
		}
		return rest
	}
	return ""
}

func (c *RealtimeClient) emit(event string, payload any) {
	c.mu.RLock()
	handlers := make([]RealtimeHandler, len(c.handlers[event]))
	copy(handlers, c.handlers[event])
	c.mu.RUnlock()

	for _, h := range handlers {
		h(RealtimeEvent{Topic: event, Payload: payload})
	}
}

func (c *RealtimeClient) buildConnection() []byte {
	timestamp := int64(time.Now().UnixNano() / int64(time.Millisecond))
	sessionIDInt, _ := strconv.ParseInt(strings.Split(c.sessionID, "%")[0], 10, 64)

	ci := map[string]any{
		"userId":                    c.userID,
		"userAgent":                 c.userAgent,
		"clientCapabilities":        int64(183),
		"endpointCapabilities":      int64(0),
		"publishFormat":             1,
		"noAutomaticForeground":     false,
		"makeUserAvailableInForeground": true,
		"deviceId":                  c.deviceID,
		"isInitiallyForeground":     true,
		"networkType":               1,
		"networkSubtype":            0,
		"clientMqttSessionId":       timestamp & 0xFFFFFFFF,
		"subscribeTopics":           []int{88, 135, 149, 150, 133, 146},
		"clientType":                "cookie_auth",
		"appId":                     int64(IGRealtimeAppID),
		"deviceSecret":              "",
		"clientStack":               3,
	}

	_ = sessionIDInt

	everclear := map[string]string{
		"inapp_notification_subscribe_comment":                "17899377895239777",
		"inapp_notification_subscribe_comment_mention_and_reply": "17899377895239777",
		"video_call_participant_state_delivery":               "17977239895057311",
		"presence_subscribe":                                  "17846944882223835",
	}
	ecJSON, _ := json.Marshal(everclear)

	locale := c.locale
	if locale == "" {
		locale = "en_US"
	}

	asi := map[string]string{
		"app_version":     c.appVersion,
		"X-IG-Capabilities": c.capabilities,
		"everclear_subscriptions": string(ecJSON),
		"User-Agent":      c.userAgent,
		"Accept-Language":  strings.ReplaceAll(locale, "_", "-"),
		"platform":        "android",
		"ig_mqtt_route":   "django",
		"pubsub_msg_type_blacklist": "direct, typing_type",
		"auth_cache_enabled": "0",
	}

	info := ConnectionInfo{
		ClientIdentifier: c.clientIdentifier,
		ClientInfo:       ci,
		Password:         fmt.Sprintf("sessionid=%s", c.sessionID),
		AppSpecificInfo:  asi,
	}

	return BuildConnectionPayload(info)
}

func (c *RealtimeClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}
