package maildir

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Maildir represents a directory structure on disk containing mail
type Maildir struct {
	mutex     sync.RWMutex
	directory string
	messages  map[string][]rune
}

// Create creates a maildir directory structure
func Create(path string) (*Maildir, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, os.FileMode(0600)); err != nil {
			return nil, err
		}
		curDir := filepath.Join(path, "cur")
		if err := os.Mkdir(curDir, os.FileMode(0600)); err != nil {
			return nil, err
		}
		tmpDir := filepath.Join(path, "tmp")
		if err := os.Mkdir(tmpDir, os.FileMode(0600)); err != nil {
			return nil, err
		}
		newDir := filepath.Join(path, "new")
		if err := os.Mkdir(newDir, os.FileMode(0600)); err != nil {
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
		messages:  make(map[string][]rune, 0),
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
	filename = msgid + ":2," + string(flags)
	return filename
}

// Write replaces an existing message with new content
func (m *Maildir) Write(msgid string, msg []byte) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	name := createUniqueName()
	tmpfilename := path.Join(m.directory, "tmp", name)
	err := ioutil.WriteFile(tmpfilename, msg, os.FileMode(0600))
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
	m.mutex.Lock()
	defer m.mutex.Unlock()
	name := createUniqueName()
	tmpfilename := path.Join(m.directory, "tmp", name)
	err := ioutil.WriteFile(tmpfilename, msg, os.FileMode(0600))
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

func (m *Maildir) Read(msgid string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.Unlock()
	msg, err := ioutil.ReadFile(path.Join(m.directory, "new", msgid))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// If we found it, we have to move it
	msg, err = ioutil.ReadFile(path.Join(m.directory, "cur", msgid))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if msg != nil {
		return msg, nil
	}
	return msg, nil
}

func (m *Maildir) List() ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.Unlock()
	msgs := make([]string, 0)
	for k := range m.messages {
		msgs = append(msgs, k)
	}
	return msgs, nil
}

// Scan checks a maildir for new messages, moving them to cur
func (m *Maildir) Scan() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	msgs, err := ioutil.ReadDir(path.Join(m.directory, "new"))
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

		// Add to the index with flags set
		m.messages[name] = make([]rune, 0)
	}
	return nil
}

func (m *Maildir) Flags(msgid string) ([]rune, error) {
	m.mutex.RLock()
	defer m.mutex.Unlock()
	runes, ok := m.messages[msgid]
	if !ok {
		return nil, errors.New("message not found")
	}
	return runes, nil
}

func (m *Maildir) SetFlag(msgid string, flag rune, on bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
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
	m.mutex.Lock()
	defer m.mutex.Unlock()

	oldflags, ok := m.messages[msgid]
	if !ok {
		return errors.New("message not found")
	}

	oldfilename := path.Join(m.directory, "cur", createFilename(msgid, oldflags))
	newfilename := path.Join(m.directory, "cur", createFilename(msgid, flags))

	err := os.Rename(oldfilename, newfilename)
	if err != nil {
		return err
	}

	// Update the index
	m.messages[msgid] = flags
	return nil
}
