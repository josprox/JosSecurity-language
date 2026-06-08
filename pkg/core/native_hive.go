package core

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// Hive binary type IDs (from Hive's FrameHelper source)
const (
	hiveTypeNull     = 0
	hiveTypeInt      = 1
	hiveTypeDouble   = 2
	hiveTypeBoolTrue = 5
	hiveTypeBoolFalse = 6
	hiveTypeString   = 7
	hiveTypeBytes    = 8
	hiveTypeList     = 9
	hiveTypeMap      = 10
	hiveTypeHiveList = 11
	hiveTypeSmallInt = 16
)

type hiveReader struct {
	data   []byte
	offset int
}

func (r *hiveReader) remaining() int {
	return len(r.data) - r.offset
}

func (r *hiveReader) readByte() byte {
	if r.offset >= len(r.data) {
		return 0
	}
	b := r.data[r.offset]
	r.offset++
	return b
}

func (r *hiveReader) readUint32() uint32 {
	if r.offset+4 > len(r.data) {
		return 0
	}
	v := binary.LittleEndian.Uint32(r.data[r.offset:])
	r.offset += 4
	return v
}

func (r *hiveReader) readInt64() int64 {
	if r.offset+8 > len(r.data) {
		return 0
	}
	v := int64(binary.LittleEndian.Uint64(r.data[r.offset:]))
	r.offset += 8
	return v
}

func (r *hiveReader) readFloat64() float64 {
	if r.offset+8 > len(r.data) {
		return 0
	}
	bits := binary.LittleEndian.Uint64(r.data[r.offset:])
	r.offset += 8
	return math.Float64frombits(bits)
}

func (r *hiveReader) readString() string {
	length := r.readUint32()
	if r.offset+int(length) > len(r.data) {
		return ""
	}
	s := string(r.data[r.offset : r.offset+int(length)])
	r.offset += int(length)
	return s
}

func (r *hiveReader) readValue() interface{} {
	if r.remaining() <= 0 {
		return nil
	}
	typeId := r.readByte()

	switch typeId {
	case hiveTypeNull:
		return nil
	case hiveTypeInt:
		return r.readInt64()
	case hiveTypeSmallInt:
		if r.offset+4 > len(r.data) {
			return 0
		}
		v := int32(binary.LittleEndian.Uint32(r.data[r.offset:]))
		r.offset += 4
		return int64(v)
	case hiveTypeDouble:
		return r.readFloat64()
	case hiveTypeBoolTrue:
		return true
	case hiveTypeBoolFalse:
		return false
	case hiveTypeString:
		return r.readString()
	case hiveTypeBytes:
		length := r.readUint32()
		if r.offset+int(length) > len(r.data) {
			return nil
		}
		bs := make([]byte, length)
		copy(bs, r.data[r.offset:r.offset+int(length)])
		r.offset += int(length)
		return bs
	case hiveTypeList:
		count := r.readUint32()
		list := make([]interface{}, 0, count)
		for i := uint32(0); i < count; i++ {
			list = append(list, r.readValue())
		}
		return list
	case hiveTypeMap:
		count := r.readUint32()
		m := make(map[string]interface{})
		for i := uint32(0); i < count; i++ {
			key := r.readValue()
			val := r.readValue()
			m[fmt.Sprintf("%v", key)] = val
		}
		return m
	case hiveTypeHiveList:
		// HiveList: boxName string + list of keys
		r.readString() // skip box name
		count := r.readUint32()
		for i := uint32(0); i < count; i++ {
			r.readValue() // skip keys
		}
		return nil
	default:
		fmt.Printf("[HiveReader] Unknown typeId: %d at offset %d\n", typeId, r.offset)
		return nil
	}
}

// ReadHiveBox reads a .hive binary file and returns all non-deleted key-value pairs.
// Returns a slice of maps: [{"key": "...", "value": {...}}]
func ReadHiveBox(filePath string) ([]map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("[HiveReader] cannot read file %s: %w", filePath, err)
	}

	fmt.Printf("[HiveReader] Reading %s (%d bytes)\n", filePath, len(data))

	var results []map[string]interface{}
	offset := 0

	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		// Frame: [byteCount uint32][checksum uint32][body...]
		// byteCount includes the 8-byte header itself
		byteCount := int(binary.LittleEndian.Uint32(data[offset:]))

		if byteCount == 0 {
			// EOF/empty frame
			break
		}

		if byteCount < 8 || offset+byteCount > len(data) {
			// Malformed frame
			fmt.Printf("[HiveReader] Malformed frame at offset %d (byteCount=%d)\n", offset, byteCount)
			break
		}

		frameBody := data[offset+8 : offset+byteCount]
		offset += byteCount

		r := &hiveReader{data: frameBody}
		key := r.readValue()
		value := r.readValue()

		// Skip tombstone/deleted entries (value == nil after a valid key)
		if key == nil {
			continue
		}
		if value == nil {
			// Deleted entry — skip
			continue
		}

		results = append(results, map[string]interface{}{
			"key":   fmt.Sprintf("%v", key),
			"value": value,
		})
	}

	fmt.Printf("[HiveReader] Parsed %d entries from %s\n", len(results), filePath)
	return results, nil
}
