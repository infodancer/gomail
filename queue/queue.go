package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var logger *log.Logger

// Queue defines an on-disk message queue
type Queue struct {
	Directory string
}

// Envelope stores envelope information for a message in the queue
type Envelope struct {
	MessagePath  string
	EnvelopePath string
	Sender       string
	From         string
	Recipients   []EnvelopeRecipient
}

// EnvelopeRecipient tracks recipients and delivery status
type EnvelopeRecipient struct {
	Recipient string
	Delivered bool
	Result    []EnvelopeDelivery
}

// EnvelopeDelivery tracks the result of the last delivery attempt
type EnvelopeDelivery struct {
	DeliveryResult string
}

func init() {
	logger = log.New(os.Stderr, "", 0)
}

// GetQueue provides a queue object based on the given directory
func GetQueue(directory string) (*Queue, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return CreateQueue(directory)
	}
	var result Queue
	result.Directory = directory
	return &result, nil
}

// CreateQueue creates a queue directory structure at the provided location
func CreateQueue(path string) (*Queue, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return nil, err
		}
		curDir := filepath.Join(path, "msg")
		if err := os.Mkdir(curDir, os.ModePerm); err != nil {
			return nil, err
		}
		tmpDir := filepath.Join(path, "tmp")
		if err := os.Mkdir(tmpDir, os.ModePerm); err != nil {
			return nil, err
		}
		newDir := filepath.Join(path, "env")
		if err := os.Mkdir(newDir, os.ModePerm); err != nil {
			return nil, err
		}
		return GetQueue(path)
	}
	return nil, errors.New("queue path does not exist")
}

// Enqueue places a message into the queue
func (q Queue) Enqueue(env Envelope, msg string) error {
	name := createUniqueName()
	envFile := filepath.Join(q.Directory, "env", name+".env")
	msgFile := filepath.Join(q.Directory, "msg", name+".msg")
	// Add the message path to the envelope
	env.MessagePath = msgFile
	env.EnvelopePath = envFile

	// Technically, to follow Maildir rules, we should write to tmp and then move
	// However, for now, we are just writing directly
	logger.Printf("Writing envelope to queue file: %v", envFile)
	envMarshalled, err := json.Marshal(env)
	if err != nil {
		return errors.New("could not marshall envelope to json")
	}
	logger.Printf("Writing envelope: %v", string(envMarshalled))

	logger.Printf("Writing message to queue file: %v", msgFile)
	err = ioutil.WriteFile(envFile, envMarshalled, 0644)
	if err != nil {
		return errors.New("could not write envelope to file")
	}

	err = ioutil.WriteFile(msgFile, []byte(msg), 0644)
	if err != nil {
		return errors.New("could not write message to file")
	}

	return nil
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
