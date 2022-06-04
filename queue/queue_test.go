package queue

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateQueue(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "queue")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	q, err := CreateQueue(tempDir)
	if err != nil {
		t.Fatalf("failed to create queue directory: %v", err)
	}

	if fileExists(tempDir) {
		msgf := filepath.Join(tempDir, "msg")
		if !fileExists(msgf) {
			t.Fatalf("queue msg directory not created!")
		}
		envf := filepath.Join(tempDir, "envf")
		if !fileExists(envf) {
			t.Fatalf("queue env directory not created!")
		}
		tmpf := filepath.Join(tempDir, "tmpf")
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
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
