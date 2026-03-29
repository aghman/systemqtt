// Package agent provides JSON envelopes for --agent CLI output.
package agent

import (
	"encoding/json"
	"fmt"
	"io"
)

// Envelope is the stable JSON shape for non-serve commands in agent mode.
type Envelope struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
	Error  *string         `json:"error"`
}

// WriteOK writes a successful envelope with JSON-marshalable data.
func WriteOK(w io.Writer, data any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(Envelope{
		Status: "ok",
		Data:   raw,
		Error:  nil,
	})
}

// WriteError writes a failed envelope.
func WriteError(w io.Writer, msg string) error {
	return json.NewEncoder(w).Encode(Envelope{
		Status: "error",
		Data:   nil,
		Error:  ptr(msg),
	})
}

func ptr(s string) *string { return &s }

// PrintlnHuman prints one line for non-agent mode errors.
func PrintlnHuman(w io.Writer, agentMode bool, err error) {
	if agentMode {
		_ = WriteError(w, err.Error())
		return
	}
	fmt.Fprintln(w, err)
}
