package privileged_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/livingsilver94/backee/privileged"
)

func TestDestMessageCodec(t *testing.T) {
	expect := privileged.SourceDestMessage{
		Source: "a/b",
		Dest:   "c/d",
	}
	buf := &bytes.Buffer{}
	err := expect.EncodeWriter(buf)
	if err != nil {
		t.Fatal(err)
	}

	var msg privileged.SourceDestMessage
	err = msg.DecodeReader(buf)
	if err != nil {
		t.Fatal(err)
	}
	if msg != expect {
		t.Fatalf("expected message %v. Got %v", expect, msg)
	}
}

func TestWriteContentMessageCodec(t *testing.T) {
	const content = "test test"
	expect := privileged.WriteContentMessage{
		Dest:    "c/d",
		Content: strings.NewReader(content),
	}
	buf := &bytes.Buffer{}
	err := expect.EncodeWriter(buf)
	if err != nil {
		t.Fatal(err)
	}

	var msg privileged.WriteContentMessage
	err = msg.DecodeReader(buf)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Dest != expect.Dest {
		t.Fatalf("expected Dest field %v. Got %v", expect.Dest, msg.Dest)
	}
	s, err := io.ReadAll(msg.Content)
	if err != nil {
		t.Fatal(err)
	}
	if string(s) != content {
		t.Fatalf("expected content %v. Got %v", content, s)
	}
}
