package maildir

import (
	"errors"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaildir(t *testing.T) {
	body := "Test message."
	tmpdir, err := os.MkdirTemp("", "maildir-test-")
	assert.NoError(t, err, "error creating tmpdir")
	defer assert.NoError(t, os.RemoveAll(tmpdir))
	md, err := Create(path.Join(tmpdir, "Maildir"))
	assert.NoError(t, err, "error creating maildir: %w", err)

	msgs, err := md.List()
	assert.NoError(t, err, "error listing empty maildir: %w", err)
	assert.Equal(t, 0, len(msgs), "expected new maildir to be empty")

	msgid, err := md.Add([]byte(body))
	if err != nil {
		t.Fatal("error writing message")
	}
	rmsg, err := md.Read(msgid)
	if err != nil {
		t.Fatal("error reading message")
	}
	if len(rmsg) <= 0 {
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
	err = md.Delete()
	assert.NoError(t, err)
}

func TestMaildirDelete(t *testing.T) {
	body := "Test message."
	tmpdir, err := os.MkdirTemp("", "maildir-test-")
	assert.NoError(t, err, "error creating tmpdir")
	defer assert.NoError(t, os.RemoveAll(tmpdir))
	md, err := Create(path.Join(tmpdir, "Maildir"))
	assert.NoError(t, err, "error creating maildir: %w", err)
	msgid, err := md.Add([]byte(body))
	assert.NoError(t, err, "error adding message: %w", err)

	err = md.Delete(msgid)
	assert.NoError(t, err, "error deleting message %s: %w", msgid, err)
	_, err = md.Read(msgid)
	if assert.Error(t, err, "no error reading deleted message") {
		assert.ErrorContains(t, err, "", "expected failure to read deleted message")
	}
	err = md.Remove()
	assert.NoError(t, err, "error removing maildir: %w", err)
}

func TestMaildirRemove(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "maildir-test-")
	assert.NoError(t, err, "error creating tmpdir")
	defer assert.NoError(t, os.RemoveAll(tmpdir))

	md, err := Create(path.Join(tmpdir, "Maildir"))
	assert.NoError(t, err, "error creating maildir: %w", err)
	err = md.Delete()
	assert.NoError(t, err, "error deleting maildir: %w", err)
	_, err = os.Stat(md.directory)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Errorf("maildir delete failed to remove maildir")
	}
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
