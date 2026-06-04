package socketmodels

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

type CommandType int
type MessageType int
type ReportType int

func PackInt32(n int, data []byte) []byte {
	buf := make([]byte, 4)
	buf[0] = byte(n >> 24)
	buf[1] = byte(n >> 16)
	buf[2] = byte(n >> 8)
	buf[3] = byte(n)
	merged := append(data, buf...)
	return merged
}

func PackUInt32(n uint32, data []byte) []byte {
	buf := make([]byte, 4)
	buf[0] = byte(n >> 24)
	buf[1] = byte(n >> 16)
	buf[2] = byte(n >> 8)
	buf[3] = byte(n)
	merged := append(data, buf...)
	return merged
}

func PackInt64(n int64, data []byte) []byte {
	buf := make([]byte, 8)
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)
	buf = append(data, buf...)
	return buf
}

func PackUInt64(n uint64, data []byte) []byte {
	buf := make([]byte, 8)
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)
	buf = append(data, buf...)
	return buf
}

func PackVarUInt32(n uint32, data []byte) []byte {
	buf := make([]byte, 5)
	num := binary.PutUvarint(buf, uint64(n))
	buf = append(data, buf[:num]...)
	return buf
}

func PrependPackVarUInt32(n uint32, data []byte) []byte {
	buf := make([]byte, 5)
	num := binary.PutUvarint(buf, uint64(n))
	buf = append(buf[:num], data...)
	return buf
}

func PrependPackUInt8(n uint8, data []byte) []byte {
	buf := make([]byte, 1)
	buf[0] = byte(n)
	buf = append(buf, data...)
	return buf
}

func PackVarUInt64(n uint64, data []byte) []byte {
	buf := make([]byte, 10)
	num := binary.PutUvarint(buf, n)
	buf = append(data, buf[:num]...)
	return buf
}

func PackVarInt64(n int64, data []byte) []byte {
	buf := make([]byte, 10)
	num := binary.PutVarint(buf, n)
	buf = append(data, buf[:num]...)
	return buf
}

func PackArrayVarInt64(n []int64, data []byte) []byte {
	var length uint16
	length = uint16(len(n))
	if length > 65535 {
		return data
	}
	data = PackVarUInt(uint32(length), data)
	for i := 0; uint16(i) < length; i++ {
		data = PackVarInt64(n[i], data)
	}
	return data
}

func PackArrayString(s []string, data []byte) []byte {
	var length uint16
	length = uint16(len(s))
	if length > 65535 {
		return data
	}
	data = PackVarUInt(uint32(length), data)
	for i := 0; uint16(i) < length; i++ {
		data = PackString(s[i], data)
	}
	return data
}

func PackVarInt(n int32, data []byte) []byte {
	buf := make([]byte, 5)
	num := binary.PutVarint(buf, int64(n))
	buf = append(data, buf[:num]...)
	return buf
}

func PackVarUInt(n uint32, data []byte) []byte {
	buf := make([]byte, 5)
	num := binary.PutUvarint(buf, uint64(n))
	buf = append(data, buf[:num]...)
	return buf
}

func PackBool(val bool, data []byte) []byte {
	if val {
		return append(data, 1)
	}
	return append(data, 0)
}

func PackInt16(n int16, data []byte) []byte {
	buf := make([]byte, 2)
	buf[0] = byte(n >> 8)
	buf[1] = byte(n)
	buf = append(data, buf...)
	return buf
}

func PackInt8(n int8, data []byte) []byte {
	buf := make([]byte, 1)
	buf[0] = byte(n)
	buf = append(data, buf...)
	return buf
}

func PackUInt8(n uint8, data []byte) []byte {
	buf := make([]byte, 1)
	buf[0] = byte(n)
	buf = append(data, buf...)
	return buf
}

func PackFloat32(f float32, data []byte) []byte {
	n := math.Float32bits(f)
	buf := make([]byte, 4)
	buf[0] = byte(n >> 24)
	buf[1] = byte(n >> 16)
	buf[2] = byte(n >> 8)
	buf[3] = byte(n)
	buf = append(data, buf...)
	return buf
}

func PackFloat64(f float64, data []byte) []byte {
	buf := make([]byte, 8)
	n := math.Float64bits(f)
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)
	buf = append(data, buf...)
	return buf
}

func PackBytes(src []byte, data []byte) []byte {
	return append(data, src...)
}

func GetStringLength(s string) int {
	return len(s) + 4
}

func UnPackInt32(n []byte) (int32, []byte) {
	var val int32
	data := n[:4]
	n = n[4:]
	val = int32(data[0]) << 24
	val |= int32(data[1]) << 16
	val |= int32(data[2]) << 8
	val |= int32(data[3])
	return val, n
}
func UnPackUInt32(n []byte) (uint32, []byte) {
	var val uint32
	data := n[:4]
	n = n[4:]
	val = uint32(data[0]) << 24
	val |= uint32(data[1]) << 16
	val |= uint32(data[2]) << 8
	val |= uint32(data[3])
	return val, n
}

func UnPackInt64(n []byte) (int64, []byte) {
	var val int64
	data := n[:8]
	n = n[8:]
	val = int64(data[0]) << 56
	val |= int64(data[1]) << 48
	val |= int64(data[2]) << 40
	val |= int64(data[3]) << 32
	val |= int64(data[4]) << 24
	val |= int64(data[5]) << 16
	val |= int64(data[6]) << 8
	val |= int64(data[7])
	return val, n
}

func UnPackUInt64(n []byte) (uint64, []byte) {
	var val uint64
	data := n[:8]
	n = n[8:]
	val = uint64(data[0]) << 56
	val |= uint64(data[1]) << 48
	val |= uint64(data[2]) << 40
	val |= uint64(data[3]) << 32
	val |= uint64(data[4]) << 24
	val |= uint64(data[5]) << 16
	val |= uint64(data[6]) << 8
	val |= uint64(data[7])
	return val, n
}

func UnPackBool(n []byte) (string, []byte) {
	data := n[:1]
	n = n[1:]
	val := uint8(data[0])
	if val == 0 {
		return "0", n
	} else {
		return "1", n
	}
}

func UnPackBoolValue(n []byte) (bool, []byte) {
	data := n[:1]
	n = n[1:]
	val := uint8(data[0])
	if val == 0 {
		return false, n
	} else {
		return true, n
	}
}

func PackString(val string, data []byte) []byte {
	var length uint32
	length = uint32(len(val))
	if length > 65535 {
		return data
	}
	data = PackVarUInt(length, data)
	data = append(data, []byte(val)...)
	return data
}

func PackStringArray(val []string, data []byte) []byte {
	var length uint32
	length = uint32(len(val))
	if length > 65535 {
		return data
	}
	data = PackVarUInt(length, data)
	// data = append(data, []byte(val)...)
	return data
}

func UnPackString(data []byte) (string, []byte) {
	var length uint32
	length, data = UnPackVarUInt32(data)
	buf := data[:length]
	data = data[length:]
	s := string(buf)
	return s, data
}

func UnPackWebString(data []byte) (string, []byte) {
	var length uint32
	length, data = UnPackUInt32(data)
	buf := data[:length]
	data = data[length:]
	s := string(buf)
	return s, data
}

func UnPackArrayString(data []byte) ([]string, []byte) {
	var length uint32
	length, data = UnPackVarUInt32(data)
	var array []string
	var val string
	for i := 0; uint32(i) < length; i++ {
		val, data = UnPackString(data)
		array = append(array, val)
	}
	return array, data
}

func UnPackArrayVarInt64(data []byte) ([]int64, []byte) {
	var length uint32
	length, data = UnPackVarUInt32(data)

	var array []int64
	var val int64
	for i := 0; uint32(i) < length; i++ {
		val, data = UnPackVarInt64(data)
		array = append(array, val)
	}
	return array, data
}

func UnPackArrayVarString(data []byte) ([]string, []byte) {
	var length uint32
	length, data = UnPackVarUInt32(data)

	var array []string
	var val string
	for i := 0; uint32(i) < length; i++ {
		val, data = UnPackString(data)
		array = append(array, val)
	}
	return array, data
}

func UnPackInt16(n []byte) (int16, []byte) {
	var val int16
	data := n[:2]
	n = n[2:]
	val = int16(data[0]) << 8
	val |= int16(data[1])
	return val, n
}

func UnPackUInt16(n []byte) (uint16, []byte) {
	var val uint16
	data := n[:2]
	n = n[2:]
	val = uint16(data[0]) << 8
	val |= uint16(data[1])
	return val, n
}

func UnPackInt8(n []byte) (int8, []byte) {
	var val int8
	data := n[:1]
	n = n[1:]
	val = int8(data[0])
	return val, n
}

func UnPackUInt8(n []byte) (uint8, []byte) {
	var val uint8
	data := n[:1]
	n = n[1:]
	val = uint8(data[0])
	return val, n
}

func UnPackFloat32(n []byte) (float32, []byte) {
	var val uint32
	data := n[:4]
	n = n[4:]
	val = uint32(data[0]) << 24
	val |= uint32(data[1]) << 16
	val |= uint32(data[2]) << 8
	val |= uint32(data[3])
	return math.Float32frombits(val), n
}

func UnPackFloat64(n []byte) (float64, []byte) {
	var val uint64
	data := n[:8]
	n = n[8:]
	val = uint64(data[0]) << 56
	val |= uint64(data[1]) << 48
	val |= uint64(data[2]) << 40
	val |= uint64(data[3]) << 32
	val |= uint64(data[4]) << 24
	val |= uint64(data[5]) << 16
	val |= uint64(data[6]) << 8
	val |= uint64(data[7])
	return math.Float64frombits(val), n
}

func UnPackBytes(src []byte, len int) []byte {
	return src[:len]
}

func UnPackVarUInt32(n []byte) (uint32, []byte) {
	num, data := binary.Uvarint(n)
	n = n[data:]
	return uint32(num), n
}

func UnPackVarUInt32SameData(n []byte) uint32 {
	num, _ := binary.Uvarint(n)
	return uint32(num)
}

func UnPackVarInt(n []byte) (int32, []byte) {
	num, data := binary.Varint(n)
	n = n[data:]
	return int32(num), n
}

func UnPackVarUInt64(n []byte) (uint64, []byte) {
	num, data := binary.Uvarint(n)
	n = n[data:]
	return num, n
}

func UnPackVarInt64(n []byte) (int64, []byte) {
	num, data := binary.Varint(n)
	n = n[data:]
	return num, n
}
func CreateCommand(message MessageType, command CommandType) uint8 {
	return uint8(command)<<4 | uint8(message)
}

func TranslateCommand(command uint8) (MessageType, CommandType) {
	return MessageType(command & 0xF), CommandType(command >> 4)
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func ArrayToString(a []int64) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", ",", -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}
func StringArrayToString(a []string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", ",", -1), "[]")
}
func Btoi(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Itob(val int32) bool {
	if val > 0 {
		return true
	}
	return false
}
