package realtime

import (
	"encoding/binary"
	"fmt"
)

const (
	ThriftStop    = 0x00
	ThriftTrue    = 0x01
	ThriftFalse   = 0x02
	ThriftByte    = 0x03
	ThriftInt16   = 0x04
	ThriftInt32   = 0x05
	ThriftInt64   = 0x06
	ThriftBinary  = 0x08
	ThriftList    = 0x09
	ThriftMap     = 0x0B
	ThriftStruct  = 0x0C
	ThriftBoolean = 0xA1

	ThriftListInt32       = (ThriftInt32 << 8) | ThriftList
	ThriftListBinary      = (ThriftBinary << 8) | ThriftList
	ThriftMapBinaryBinary = (ThriftBinary << 8) | ThriftMap
)

type ThriftField struct {
	Name     string
	ID       int
	Type     int
	Children []ThriftField
}

func WriteThriftObject(data map[string]any, descriptors []ThriftField) []byte {
	w := &thriftWriter{buf: make([]byte, 0, 512)}
	writeThriftStruct(w, data, descriptors)
	w.buf = append(w.buf, ThriftStop)
	return w.buf
}

type thriftWriter struct {
	buf        []byte
	fieldStack []int
	field      int
}

func (w *thriftWriter) writeField(fieldID int, thriftType int) {
	delta := fieldID - w.field
	if delta > 0 && delta <= 15 {
		w.buf = append(w.buf, byte((delta<<4)|thriftType))
	} else {
		w.buf = append(w.buf, byte(thriftType))
		w.writeVarint(zigzag(fieldID, 16))
	}
	w.field = fieldID
}

func (w *thriftWriter) writeVarint(value int) {
	for {
		if value&^0x7F == 0 {
			w.buf = append(w.buf, byte(value))
			return
		}
		w.buf = append(w.buf, byte((value&0x7F)|0x80))
		value >>= 7
	}
}

func (w *thriftWriter) writeStringDirect(value string) {
	raw := []byte(value)
	w.writeVarint(len(raw))
	w.buf = append(w.buf, raw...)
}

func (w *thriftWriter) writeString(fieldID int, value string) {
	w.writeField(fieldID, ThriftBinary)
	w.writeStringDirect(value)
}

func (w *thriftWriter) writeBool(fieldID int, value bool) {
	if value {
		w.writeField(fieldID, ThriftTrue)
	} else {
		w.writeField(fieldID, ThriftFalse)
	}
}

func (w *thriftWriter) writeInt64(fieldID int, value int64) {
	w.writeField(fieldID, ThriftInt64)
	w.writeVarint(zigzag(int(value), 64))
}

func (w *thriftWriter) writeInt32(fieldID int, value int) {
	w.writeField(fieldID, ThriftInt32)
	w.writeVarint(zigzag(value, 32))
}

func (w *thriftWriter) writeByte(fieldID int, value int) {
	w.writeField(fieldID, ThriftByte)
	w.buf = append(w.buf, byte(int8(value)))
}

func (w *thriftWriter) pushStruct(fieldID int) {
	w.writeField(fieldID, ThriftStruct)
	w.fieldStack = append(w.fieldStack, w.field)
	w.field = 0
}

func (w *thriftWriter) writeStop() {
	w.buf = append(w.buf, ThriftStop)
	if len(w.fieldStack) > 0 {
		w.field = w.fieldStack[len(w.fieldStack)-1]
		w.fieldStack = w.fieldStack[:len(w.fieldStack)-1]
	}
}

func writeThriftStruct(w *thriftWriter, data map[string]any, descriptors []ThriftField) {
	for _, d := range descriptors {
		val, ok := data[d.Name]
		if !ok || val == nil {
			continue
		}
		thriftType := d.Type & 0xFF
		switch thriftType {
		case ThriftBoolean:
			w.writeBool(d.ID, val.(bool))
		case ThriftByte:
			w.writeByte(d.ID, toInt(val))
		case ThriftInt32:
			w.writeInt32(d.ID, toInt(val))
		case ThriftInt64:
			w.writeInt64(d.ID, toInt64(val))
		case ThriftBinary:
			w.writeString(d.ID, fmt.Sprintf("%v", val))
		case ThriftStruct:
			w.pushStruct(d.ID)
			if m, ok := val.(map[string]any); ok {
				writeThriftStruct(w, m, d.Children)
			}
			w.writeStop()
		case ThriftList:
			w.writeField(d.ID, ThriftList)
			itemType := d.Type >> 8
			writeThriftList(w, itemType, val)
		case ThriftMap:
			writeThriftMap(w, d.ID, val)
		default:
			panic(fmt.Sprintf("unsupported thrift type: %d", d.Type))
		}
	}
}

func writeThriftList(w *thriftWriter, itemType int, val any) {
	items := toSlice(val)
	size := len(items)
	if size < 0x0F {
		w.buf = append(w.buf, byte((size<<4)|itemType))
	} else {
		w.buf = append(w.buf, byte(0xF0|itemType))
		w.writeVarint(size)
	}
	for _, item := range items {
		switch itemType {
		case ThriftInt32:
			w.writeVarint(zigzag(toInt(item), 32))
		case ThriftBinary:
			w.writeStringDirect(fmt.Sprintf("%v", item))
		default:
			panic(fmt.Sprintf("unsupported thrift list type: %d", itemType))
		}
	}
}

func writeThriftMap(w *thriftWriter, fieldID int, val any) {
	w.writeField(fieldID, ThriftMap)
	m := toStringMap(val)
	w.writeVarint(len(m))
	if len(m) == 0 {
		return
	}
	w.buf = append(w.buf, byte((ThriftBinary<<4)|ThriftBinary))
	for key, value := range m {
		w.writeStringDirect(key)
		w.writeStringDirect(value)
	}
}

func toInt(val any) int {
	switch v := val.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func toInt64(val any) int64 {
	switch v := val.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func toSlice(val any) []any {
	switch v := val.(type) {
	case []int:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = item
		}
		return out
	case []any:
		return v
	default:
		return nil
	}
}

func toStringMap(val any) map[string]string {
	switch v := val.(type) {
	case map[string]string:
		return v
	case map[string]any:
		out := make(map[string]string, len(v))
		for key, value := range v {
			out[key] = fmt.Sprintf("%v", value)
		}
		return out
	default:
		return nil
	}
}

func zigzag(value, bits int) int {
	return (value << 1) ^ (value >> (bits - 1))
}

func Zigzag64(value int64) int64 {
	return (value << 1) ^ (value >> 63)
}

func VarintZigzag(value int) []byte {
	z := zigzag(value, 32)
	return encodeVarint(z)
}

func encodeVarint(value int) []byte {
	var buf []byte
	for {
		if value&^0x7F == 0 {
			buf = append(buf, byte(value))
			return buf
		}
		buf = append(buf, byte((value&0x7F)|0x80))
		value >>= 7
	}
}

type ConnectionInfo struct {
	ClientIdentifier string
	ClientInfo       map[string]any
	Password         string
	AppSpecificInfo  map[string]string
}

var clientInfoDescriptors = []ThriftField{
	{Name: "userId", ID: 1, Type: ThriftInt64},
	{Name: "userAgent", ID: 2, Type: ThriftBinary},
	{Name: "clientCapabilities", ID: 3, Type: ThriftInt64},
	{Name: "endpointCapabilities", ID: 4, Type: ThriftInt64},
	{Name: "publishFormat", ID: 5, Type: ThriftInt32},
	{Name: "noAutomaticForeground", ID: 6, Type: ThriftBoolean},
	{Name: "makeUserAvailableInForeground", ID: 7, Type: ThriftBoolean},
	{Name: "deviceId", ID: 8, Type: ThriftBinary},
	{Name: "isInitiallyForeground", ID: 9, Type: ThriftBoolean},
	{Name: "networkType", ID: 10, Type: ThriftInt32},
	{Name: "networkSubtype", ID: 11, Type: ThriftInt32},
	{Name: "clientMqttSessionId", ID: 12, Type: ThriftInt64},
	{Name: "clientIpAddress", ID: 13, Type: ThriftBinary},
	{Name: "subscribeTopics", ID: 14, Type: ThriftListInt32},
	{Name: "clientType", ID: 15, Type: ThriftBinary},
	{Name: "appId", ID: 16, Type: ThriftInt64},
	{Name: "deviceSecret", ID: 20, Type: ThriftBinary},
	{Name: "clientStack", ID: 21, Type: ThriftByte},
}

var connectionDescriptors = []ThriftField{
	{Name: "clientIdentifier", ID: 1, Type: ThriftBinary},
	{Name: "willTopic", ID: 2, Type: ThriftBinary},
	{Name: "willMessage", ID: 3, Type: ThriftBinary},
	{Name: "clientInfo", ID: 4, Type: ThriftStruct, Children: clientInfoDescriptors},
	{Name: "password", ID: 5, Type: ThriftBinary},
	{Name: "getDiffsRequests", ID: 6, Type: ThriftListBinary},
	{Name: "zeroRatingTokenHash", ID: 9, Type: ThriftBinary},
	{Name: "appSpecificInfo", ID: 10, Type: ThriftMapBinaryBinary},
}

func BuildConnectionPayload(info ConnectionInfo) []byte {
	data := map[string]any{
		"clientIdentifier": info.ClientIdentifier,
		"clientInfo":       info.ClientInfo,
		"password":         info.Password,
	}
	if len(info.AppSpecificInfo) > 0 {
		data["appSpecificInfo"] = info.AppSpecificInfo
	}
	return WriteThriftObject(data, connectionDescriptors)
}

func BuildConnectionPacket(info ConnectionInfo) []byte {
	payload := BuildConnectionPayload(info)
	return WriteConnectPacket(payload, 20)
}

// ParseListInt32 parses a Thrift list of INT32 values.
func ParseListInt32(data []byte) ([]int, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("list too short")
	}
	header := data[0]
	size := int(header >> 4)
	itemType := header & 0x0F
	pos := 1

	if size == 0x0F {
		var n int
		n, pos = decodeVarintAt(data, pos)
		size = n
	}

	if itemType != ThriftInt32 {
		return nil, fmt.Errorf("unexpected item type: %d", itemType)
	}

	result := make([]int, 0, size)
	for i := 0; i < size && pos < len(data); i++ {
		var val int
		val, pos = decodeVarintAt(data, pos)
		result = append(result, val)
	}
	return result, nil
}

func decodeVarintAt(data []byte, pos int) (int, int) {
	result := 0
	shift := 0
	for pos < len(data) {
		b := data[pos]
		pos++
		result |= int(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return result, pos
}

// ParseBinaryField parses a Thrift binary field from raw payload.
func ParseBinaryField(data []byte) (string, int) {
	if len(data) < 3 {
		return "", 0
	}
	size := int(binary.BigEndian.Uint16(data[0:2]))
	if 2+size > len(data) {
		return string(data[2:]), len(data)
	}
	return string(data[2 : 2+size]), 2 + size
}