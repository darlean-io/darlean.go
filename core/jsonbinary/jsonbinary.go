package jsonbinary

import (
	"bytes"
	"context"
	"core/binary"
	"core/variant"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
)

const c_JB_MAJOR = '0'
const c_JB_MINOR = '1'
const c_JB_MAGIC = "JB"
const c_JB_TAG = c_JB_MAGIC + string(c_JB_MAJOR) + string(c_JB_MINOR)
const c_JB_HEADER = c_JB_TAG + "\n"

type BytesPool interface {
	Get(size int) []byte
	Put(buf []byte)
}

func Serialize(data any, pool BytesPool) (result []byte, err error) {
	var bufferdata binary.BufferContext
	ctx := binary.NewContext(context.Background(), &bufferdata)
	json, err := json.MarshalContext(ctx, data)
	if err != nil {
		return nil, err
	}
	if len(bufferdata.Buffers) == 0 {
		buf := bytes.NewBufferString(c_JB_HEADER)
		buf.Write(json)
		return buf.Bytes(), nil
	}

	seed := bufferdata.Uid
	totalSize := 0
	header := c_JB_TAG + ";" + seed + ";" + strconv.FormatInt(int64(len(json)), 10) + ";"
	for i, buf := range bufferdata.Buffers {
		if i > 0 {
			header = header + ","
		}
		header = header + strconv.FormatInt(int64(len(buf)), 10)
		totalSize += len(buf) + 1
	}
	header = header + "\n"

	headerBuf := []byte(header)
	totalSize += len(headerBuf) + len(json) + 1
	//var resultbuf bytes.Buffer
	//fmt.Printf("TotalSize %v", totalSize)
	if pool == nil {
		result = make([]byte, totalSize)
	} else {
		result = pool.Get(totalSize)
		defer func() {
			if r := recover(); r != nil {
				pool.Put(result)
				err = fmt.Errorf("jsonbinary: unexpected panic: %v", r)
			}
		}()
	}
	copy(result, headerBuf)
	copy(result[len(headerBuf):], json)
	offset := len(headerBuf) + len(json)
	copy(result[len(headerBuf)+len(json):], []byte{'\n'})
	offset += 1
	for _, buffer := range bufferdata.Buffers {
		copy(result[offset:], buffer)
		copy(result[offset+len(buffer):], []byte{'\n'})
		offset += len(buffer) + 1
	}
	return
}

func Deserialize(data []byte, value any) error {
	buf := bytes.NewBuffer(data)
	header, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.HasPrefix(header, c_JB_MAGIC) {
		return fmt.Errorf("jsonbinary: invalid magic: %s", header[0:2])
	}
	if len(header) < len(c_JB_HEADER) {
		return errors.New("jsonbinary: invalid header")
	}
	if header[2] > c_JB_MAJOR {
		return fmt.Errorf("jsonbinary: unsupported major version: %s", header[2:3])
	}

	if len(header) == len(c_JB_HEADER) {
		// No buffers
		return json.Unmarshal(buf.Bytes(), value)
	}

	parts := strings.Split(header[:len(header)-1], ";")
	if len(parts) == 0 {
		return errors.New("jsonbinary: invalid header")
	}
	if len(parts) >= 3 {
		lengths := strings.Split(parts[3], ",")

		if len(lengths) > 0 {
			jsonLen, err := strconv.ParseInt(parts[2], 10, 32)
			if err != nil {
				return err
			}
			offset := len(header) + int(jsonLen) + 1
			buffers := make([][]byte, len(lengths))
			for i, length := range lengths {
				lengthInt, err := strconv.ParseInt(length, 10, 32)
				if err != nil {
					return err
				}
				buffer := data[offset : offset+int(lengthInt)]
				buffers[i] = buffer
				offset = offset + int(lengthInt) + 1
			}
			bufferdata := binary.BufferContext{
				Uid:     parts[1],
				Buffers: buffers,
			}
			var temp any
			// Unmarshall to new "any" so that we receive plain maps instead of filled-in structs
			// that are harder to iterate
			offset = len(header)
			err = json.Unmarshal(data[offset:offset+int(jsonLen)], &temp)
			if err != nil {
				return err
			}
			// Find all buffer occurrances and add the "data" to them
			analyzeBufferStructs(&bufferdata, temp)
			// Try to map the plain structure onto the result value
			variant.Assign(temp, value)
			return nil
		}
	}
	return json.Unmarshal(buf.Bytes(), value)
}

func analyzeBufferStructs(bufferData *binary.BufferContext, value any) {
	m, ok := value.(map[string]any)
	if ok {
		b, has := m["__b"]
		if has {
			if b == bufferData.Uid {
				// Bingo, we have a legitimate buffer

				// Check whether we have an index (introduced in JB01). If not, fall back
				// to the original behaviour of processen buffers in order. But this may not
				// be ok because go shuffles map items.
				idx, has := m["i"]
				if has {
					m[binary.INTERNAL_FIELD] = bufferData.Buffers[int(idx.(float64))]
				} else {
					m[binary.INTERNAL_FIELD] = bufferData.Buffers[0]
					bufferData.Buffers = bufferData.Buffers[1:]
				}
			}
			return
		}
		for _, v := range m {
			analyzeBufferStructs(bufferData, v)
		}
	}
	a, ok := value.([]any)
	if ok {
		for _, v := range a {
			analyzeBufferStructs(bufferData, v)
		}
	}
}
