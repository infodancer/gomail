package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/queue"
)

var recipientLimit *int
var logger *log.Logger

func main() {
	helo := flag.String("helo", "h", "The helo string to use when greeting clients")

	logger = log.New(os.Stderr, "", 0)

	cfg := Config{
		ServerName: "",
		Banner:     *helo,
		Spamc:      "",
		maxsize:    0,
		MQueue:     &queue.Queue{},
	}

	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	var c connect.TCPConnection
	c, err := connect.NewStandardIOConnection(r, w)
	if err != nil {
		logger.Println("error creating new StandardIOConnection")
		os.Exit(1)
	}
	s, err := cfg.Start(&c)
	if err != nil {
		logger.Println("error sending greeting")
		os.Exit(2)
	}
	err = s.HandleConnection()
	if err != nil {
		logger.Println("error handling connection")
		os.Exit(3)
	}
}

// extractAddress parses an SMTP command line for an @ address within <>
func extractAddressPart(line string) (*string, error) {
	begin := strings.Index(line, "<") + 1
	end := strings.LastIndex(line, ">")
	if begin == -1 || end == -1 {
		return nil, errors.New("Address not found in command")
	}
	value := line[begin:end]
	// RFC 5321 https://tools.ietf.org/html/rfc5321#section-4.5.3
	if len(value) > 254 {
		return nil, errors.New("Address exceeds maximum length of email address")
	}
	return &value, nil
}

// extractUsername decodes the username from the client's response
func extractUsername(resp string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(resp)
	if err != nil {
		return "", errors.New("Base64 decode failed")
	}
	logger.Println("CRAM-MD5 response: ", decoded)
	return "", nil
}

// extractPassword finds the password for the given user
func extractPassword(username string) string {
	return ""
}

// createChallenge creates a challenge for use in CRAM-MD5
func createChallenge() string {
	challenge := ""
	encoded := base64.StdEncoding.EncodeToString([]byte(challenge))
	return encoded
}

// isSuspiciousInput looks for input that contains filename elements
// This method should be used to check addresses or domain names coming from external sources
// It's not perfect, but it works for now
func isSuspiciousAddress(input string) bool {
	// isSafe := regexp.MustCompile(`^[A-Za-z]+@[A-Za-z]+$`).MatchString
	i := strings.Index(input, "..")
	if i == -1 {
		i = strings.Index(input, "/")
		if i == -1 {
			i = strings.Index(input, "\\")
			if i == -1 {
				return false
			}
		}
	}
	return true
}
