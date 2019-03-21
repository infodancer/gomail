package maildir

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var deliveryCounter int64

// Maildir represents a directory structure on disk containing mail
type Maildir struct {
	Directory *string
}

// CreateMaildir creates a maildir directory structure
func CreateMaildir(path string) (*Maildir, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return nil, err
		}
		curDir := filepath.Join(path, "cur")
		if err := os.Mkdir(curDir, os.ModePerm); err != nil {
			return nil, err
		}
		tmpDir := filepath.Join(path, "tmp")
		if err := os.Mkdir(tmpDir, os.ModePerm); err != nil {
			return nil, err
		}
		newDir := filepath.Join(path, "new")
		if err := os.Mkdir(newDir, os.ModePerm); err != nil {
			return nil, err
		}
		return GetMaildir(path)
	}
	return nil, errors.New("maildir does not exist")
}

// GetMaildir retrieves a maildir representation of an existing Maildir format directory structure
func GetMaildir(path string) (*Maildir, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	result := Maildir{}
	result.Directory = &path
	return &result, nil
}

func createUniqueName() string {
	date := time.Now()
	left := date.Nanosecond()
	center := rand.Int63()
	right, err := os.Hostname()
	if err != nil {
	}

	result := fmt.Sprintf("%v.%v.%v", left, center, right)
	return result
}
