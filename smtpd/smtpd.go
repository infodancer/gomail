package smtpd

import (
	"bufio"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"gomail/address"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Session describes the current session
type Session struct {
	// Started indicates the time the session began
	Started time.Time
	// Activity indicates the time of last activity
	Activity time.Time

	// Sender is the authenticated user sending the message; nil if not authenticated
	Sender *address.Address
	// From is the claimed sender of the message
	From *address.Address
	// Recipients is the array of recipients
	Recipients []address.Address
	// Data is the array of lines in the message itself
	Data []string
}

var recipientLimit *int
var logger *log.Logger
var reader *bufio.Reader
var helo *string
var spamc *string

func main() {
	helo = flag.String("helo", "h", "The helo string to use when greeting clients")
	recipientLimit = flag.Int("maxrcpt", 100, "The maximum number of recipients on a single message")
	spamc = flag.String("spamc", "/usr/bin/spamc", "The path to spamassassin's spamc client")

	logger = log.New(os.Stderr, "", 0)
	logger.Println("gomail smtpd started")
	session := Session{}
	handleConnection(&session)
}

// sendLine accepts a line without linefeeds and sends it with network linefeeds
func sendCodeLine(code int, line string) {
	fmt.Print(code, " ", line, "\r\n")
}

// sendLine accepts a line without linefeeds and sends it with network linefeeds
func sendLine(line string) {
	fmt.Print(line, "\r\n")
}

func handleConnection(session *Session) {
	session.Started = time.Now()
	sendCodeLine(220, *helo+" ESMTP")

	reader = bufio.NewReader(os.Stdin)
	for finished := false; !finished; {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		logger.Println(">" + line)
		code, resp, finished := handleInputLine(session, line)
		sendCodeLine(code, resp)
		if finished {
			break
		}
	}
}

func handleInputLine(session *Session, line string) (int, string, bool) {
	session.Activity = time.Now()
	cmd := strings.Split(line, " ")
	command := strings.ToUpper(strings.TrimSpace(cmd[0]))
	switch command {
	case "HELO":
		return processHELO(session, line)
	case "EHLO":
		{
			// This is a bit of a special case because of extensions
			sendLine("250-8BITMIME")
			sendLine("250-PIPELINING")
			sendLine("250-AUTH CRAM-MD5")
			return processEHLO(session, line)
		}
	case "AUTH":
		return processAUTH(session, line)
	case "RCPT":
		return processRCPT(session, line)
	case "MAIL":
		return processMAIL(session, line)
	case "DATA":
		return processDATA(session, line)

	// These commands are not vital
	case "RSET":
		return processRSET(session, line)
	case "NOOP":
		return processNOOP(session, line)
	case "VRFY":
		return processVRFY(session, line)

	// QUIT terminates the session
	case "QUIT":
		return processQUIT(session, line)
	default:
		return 500, "Unrecognized command", false
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

// processAUTH handles the auth process
func processAUTH(session *Session, line string) (int, string, bool) {
	// For now, we haven't implemented this
	if strings.HasPrefix(line, "AUTH ") {
		authType := line[5:len(line)]
		// Reject insecure authentication methods
		if authType != "CRAM-MD5" {
			return 500, "Unrecognized command", false
		}
		challenge := createChallenge()
		sendCodeLine(354, challenge)
		resp, err := reader.ReadString('\n')
		if err != nil {
			return 550, "Authentication failed", false
		}
		username, err := extractUsername(resp)
		if err != nil {
			return 550, "Authentication failed", false
		}
		password := extractPassword(username)
		if !validateCramMD5(resp, password) {
			return 550, "Authentication failed", false
		}
		return 235, "Authentication successful", false
	}
	return 500, "Unrecognized command", false
}

// createChallenge creates a challenge for use in CRAM-MD5
func validateCramMD5(resp string, password string) bool {
	return false
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

// processHELO handles the standard SMTP helo
func processHELO(session *Session, line string) (int, string, bool) {
	return 250, "Hello", false
}

// processEHLO handles the extended EHLO command, but the extensions are listed elsewhere
func processEHLO(session *Session, line string) (int, string, bool) {
	return 250, "Hello", false
}

func processRCPT(session *Session, line string) (int, string, bool) {
	addr, err := extractAddressPart(line)
	if err != nil {
		return 550, "Invalid address", false
	}
	// Check if the sender has been set
	if session.From == nil {
		return 503, "need MAIL before RCPT", false
	}
	// Check for number of recipients
	if len(session.Recipients) >= *recipientLimit {
		return 452, "Too many recipients", false
	}
	// Check if this is being sent to a bounce address
	if len(*addr) == 0 {
		return 503, "We don't accept mail to that address", false
	}
	// Check for relay and allow only if sender has authenticated
	if !isLocalAddress(*addr) && session.Sender == nil {
		return 553, "We don't relay mail to remote addresses", false
	}

	recipient, err := address.CreateAddress(*addr)
	if err != nil {
		return 550, "Invalid address", false
	}

	session.Recipients = append(session.Recipients, *recipient)
	return 250, "OK", false
}

func processMAIL(session *Session, line string) (int, string, bool) {
	if session.From != nil {
		return 400, "MAIL FROM already sent", false
	}
	addr, err := extractAddressPart(line)
	if err != nil {
		return 451, "Invalid address", false
	}
	// Check if this is a bounce message
	if len(*addr) == 0 {
		return 551, "We don't accept mail to that address", false
	}

	session.From, err = address.CreateAddress(*addr)
	if err != nil {
		return 451, "Invalid address", false
	}
	return 250, "OK", false
}

func createReceived(session *Session) (string, error) {
	rcv := "Received: from "
	// remote server info
	rcv += " by "
	rcv += *helo
	rcv += " with SMTP; "
	rcv += time.Now().String()
	return rcv, nil
}

func processDATA(session *Session, line string) (int, string, bool) {
	// Did the user specify an envelope?
	// Check if the sender has been set
	if session.From == nil {
		return 503, "need MAIL before DATA", false
	}
	// Check for number of recipients
	if len(session.Recipients) == 0 {
		return 503, "need RCPT before DATA", false
	}
	// Generate a received header
	rcv, err := createReceived(session)
	if err != nil {
		return 451, "message could not be accepted at this time, try again later", false
	}
	session.Data = append(session.Data, rcv)

	// Accept the start of message data
	sendCodeLine(354, "Send message content; end with <CRLF>.<CRLF>")
	for finished := false; !finished; {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		logger.Println(">" + line)
		session.Data = append(session.Data, line)
		if strings.HasPrefix(line, ".") && !strings.HasPrefix(line, "..") {
			// Check with spamc if needed
			if spamc != nil {
				msg, err := checkSpam(session)
				if err != nil {
					// Temporary failure if we can't check it
					return 451, "message could not be accepted at this time, try again later", false
				}
				// We don't block here; let the user use their filters
				session.Data = msg
			}
			err := enqueue(session)
			if err != nil {
				logger.Println("Unable to enqueue message!")
				return 451, "message could not be accepted at this time, try again later", false
			}
			return 250, "message accepted for delivery", false
		}
	}
	// If we somehow get here without the message being completed, return a temporary failure
	return 451, "message could not be accepted at this time, try again later", false
}

func checkSpam(session *Session) ([]string, error) {
	result := make([]string, len(session.Data))
	cmd := exec.Command(*spamc)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err2 := cmd.StdoutPipe()
	if err2 != nil {
		log.Fatal(err2)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	defer stdin.Close()
	defer stdout.Close()
	for _, line := range session.Data {
		// stdin.Write(line)
	}

	spamreader := bufio.NewReader(stdout)
	for finished := false; !finished; {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		logger.Println(">" + line)
		code, resp, finished := handleInputLine(session, line)
		sendCodeLine(code, resp)
		if finished {
			break
		}
	}
	for {
		stdout.Read()
	}
}

// enqueue places the current message (as contained in the session) into the disk queue; ie accepting delivery
func enqueue(session *Session) error {
	return errors.New("Queuing code not yet implemented")
}

// processQUIT simply terminates the session
func processQUIT(session *Session, line string) (int, string, bool) {
	return 221, "goodbye", true
}

// processRSET clears the session information
func processRSET(session *Session, line string) (int, string, bool) {
	session.Sender = nil
	session.From = nil
	session.Recipients = make([]address.Address, 0)
	session.Data = make([]string, 0)
	return 250, "OK", false
}

func processNOOP(session *Session, line string) (int, string, bool) {
	return 250, "OK", false
}

func processVRFY(session *Session, line string) (int, string, bool) {
	return 500, "VRFY not supported", false
}

// isLocalAddress checks whether the address is local; this should probably be moved to some sort of domain package later
func isLocalAddress(input string) bool {
	return false
}

// extractDomainPath transforms a domain into a filesystem path
func extractDomainPath(input string) (string, error) {
	var result string
	parts := strings.Split(input, ".")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return "", errors.New("extraneous dot detected")
		}

	}

	return result, nil
}
