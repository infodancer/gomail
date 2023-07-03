package maildir

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaildir(t *testing.T) {
	body := "Test message."
	tmpdir := os.TempDir()
	tmpname := createUniqueName()
	md, err := Create(path.Join(tmpdir, tmpname))
	assert.NoError(t, err, "error creating maildir: %w", err)

	msgs, err := md.List()
	if err != nil {
		t.Fatal("error listing messages")
	}
	if len(msgs) > 0 {
		t.Fatal("messages should be empty")
	}
	msgid, err := md.Add([]byte(body))
	if err != nil {
		t.Fatal("error writing message")
	}
	rmsg, err := md.Read(msgid)
	if err != nil {
		t.Fatal("error reading message")
	}
	if rmsg == nil || len(rmsg) <= 0 {
		t.Fatal("error reading message")
	}
	msgs, err = md.List()
	if err != nil {
		t.Fatal("error reading message list")
	}
	for _, msgid := range msgs {
		msg, err := md.Read(msgid)
		if err != nil {
			t.Fatalf("error reading msg %s", msgid)
		}
		strings.Contains(string(msg), body)
	}
	os.RemoveAll(path.Join(tmpdir, tmpname))
}

func TestCreateUniqueName(t *testing.T) {
	names := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		name := createUniqueName()
		_, ok := names[name]
		if ok {
			t.Logf("createUniqueName created a duplicate name %v", name)
			t.Fail()
		}
	}
}
