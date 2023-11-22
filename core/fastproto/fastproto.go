package fastproto

import (
	"bytes"
	"core/anny"
	"core/jsonbinary"
	"strconv"
)

type JsonAny struct {
	value []byte
}

func (anny *JsonAny) Get(value any) (any, error) {
	err := jsonbinary.Deserialize(anny.value, value)
	return value, err
}

func newJsonAny(value []byte) anny.Anny {
	if value == nil || len(value) == 0 {
		return nil
	}
	anny := JsonAny{
		value: value,
	}
	return &anny
}

const CHAR_CODE_ZERO_DIGITS = 'a'
const CHAR_CODE_BUFFER = 'b'
const CHAR_CODE_FALSE = 'f'
const CHAR_CODE_JSON = 'j'
const CHAR_CODE_NUMBER = 'n'
const CHAR_CODE_STRING = 's'
const CHAR_CODE_TRUE = 't'
const CHAR_CODE_UNDEFINED = 'u'

func WriteUnsignedInt(buf *bytes.Buffer, value int) error {
	if value == 0 {
		return buf.WriteByte(CHAR_CODE_ZERO_DIGITS)
	}
	str := strconv.FormatUint(uint64(value), 10)
	strlen := len(str)
	buf.WriteByte(CHAR_CODE_ZERO_DIGITS + byte(strlen))
	buf.WriteString(str)
	return nil
}

func ReadUnsignedInt(buf *bytes.Buffer) (uint32, error) {
	lenbyte, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}
	if lenbyte == CHAR_CODE_ZERO_DIGITS {
		return 0, nil
	}
	strlen := lenbyte - CHAR_CODE_ZERO_DIGITS
	localbuf := make([]byte, strlen)
	_, err = buf.Read(localbuf)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseUint(string(localbuf), 10, 32)
	return uint32(parsed), err
}

func WriteString(buf *bytes.Buffer, value *string) error {
	if (value == nil) || (len(*value) == 0) {
		return WriteUnsignedInt(buf, 0)
	}
	WriteUnsignedInt(buf, len(*value))
	buf.WriteString(*value)
	return nil
}

func ReadString(buf *bytes.Buffer) (*string, error) {
	strlen, err := ReadUnsignedInt(buf)
	if err != nil {
		return nil, err
	}
	localbuf := make([]byte, strlen)
	_, err = buf.Read(localbuf)
	if err != nil {
		return nil, err
	}
	asString := string(localbuf)
	return &asString, nil
}

func WriteChar(buf *bytes.Buffer, value int) error {
	return buf.WriteByte(byte(value))
}

func ReadChar(buf *bytes.Buffer) (int, error) {
	value, err := buf.ReadByte()
	return int(value), err
}

func WriteBinary(buf *bytes.Buffer, value *[]byte) error {
	if (value == nil) || (len(*value) == 0) {
		return WriteUnsignedInt(buf, 0)
	}
	WriteUnsignedInt(buf, len(*value))
	buf.Write(*value)
	return nil
}

func ReadBinary(buf *bytes.Buffer) (*[]byte, error) {
	strlen, err := ReadUnsignedInt(buf)
	if err != nil {
		return nil, err
	}
	localbuf := make([]byte, strlen)
	_, err = buf.Read(localbuf)
	return &localbuf, err
}

func WriteJson(buf *bytes.Buffer, value any) error {
	if value == nil {
		return WriteBinary(buf, nil)
	}
	serialized, err := jsonbinary.Serialize(value)
	if err != nil {
		return err
	}
	WriteBinary(buf, &serialized)
	return nil
}

func ReadJson(buf *bytes.Buffer) (anny.Anny, error) {
	data, err := ReadBinary(buf)
	if err != nil {
		return nil, err
	}
	return newJsonAny(*data), nil
}

func WriteVariant(buf *bytes.Buffer, value any) error {
	switch v := (value).(type) {
	case nil:
		return WriteChar(buf, CHAR_CODE_UNDEFINED)
	case int64:
		WriteChar(buf, CHAR_CODE_NUMBER)
		str := strconv.FormatInt(v, 10)
		return WriteString(buf, &str)
	case float64:
		WriteChar(buf, CHAR_CODE_NUMBER)
		str := strconv.FormatFloat(v, 'e', 15, 64)
		return WriteString(buf, &str)
	case string:
		WriteChar(buf, CHAR_CODE_STRING)
		return WriteString(buf, &v)
	case bool:
		if v {
			return WriteChar(buf, CHAR_CODE_TRUE)
		}
		return WriteChar(buf, CHAR_CODE_FALSE)
	case bytes.Buffer:
		WriteChar(buf, CHAR_CODE_BUFFER)
		b := v.Bytes()
		return WriteBinary(buf, &b)
	default:
		WriteChar(buf, CHAR_CODE_JSON)
		return WriteJson(buf, v)
	}
}

func ReadVariant(buf *bytes.Buffer) (anny.Anny, error) {
	kind, err := ReadChar(buf)
	if err != nil {
		return nil, err
	}
	switch kind {
	case CHAR_CODE_UNDEFINED:
		return anny.New(nil), nil
	case CHAR_CODE_STRING:
		val, err := ReadString(buf)
		return anny.New(*val), err
	case CHAR_CODE_NUMBER:
		str, err := ReadString(buf)
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseFloat(*str, 64)
		return anny.New(value), nil
	case CHAR_CODE_JSON:
		return ReadJson(buf)
	case CHAR_CODE_FALSE:
		return anny.New(false), nil
	case CHAR_CODE_TRUE:
		return anny.New(true), nil
	case CHAR_CODE_BUFFER:
		val, err := ReadBinary(buf)
		return anny.New(*val), err
	default:
		panic("Not supported")
	}
}
