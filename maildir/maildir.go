package maildir

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

// Maildir represents a directory structure on disk containing mail
type Maildir struct {
	directory string
}

// Create creates a maildir directory structure
func Create(path string) (*Maildir, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.FileMode(0755)); err != nil {
			return nil, err
		}
		curDir := filepath.Join(path, "cur")
		if err := os.MkdirAll(curDir, os.FileMode(0755)); err != nil {
			return nil, err
		}
		tmpDir := filepath.Join(path, "tmp")
		if err := os.MkdirAll(tmpDir, os.FileMode(0755)); err != nil {
			return nil, err
		}
		newDir := filepath.Join(path, "new")
		if err := os.MkdirAll(newDir, os.FileMode(0755)); err != nil {
			return nil, err
		}
		return New(path)
	}
	return nil, errors.New("maildir does not exist")
}

// New creates a maildir representation of an existing Maildir format directory structure
func New(dir string) (*Maildir, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(path.Join(dir, "new")); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(path.Join(dir, "tmp")); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(path.Join(dir, "cur")); os.IsNotExist(err) {
		return nil, err
	}

	m := Maildir{
		directory: dir,
	}
	err := m.Scan()
	if err != nil {
		log.Printf("error scanning maildir: %v", err)
		return nil, errors.New("error scanning maildir")
	}
	return &m, nil
}

var deliveryCount int64

func createUniqueName() string {
	date := time.Now()
	left := date.Second()
	random := rand.Int63()
	right, err := os.Hostname()
	if err != nil {
		right = `localhost`
	}
	atomic.AddInt64(&deliveryCount, 1)
	pid := syscall.Getpid()
	center := fmt.Sprintf("P%vM%vR%vQ%d", pid, date.Nanosecond(), random, deliveryCount)
	result := fmt.Sprintf("%v.%v.%v", left, center, right)
	return result
}

func createFilename(msgid string, flags []rune) string {
	var filename string
	i := strings.LastIndex(msgid, ":")
	if i > 0 {
		filename = msgid[0:i]
	} else {
		filename = msgid
	}
	filename += ":2," + string(flags)
	return filename
}

// Delete removes the maildir itself
func (m *Maildir) Delete() error {
	return os.RemoveAll(m.directory)
}

// Write replaces an existing message with new content
func (m *Maildir) Write(msgid string, msg []byte) (string, error) {
	name := createUniqueName()
	tmpfilename := path.Join(m.directory, "tmp", name)
	err := os.WriteFile(tmpfilename, msg, os.FileMode(0600))
	if err != nil {
		return ``, err
	}
	newfilename := path.Join(m.directory, "new", name)
	err = os.Rename(tmpfilename, newfilename)
	if err != nil {
		return ``, err
	}
	return name, nil
}

// Add adds a new message to a maildir
func (m *Maildir) Add(msg []byte) (string, error) {
	name := createUniqueName()
	tmpfilename := path.Join(m.directory, "tmp", name)
	err := os.WriteFile(tmpfilename, msg, os.FileMode(0600))
	if err != nil {
		return ``, err
	}
	newfilename := path.Join(m.directory, "new", name)
	err = os.Rename(tmpfilename, newfilename)
	if err != nil {
		return ``, err
	}
	return name, nil
}

// findByMsgID returns the filename holding msgid
func findByMsgID(directory, msgid string) (string, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		name := file.Name()
		if strings.HasPrefix(name, msgid) {
			path := filepath.Join(directory, name)
			if file.IsDir() {
				return path, errors.New("msgid " + msgid + " is a directory")
			}
			return path, nil
		}
	}
	return "", os.ErrNotExist
}

// Strips the message flags off for just the msgid
func getMsgIDFromFilename(f string) string {
	i := strings.Index(f, ":")
	if i == -1 {
		return f
	}
	return f[0:i]
}

func (m *Maildir) Read(msgid string) ([]byte, error) {
	msgpath, err := findByMsgID(path.Join(m.directory, "new"), msgid)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			msgpath, err = findByMsgID(path.Join(m.directory, "cur"), msgid)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	msg, err := os.ReadFile(msgpath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return msg, nil
}

// List returns an array of valid message identifiers
func (m *Maildir) List() ([]string, error) {
	// Check for new messages so we only have to read one dir
	m.Scan()

	msgs := make([]string, 0)
	files, err := os.ReadDir(path.Join(m.directory, "cur"))
	if err != nil {
		return msgs, err
	}
	for _, file := range files {
		msgid := getMsgIDFromFilename(file.Name())
		if !file.IsDir() {
			msgs = append(msgs, msgid)
		}
	}
	return msgs, nil
}

// Scan checks a maildir for new messages, moving them to cur
func (m *Maildir) Scan() error {
	msgs, err := os.ReadDir(path.Join(m.directory, "new"))
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		name := msg.Name()
		newfilename := path.Join(m.directory, "new", name)
		curfilename := path.Join(m.directory, "cur", name+":2,")
		err = os.Rename(newfilename, curfilename)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Maildir) Flags(msgid string) ([]rune, error) {
	curpath := path.Join(m.directory, "cur")
	name, err := findByMsgID(curpath, msgid)
	if err != nil {
		return nil, err
	}
	result := make([]rune, 0)
	i := strings.Index(name, ":2,")
	if i != -1 {
		for _, r := range name {
			result = append(result, r)
		}

		sort.Slice(result, func(i, j int) bool {
			return i < j
		})
	}
	return result, nil
}

func (m *Maildir) SetFlag(msgid string, flag rune, on bool) error {
	flags, err := m.Flags(msgid)
	if err != nil {
		return err
	}
	nflags := make([]rune, len(flags))
	for _, r := range flags {
		if r == flag {
			if on {
				nflags = append(nflags, r)
			}
		} else {
			nflags = append(nflags, r)
		}
	}
	err = m.SetFlags(msgid, nflags)
	if err != nil {
		return err
	}
	return nil
}

func (m *Maildir) SetFlags(msgid string, flags []rune) error {
	oldfilename, err := findByMsgID(path.Join(m.directory, "cur"), msgid)
	if err != nil {
		return err
	}
	newfilename := path.Join(m.directory, "cur", createFilename(msgid, flags))
	err = os.Rename(oldfilename, newfilename)
	if err != nil {
		return err
	}

	return nil
}
