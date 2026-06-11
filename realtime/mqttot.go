// Package realtime implements Instagram's MQTToT protocol for real-time DMs.
package realtime

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	igerr "github.com/motovax/motoig/errors"
)

const (
	ProtocolName  = "MQTToT"
	ProtocolLevel = 3

	PacketConnect    = 1
	PacketConnack    = 2
	PacketPublish    = 3
	PacketPuback     = 4
	PacketSubscribe  = 8
	PacketPingreq    = 12
	PacketPingresp   = 13
	PacketDisconnect = 14

	TopicPubSub              = "88"
	TopicForegroundState     = "102"
	TopicSendMessage         = "132"
	TopicSendMessageResponse = "133"
	TopicIrisSub             = "134"
	TopicIrisSubResponse     = "135"
	TopicMessageSync         = "146"
	TopicRealtimeSub         = "149"
	TopicRegionHint          = "150"
)

type MQTToTTransport struct {
	host    string
	port    int
	timeout time.Duration
	conn    net.Conn
	tlsConn *tls.Conn
	mu      sync.Mutex
}

func NewMQTToTTransport(host string) *MQTToTTransport {
	return &MQTToTTransport{
		host:    host,
		port:    443,
		timeout: 30 * time.Second,
	}
}

func (t *MQTToTTransport) Connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	raw, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", t.host, t.port), t.timeout)
	if err != nil {
		return igerr.Wrap("MQTToT.Connect", "dial", err)
	}

	config := &tls.Config{
		ServerName: t.host,
	}
	tlsConn := tls.Client(raw, config)
	if err := tlsConn.Handshake(); err != nil {
		raw.Close()
		return igerr.Wrap("MQTToT.Connect", "tls handshake", err)
	}

	t.conn = raw
	t.tlsConn = tlsConn
	return nil
}

func (t *MQTToTTransport) Send(packet []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.tlsConn == nil {
		return igerr.New("MQTToT.Send", "not connected")
	}
	_, err := t.tlsConn.Write(packet)
	return err
}

func (t *MQTToTTransport) RecvPacket() ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.tlsConn == nil {
		return nil, igerr.New("MQTToT.RecvPacket", "not connected")
	}
	return readPacket(t.tlsConn)
}

func (t *MQTToTTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.tlsConn != nil {
		err := t.tlsConn.Close()
		t.tlsConn = nil
		t.conn = nil
		return err
	}
	return nil
}

func readPacket(r io.Reader) ([]byte, error) {
	first := make([]byte, 1)
	if _, err := io.ReadFull(r, first); err != nil {
		return nil, err
	}

	var remaining []byte
	for {
		b := make([]byte, 1)
		if _, err := io.ReadFull(r, b); err != nil {
			return nil, err
		}
		remaining = append(remaining, b[0])
		if b[0]&0x80 == 0 {
			break
		}
	}

	size, _ := DecodeRemainingLength(remaining)
	body := make([]byte, size)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}

	result := make([]byte, 0, 1+len(remaining)+size)
	result = append(result, first...)
	result = append(result, remaining...)
	result = append(result, body...)
	return result, nil
}

type MQTToTPacket struct {
	PacketType   byte
	Topic        string
	Payload      []byte
	QoS          byte
	PacketID     uint16
	ReturnCode   byte
	ProtocolName string
}

func DecodePacketData(data []byte) (*MQTToTPacket, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("packet too short")
	}
	p := &MQTToTPacket{
		PacketType: data[0] >> 4,
	}

	switch p.PacketType {
	case PacketPublish:
		p.Topic, p.Payload, p.QoS, p.PacketID = ParsePublishBody(data)
	case PacketConnack:
		if len(data) >= 4 {
			p.ReturnCode = data[3]
		}
	case PacketPingresp:
	}
	return p, nil
}

func ParsePublishBody(data []byte) (string, []byte, byte, uint16) {
	flags := data[0] & 0x0F
	qos := (flags >> 1) & 0x03

	if len(data) < 4 {
		return "", data, 0, 0
	}
	topicLen := int(binary.BigEndian.Uint16(data[2:4]))
	pos := 4
	if pos+topicLen > len(data) {
		return string(data[2:4]), data[pos:], qos, 0
	}
	topic := string(data[pos : pos+topicLen])
	pos += topicLen

	var packetID uint16
	if qos > 0 && pos+2 <= len(data) {
		packetID = binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2
	}

	return topic, data[pos:], qos, packetID
}

func CompressPayload(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecompressPayload(data []byte) ([]byte, error) {
	if len(data) == 0 || data[0] != 0x78 {
		return data, nil
	}
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return data, nil
	}
	defer r.Close()
	return io.ReadAll(r)
}

func WriteConnectPacket(connectionInfo []byte, keepAlive int) []byte {
	compressed, err := CompressPayload(connectionInfo)
	if err != nil {
		compressed = connectionInfo
	}

	protoName := EncodeUTF8String(ProtocolName)
	level := []byte{ProtocolLevel, 0xC2}
	ka := make([]byte, 2)
	binary.BigEndian.PutUint16(ka, uint16(keepAlive))

	var variable []byte
	variable = append(variable, protoName...)
	variable = append(variable, level...)
	variable = append(variable, ka...)
	variable = append(variable, compressed...)

	return BuildPacket(PacketConnect, variable)
}

func WritePublishPacket(topic string, payload []byte, qos byte, packetID uint16) []byte {
	var body []byte
	body = append(body, EncodeUTF8String(topic)...)
	if qos > 0 {
		pid := make([]byte, 2)
		binary.BigEndian.PutUint16(pid, packetID)
		body = append(body, pid...)
	}
	body = append(body, payload...)

	flags := byte(0x30 | (qos << 1))
	return BuildPacketRaw(flags, body)
}

func WriteSubscribePacket(topic string, packetID uint16, qos byte) []byte {
	pid := make([]byte, 2)
	binary.BigEndian.PutUint16(pid, packetID)

	var body []byte
	body = append(body, pid...)
	body = append(body, EncodeUTF8String(topic)...)
	body = append(body, qos)

	return BuildPacketRaw(0x82, body)
}

func WritePingreqPacket() []byte {
	return []byte{0xC0, 0x00}
}

func WriteDisconnectPacket() []byte {
	return []byte{0xE0, 0x00}
}

func WritePubackPacket(packetID uint16) []byte {
	pid := make([]byte, 2)
	binary.BigEndian.PutUint16(pid, packetID)
	return BuildPacket(PacketPuback, pid)
}

func BuildPacket(packetType byte, payload []byte) []byte {
	return BuildPacketRaw(byte(packetType<<4), payload)
}

func BuildPacketRaw(flags byte, payload []byte) []byte {
	remaining := EncodeRemainingLength(len(payload))
	result := make([]byte, 0, 1+len(remaining)+len(payload))
	result = append(result, flags)
	result = append(result, remaining...)
	result = append(result, payload...)
	return result
}

func EncodeUTF8String(s string) []byte {
	b := []byte(s)
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(b)))
	return append(length, b...)
}

func EncodeRemainingLength(length int) []byte {
	var encoded []byte
	for {
		byteVal := length % 128
		length /= 128
		if length > 0 {
			byteVal |= 0x80
		}
		encoded = append(encoded, byte(byteVal))
		if length == 0 {
			break
		}
	}
	return encoded
}

func DecodeRemainingLength(data []byte) (int, int) {
	multiplier := 1
	value := 0
	offset := 0
	for {
		if offset >= len(data) {
			return value, offset
		}
		b := data[offset]
		offset++
		value += int(b&0x7F) * multiplier
		if b&0x80 == 0 {
			break
		}
		multiplier *= 128
	}
	return value, offset
}
