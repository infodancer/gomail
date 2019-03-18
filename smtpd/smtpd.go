package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"gomail/address"
	"gomail/domain"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Session describes the current session
type Session struct {
	// Started indicates the time the session began
	Started time.Time
	// Activity indicates the time of last activity
	Activity time.Time
	// Protocol contains the protocol, usually TCP
	Protocol string
	// LocalIP contains the local (listening) ip address
	LocalIP string
	// LocalPort contains the local (listening) port number
	LocalPort int
	// LocalHost contains the hostname of the server
	LocalHost string
	// RemoteIP contains the remote ip address
	RemoteIP string
	// RemotePort contains the remote port number
	RemotePort int
	// RemoteHost contains the hostname of the remote client
	RemoteHost string
	// Sender is the authenticated user sending the message; nil if not authenticated
	Sender *address.Address
	// From is the claimed sender of the message
	From *address.Address
	// Recipients is the array of recipients
	Recipients []address.Address
	// Data is the message body itself
	Data string
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
	// Default helo to localhost if not manually set
	if helo == nil {
		helo = &session.LocalHost
		if helo == nil || len(*helo) == 0 {
			helo = &session.LocalIP
		}
	}
	handleConnection(&session)
}

func initializeSessionFromEnvironment(session *Session) {
	session.Protocol = os.Getenv("PROTO")
	session.LocalIP = os.Getenv("TCPLOCALIP")
	lport, err := strconv.Atoi(os.Getenv("TCPLOCALPORT"))
	if err != nil {
		session.LocalPort = 0
	}
	session.LocalPort = lport
	if err != nil {
	}
	session.LocalHost = os.Getenv("TCPLOCALHOST")
	session.RemoteIP = os.Getenv("TCPREMOTEIP")
	rport, err := strconv.Atoi(os.Getenv("TCPREMOTEPORT"))
	if err != nil {
		session.RemotePort = 0
	}
	session.RemotePort = rport
	session.RemoteHost = os.Getenv("TCPREMOTEHOST")
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

	recipient, err := address.CreateAddress(*addr)
	if err != nil {
		return 550, "Invalid address", false
	}

	// Check for relay and allow only if sender has authenticated
	if !isLocalAddress(recipient) && session.Sender == nil {
		return 553, "We don't relay mail to remote addresses", false
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
	rcv += os.Getenv("TCPREMOTEIP")
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
	session.Data = rcv

	// Accept the start of message data
	sendCodeLine(354, "Send message content; end with <CRLF>.<CRLF>")
	for finished := false; !finished; {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		logger.Println(">" + line)
		if strings.HasPrefix(line, ".") {
			if strings.HasPrefix(line, "..") {
				// Remove escaped period character
				line = line[1:len(line)]
			} else {
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
			session.Data += line
		}
	}
	// If we somehow get here without the message being completed, return a temporary failure
	return 451, "message could not be accepted at this time, try again later", false
}

// CheckSpam runs spamc to see if a message is spam, and returns either an error, or the modified message
func checkSpam(session *Session) (string, error) {
	result := ""
	cmd := exec.Command(*spamc)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	defer stdin.Close()
	defer stdout.Close()

	spamwriter := bufio.NewWriter(stdin)
	spamwriter = bufio.NewWriterSize(spamwriter, len(session.Data))
	spamwriter.WriteString(session.Data)

	// Create a reader at least as big as the original message with extra space for headers
	spamreader := bufio.NewReader(stdout)
	spamreader = bufio.NewReaderSize(spamreader, len(session.Data)+1024)

	// Read the message back out
	for finished := false; !finished; {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		result += line
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return result, nil
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
	session.Data = ""
	return 250, "OK", false
}

func processNOOP(session *Session, line string) (int, string, bool) {
	return 250, "OK", false
}

func processVRFY(session *Session, line string) (int, string, bool) {
	return 500, "VRFY not supported", false
}

// isLocalAddress checks whether the address is local
func isLocalAddress(addr *address.Address) bool {
	if dom, err := domain.GetDomain(*addr.Domain); err == nil && dom != nil {
		if user, err := dom.GetUser(*addr.User); err == nil && user != nil {
			return true
		}
	}
	return false
}
