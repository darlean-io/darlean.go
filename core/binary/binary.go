package binary

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
)

type BufferContext struct {
	Uid     string
	Buffers [][]byte
}

type bufferContextKeyType int

const INTERNAL_FIELD = "Internal"

const bufferContextKey = bufferContextKeyType(0)

func NewContext(ctx context.Context, data *BufferContext) context.Context {
	return context.WithValue(ctx, bufferContextKey, data)
}

func FromContext(ctx context.Context) *BufferContext {
	return ctx.Value(bufferContextKey).(*BufferContext)
}

type Binary struct {
	// For internal use. To access the byte data, use one of the
	// accessor functions. Must unfortunately be public to allow
	// JSON serialization and cloning.
	Internal []byte
}

func FromBytes(data []byte) Binary {
	return Binary{
		Internal: data,
	}
}

func FromBuffer(data bytes.Buffer) Binary {
	return Binary{
		Internal: data.Bytes(),
	}
}

func (binary Binary) Bytes() []byte {
	return binary.Internal
}

func (binary Binary) Buffer() bytes.Buffer {
	return *bytes.NewBuffer(binary.Internal)
}

func (binary *Binary) SetBytes(data []byte) {
	binary.Internal = data
}

func (binary *Binary) SetBuffer(data bytes.Buffer) {
	binary.Internal = data.Bytes()
}

func (binary Binary) MarshalJSON(ctx context.Context) ([]byte, error) {
	c := FromContext(ctx)
	if c == nil {
		return nil, errors.New("binary: no marshall context")
	}
	if c.Uid == "" {
		c.Uid = newSeed()
	}
	if c.Buffers == nil {
		c.Buffers = [][]byte{binary.Internal}
	} else {
		c.Buffers = append(c.Buffers, binary.Internal)
	}

	// Buffer index (i) is important because order for map items is undefined in JSON. So parsers
	// can choose to reshuffle map items.
	str := "{\"__b\":\"" + c.Uid + "\", \"i\":" + strconv.FormatInt(int64(len(c.Buffers))-1, 10) + "}"
	return []byte(str), nil
}

func newSeed() string {
	bytes := [6]byte{0, 0, 0, 0, 0, 0}
	slice := bytes[:]
	_, err := rand.Read(slice)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(slice)
}
