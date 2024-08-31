package queue

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateQueue(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "queue")
	assert.NoError(t, err, "failed to create temporary directory: %w", err)
	q, err := CreateQueue(tempDir)
	assert.NoError(t, err, "failed to create queue directory: %w", err)

	if fileExists(tempDir) {
		msgf := filepath.Join(tempDir, "msg")
		if !fileExists(msgf) {
			t.Fatalf("queue msg directory not created!")
		}
		envf := filepath.Join(tempDir, "env")
		if !fileExists(envf) {
			t.Fatalf("queue env directory not created!")
		}
		tmpf := filepath.Join(tempDir, "tmp")
		if !fileExists(tmpf) {
			t.Fatalf("queue tmp directory not created!")
		}

		sender := "sender@example.com"
		recipients := make([]string, 0)
		recipients = append(recipients, "recipient@example.com")
		msg := []byte("This is a test message.")
		err = q.Enqueue(sender, recipients, msg)
		if err != nil {
			t.Fatalf("enqueue failed: %v", err)
		}
	} else {
		t.Fatalf("queue directory not created!")
	}
}

func TestGetQueue(t *testing.T) {
}

func TestCreateUniqueName(t *testing.T) {
}

func TestEnqueue(t *testing.T) {
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		return false
	}
	return true
}
