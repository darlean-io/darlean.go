package jsonbinary

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// TODO: Implement buffers
func Serialize(data any) ([]byte, error) {
	binary, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBufferString("JB00\n")
	buf.Write(binary)
	return buf.Bytes(), nil
}

func Deserialize(data []byte, value any) error {
	buf := bytes.NewBuffer(data)
	header, err := buf.ReadString('\n')
	if err != nil {
		return err
	}
	if header != "JB00\n" {
		return fmt.Errorf("jsonbinary: invalid header: %v", header)
	}

	return json.Unmarshal(buf.Bytes(), value)
}
