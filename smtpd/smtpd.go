package smtpd

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"infodancer.org/gomail/address"
	"infodancer.org/gomail/domain"
	"infodancer.org/gomail/queue"
)

// Session describes the current session
type Session struct {
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

// ProtocolHandler handles server-side SMTP
type ProtocolHandler struct {
	maxsize  *int64
	spamc    *string
	Session  *Session
	msgqueue *queue.Queue
}

// CreateSmtpdProtocolHandler creates a protocol handler for smtpd
func CreateSmtpdProtocolHandler(properties map[string]string) *ProtocolHandler {
	var err error
	queuedir := properties["queuedir"]
	handler := ProtocolHandler{}
	handler.msgqueue, err = queue.GetQueue(queuedir)
	if err != nil {
		logger.Printf("error opening mail queue from %v: %v", queuedir, err)
	}

	return &handler
}

// SendCodeLine accepts a line without linefeeds and sends it with network linefeeds and the provided response code
func (handler *ProtocolHandler) SendCodeLine(code int, line string) {
	fmt.Print(code, " ", line, "\r\n")
}

// SendLine accepts a line without linefeeds and sends it with network linefeeds
func (handler *ProtocolHandler) SendLine(line string) {
	fmt.Print(line, "\r\n")
}

// HandleInputLine accepts a line and handles it
func (handler *ProtocolHandler) HandleInputLine(session *Session, line string) (int, string, bool) {
	cmd := strings.Split(line, " ")
	command := strings.ToUpper(strings.TrimSpace(cmd[0]))
	switch command {
	case "HELO":
		return handler.processHELO(session, line)
	case "EHLO":
		{
			// This is a bit of a special case because of extensions
			handler.SendLine("250-8BITMIME")
			handler.SendLine("250-PIPELINING")
			handler.SendLine("250-AUTH CRAM-MD5")
			if handler.maxsize != nil && *handler.maxsize != 0 {
				size := strconv.FormatInt(*handler.maxsize, 10)
				handler.SendLine("250-SIZE " + size)
			}
			return handler.processEHLO(session, line)
		}
	case "AUTH":
		return handler.processAUTH(session, line)
	case "RCPT":
		return handler.processRCPT(session, line)
	case "MAIL":
		return handler.processMAIL(session, line)
	case "DATA":
		return handler.processDATA(session, line)

	// These commands are not vital
	case "RSET":
		return handler.processRSET(session, line)
	case "NOOP":
		return handler.processNOOP(session, line)
	case "VRFY":
		return handler.processVRFY(session, line)

	// QUIT terminates the session
	case "QUIT":
		return handler.processQUIT(session, line)
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
func (handler *ProtocolHandler) processAUTH(session *Session, line string) (int, string, bool) {
	// For now, we haven't implemented this
	if strings.HasPrefix(line, "AUTH ") {
		authType := line[5:len(line)]
		// Reject insecure authentication methods
		if authType != "CRAM-MD5" {
			return 500, "Unrecognized command", false
		}
		challenge := createChallenge()
		handler.SendCodeLine(354, challenge)
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
func (handler *ProtocolHandler) processHELO(session *Session, line string) (int, string, bool) {
	return 250, "Hello", false
}

// processEHLO handles the extended EHLO command, but the extensions are listed elsewhere
func (handler *ProtocolHandler) processEHLO(session *Session, line string) (int, string, bool) {
	return 250, "Hello", false
}

func (handler *ProtocolHandler) processRCPT(session *Session, line string) (int, string, bool) {
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
		logger.Println("Rejecting RCPT TO to bounce address: " + *addr)
		return 503, "We don't accept mail to that address", false
	}

	// Before we actually do filesystem operations, sanitize the input
	if isSuspiciousAddress(*addr) {
		logger.Println("Rejecting suspicious RCPT TO: " + *addr)
		return 550, "Invalid address", false
	}

	recipient, err := address.CreateAddress(*addr)
	if err != nil {
		logger.Println("CreateAddress failed: " + *addr)
		return 550, "Invalid address", false
	}

	// Check for relay and allow only if sender has authenticated
	dom, err := domain.GetDomain(*recipient.Domain)
	if session.Sender == nil {
		// Only bother to check domain if the sender is nil
		if dom == nil {
			return 551, "We don't relay mail to remote addresses", false
		}
	}

	// Check for local recipient existing if the domain is local
	if dom != nil {
		// We know the domain exists locally now
		user, err := dom.GetUser(*recipient.User)
		// Temporary error if we couldn't access the user for some reason
		if err != nil {
			logger.Println("Error from GetUser: ", err)
			return 451, "Address does not exist or cannot receive mail at this time, try again later", false
		}
		// If we got back nil without error, they really don't exist
		if user == nil {
			return 550, "User does not exist", false
		}
		// But if they do exist, check that their mailbox also exists
		maildir, err := dom.GetUserMaildir(*recipient.User)
		if err != nil {
			logger.Println("User exists but GetUserMaildir errors: ", err)
			return 451, "Address does not exist or cannot receive mail at this time, try again later", false
		}
		// If we got back nil without error, the maildir doesn't exist, but this is a temporary (hopefully) setup problem
		if maildir == nil {
			logger.Println("User exists but maildir is nil: ", err)
			return 451, "Maildir does not exist; try again later", false
		}
	}

	// At this point, we are willing to accept this recipient
	session.Recipients = append(session.Recipients, *recipient)
	logger.Println("Recipient accepted: ", *addr)
	return 250, "OK", false
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

func (handler *ProtocolHandler) processMAIL(session *Session, line string) (int, string, bool) {
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

func (handler *ProtocolHandler) createReceived(session *Session) (string, error) {
	rcv := "Received: from "
	// remote server info
	rcv += os.Getenv("TCPREMOTEIP")
	rcv += " by "
	rcv += " with SMTP; "
	rcv += time.Now().String()
	rcv += "\n"
	return rcv, nil
}

func (handler *ProtocolHandler) processDATA(session *Session, line string) (int, string, bool) {
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
	rcv, err := handler.createReceived(session)
	if err != nil {
		return 451, "message could not be accepted at this time, try again later", false
	}
	session.Data = rcv

	// Accept the start of message data
	handler.SendCodeLine(354, "Send message content; end with <CRLF>.<CRLF>")
	for finished := false; !finished; {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		logger.Printf(">%v\n", line)
		if strings.HasPrefix(line, ".") {
			if strings.HasPrefix(line, "..") {
				// Remove escaped period character
				line = line[1:len(line)]
			} else {
				// Check with spamc if needed
				if handler.spamc != nil {
					logger.Printf("session.Data is %v bytes", len(session.Data))
					logger.Printf("session.Data:\n%v", session.Data)
					msg, err := handler.checkSpam(session)
					if err != nil {
						// Temporary failure if we can't check it
						return 451, "message could not be accepted at this time, try again later", false
					}
					// We don't block here; let the user use their filters
					session.Data = msg
				}
				err := handler.enqueue(session)
				if err != nil {
					logger.Println("Unable to enqueue message!")
					return 451, "message could not be accepted at this time, try again later", false
				}
				return 250, "message accepted for delivery", false
			}
		}
		logger.Printf("Appended %v bytes to existing %v bytes in session.Data", len(line), len(session.Data))
		session.Data += line
		session.Data += "\n"
		logger.Printf("New session.Data is %v bytes", len(session.Data))
	}
	// If we somehow get here without the message being completed, return a temporary failure
	return 451, "message could not be accepted at this time, try again later", false
}

// CheckSpam runs spamc to see if a message is spam, and returns either an error, or the modified message
func (handler *ProtocolHandler) checkSpam(session *Session) (string, error) {
	if len(*handler.spamc) > 0 {
		cmd := exec.Command(*handler.spamc)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		logger.Printf("Executing spamc: %v", *handler.spamc)
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		logger.Printf("Beginning output processing: %v", *handler.spamc)
		result := ""

		// Create reader and writer
		spamwriter := bufio.NewWriter(stdin)
		spamwriter = bufio.NewWriterSize(spamwriter, len(session.Data))
		// spamreader := bufio.NewReader(stdout)
		// spamreader = bufio.NewReaderSize(spamreader, len(session.Data)+1024)

		// Write and flush
		l, err := spamwriter.WriteString(session.Data)
		logger.Printf("Wrote %v bytes of %v to spamwriter", l, len(session.Data))
		spamwriter.Flush()
		stdin.Close()
		logger.Printf("Message written to spamc")

		// Create a reader at least as big as the original message with extra space for headers

		logger.Printf("Reading message back from spamc")

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			logger.Println(line)
			result += line
			result += "\n"
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		logger.Printf("Waiting for spamc to exit")
		err = cmd.Wait()
		if err != nil {
			log.Fatal(err)
			return "", err
		}
		return result, nil
	}
	logger.Printf("spamc not configured!")
	return session.Data, nil
}

// enqueue places the current message (as contained in the session) into the disk queue; ie accepting delivery
func (handler *ProtocolHandler) enqueue(session *Session) error {
	if session != nil {
		var err error
		env := queue.Envelope{}
		if session.Sender != nil {
			env.Sender, err = session.Sender.ToString()
			if err != nil {
				log.Fatal(err)
				return errors.New("queue attempt failed, invalid sender")
			}
		}

		if session.From != nil {
			env.From, err = session.From.ToString()
			if err != nil {
				log.Fatal(err)
				return errors.New("queue attempt failed, invalid sender")
			}
		} else {
			// This is a bounce message
			env.From = "<>"
		}

		for _, recipient := range session.Recipients {
			er := queue.EnvelopeRecipient{}
			er.Recipient, err = recipient.ToString()
			if err != nil {
				log.Fatal(err)
				return errors.New("queue attempt failed, invalid recipient")
			}
			logger.Printf("adding recipient to envelope: %v", er.Recipient)
			er.Delivered = false
			env.Recipients = append(env.Recipients, er)
		}
		err = handler.msgqueue.Enqueue(env, session.Data)
		if err != nil {
			log.Fatal(err)
			return errors.New("queue attempt failed")
		}
		return nil
	}
	return errors.New("no current session to enqueue")
}

// processQUIT simply terminates the session
func (handler *ProtocolHandler) processQUIT(session *Session, line string) (int, string, bool) {
	return 221, "goodbye", true
}

// processRSET clears the session information
func (handler *ProtocolHandler) processRSET(session *Session, line string) (int, string, bool) {
	session.Sender = nil
	session.From = nil
	session.Recipients = make([]address.Address, 0)
	session.Data = ""
	return 250, "OK", false
}

func (handler *ProtocolHandler) processNOOP(session *Session, line string) (int, string, bool) {
	return 250, "OK", false
}

func (handler *ProtocolHandler) processVRFY(session *Session, line string) (int, string, bool) {
	return 500, "VRFY not supported", false
}
