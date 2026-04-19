package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestWriteJSON_Success(t *testing.T) {
	var out bytes.Buffer

	err := WriteJSON(&out, map[string]string{"category": "bugfix"})
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "{\n  \"category\": \"bugfix\"\n}\n") {
		t.Fatalf("WriteJSON output = %q, want indented JSON", got)
	}
}

func TestWriteJSON_EncodeError(t *testing.T) {
	err := WriteJSON(failingWriter{}, map[string]string{"category": "bugfix"})
	if err == nil {
		t.Fatal("WriteJSON error = nil, want encode error")
	}
	if !strings.Contains(err.Error(), "encode json") {
		t.Fatalf("WriteJSON error = %q, want wrapped encode json error", err)
	}
}

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}
