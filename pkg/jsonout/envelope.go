package jsonout

import (
	"encoding/json"
	"fmt"
	"time"
)

type Envelope struct {
	Success   bool   `json:"success"`
	Command   string `json:"command"`
	Timestamp string `json:"timestamp"`
	Data      any    `json:"data"`
	Error     any    `json:"error,omitempty"`
}

func OK(command string, data any) []byte {
	env := Envelope{
		Success:   true,
		Command:   command,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
		Error:     nil,
	}
	b, _ := json.MarshalIndent(env, "", "  ")
	return b
}

func Fail(command string, err error) []byte {
	env := Envelope{
		Success:   false,
		Command:   command,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      nil,
		Error:     err.Error(),
	}
	b, _ := json.MarshalIndent(env, "", "  ")
	return b
}

var _ = fmt.Sprintf
