package smtpd

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/infodancer/gomail/address"
	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/domain"
)

// Session describes the current session
type Session struct {
	// Config holds the server configuration
	Config Config
	// Connection holds the client connection information
	Conn connect.TCPConnection
	// Sender is the authenticated user sending the message; nil if not authenticated
	Sender string
	// From is the claimed sender of the message
	From string
	// Recipients is the array of recipients
	Recipients     []string
	RecipientLimit int

	// Headers are only the headers this mail system is adding
	Headers []string
	// Data contains all the data received from the client
	Data string

	// maxsize is the max message size in bytes for this session
	maxsize int64
}

func Create(cfg Config, conn connect.TCPConnection) *Session {
	s := Session{Config: cfg, Conn: conn}

	return &s
}

func (s *Session) Printf(v ...any) error {
	_, err := fmt.Fprintf(os.Stderr, v[0].(string), v[1:])
	return err
}

func (s *Session) Println(v ...any) error {
	_, err := fmt.Fprintln(os.Stderr, v...)
	return err
}

func (s *Session) HandleConnection() error {
	defer func() {
		if err := s.Conn.Close(); err != nil {
			s.Conn.Logger().Printf("error: %s", err)
		}
	}()
	for {
		line, err := s.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			if err := s.Println("io error reading from connection"); err != nil {
				s.Conn.Logger().Printf("error: %s", err)
			}
			break
		}
		code, message, finished := s.HandleInputLine(line)
		err = s.SendCodeLine(code, message)
		if err != nil {
			if err := s.Println("io error sending response"); err != nil {
				s.Conn.Logger().Printf("error: %s", err)
			}
			break
		}
		if finished {
			break
		}
	}
	return nil
}

// SendCodeLine accepts a line without linefeeds and sends it with a CRLF and the provided response code
func (s *Session) SendCodeLine(code int, line string) error {
	cline := fmt.Sprintf("%d %s", code, line)
	if err := s.Println("S:" + cline); err != nil {
		s.Conn.Logger().Printf("error: %s", err)
	}
	return s.Conn.WriteLine(cline + "\r\n")
}

// SendLine accepts a line without linefeeds and sends it with a CRLF and the provided response code
func (s *Session) SendLine(line string) error {
	if err := s.Println("S:" + line); err != nil {
		s.Conn.Logger().Printf("error: %s", err)
	}
	return s.Conn.WriteLine(line)
}

// ReadLine reads a line
func (s *Session) ReadLine() (string, error) {
	return s.Conn.ReadLine()
}

// HandleInputLine accepts a line and handles it
func (s *Session) HandleInputLine(line string) (int, string, bool) {
	var err error
	cmd := strings.Split(line, " ")
	command := strings.ToUpper(strings.TrimSpace(cmd[0]))
	switch command {
	case "HELO":
		return s.processHELO(line)
	case "EHLO":
		{
			// This is a bit of a special case because of extensions
			err = s.SendLine("250-8BITMIME")
			if err != nil {
				return 500, "i/o error", false
			}
			err = s.SendLine("250-PIPELINING")
			if err != nil {
				return 500, "i/o error", false
			}
			err = s.SendLine("250-AUTH CRAM-MD5")
			if err != nil {
				return 500, "i/o error", false
			}
			if s.Config.Maxsize != 0 && s.maxsize != 0 {
				size := strconv.FormatInt(s.maxsize, 10)
				err = s.SendLine("250-SIZE " + size)
				if err != nil {
					return 500, "Unrecognized command", false
				}
			}
			return s.processEHLO(line)
		}
	case "AUTH":
		return s.processAUTH(line)
	case "RCPT":
		return s.processRCPT(line)
	case "MAIL":
		return s.processMAIL(line)
	case "DATA":
		return s.processDATA(line)

	// These commands are not vital
	case "RSET":
		return s.processRSET(line)
	case "NOOP":
		return s.processNOOP(line)
	case "VRFY":
		return s.processVRFY(line)

	// QUIT terminates the session
	case "QUIT":
		return s.processQUIT(line)
	default:
		return 500, "Unrecognized command", false
	}
}

// processHELO handles the standard SMTP helo
func (s *Session) processHELO(line string) (int, string, bool) {
	return 250, s.Config.ServerName, false
}

// processEHLO handles the extended EHLO command, but the extensions are listed elsewhere
func (s *Session) processEHLO(line string) (int, string, bool) {
	return 250, s.Config.ServerName, false
}

// processQUIT simply terminates the session
func (s *Session) processQUIT(line string) (int, string, bool) {
	return 221, "goodbye", true
}

// processRSET clears the session information
func (s *Session) processRSET(line string) (int, string, bool) {
	s.Sender = ""
	s.From = ""
	s.Recipients = make([]string, 0)
	s.Data = ""
	return 250, "OK", false
}

func (s *Session) processNOOP(line string) (int, string, bool) {
	return 250, "OK", false
}

func (s *Session) processVRFY(line string) (int, string, bool) {
	return 500, "VRFY not supported", false
}

// processAUTH handles the auth process
func (s *Session) processAUTH(line string) (int, string, bool) {
	// For now, we haven't implemented this
	if strings.HasPrefix(line, "AUTH ") {
		authType := line[5:]
		// Reject insecure authentication methods
		if authType != "CRAM-MD5" {
			return 500, "Unrecognized command", false
		}
		challenge := createChallenge()
		err := s.SendCodeLine(354, challenge)
		if err != nil {
			return 550, "i/o error", false
		}
		resp, err := s.ReadLine()
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

func (s *Session) processRCPT(line string) (int, string, bool) {
	addr, err := extractAddressPart(line)
	if err != nil {
		return 550, "Invalid address", false
	}
	// Check if the sender has been set
	if len(s.From) == 0 {
		return 503, "need MAIL before RCPT", false
	}
	// Check for number of recipients
	if len(s.Recipients) >= s.RecipientLimit {
		if err := s.Printf("Rejecting RCPT TO %d recipients already", len(s.Recipients)); err != nil {
			s.Conn.Logger().Print(err)
		}
		return 452, "Too many recipients", false
	}
	// Check if this is being sent to a bounce address
	if len(*addr) == 0 {
		if err := s.Println("Rejecting RCPT TO to bounce address: " + *addr); err != nil {
			s.Conn.Logger().Print(err)
		}
		return 503, "We don't accept mail to that address", false
	}

	// Before we actually do filesystem operations, sanitize the input
	if IsSuspiciousAddress(*addr) {
		if err := s.Println("Rejecting suspicious RCPT TO: " + *addr); err != nil {
			s.Conn.Logger().Print(err)
		}
		return 550, "Invalid address", false
	}

	recipient, err := address.CreateAddress(*addr)
	if err != nil {
		return 550, "Invalid address", false
	}

	// Check for relay and allow only if sender has authenticated
	dom, err := domain.GetDomain(recipient.Domain)
	if err != nil {
		return 550, "Invalid address", false
	}
	if len(s.Sender) == 0 {
		// Only bother to check domain if the sender is nil
		if dom == nil {
			return 551, "We don't relay mail to remote addresses", false
		}
	}

	// Check for local recipient existing if the domain is local
	if dom != nil {
		// We know the domain exists locally now
		user, err := dom.GetUser(recipient.User)
		// Temporary error if we couldn't access the user for some reason
		if err != nil {
			if err := s.Println("Error from GetUser: ", err); err != nil {
				s.Conn.Logger().Print(err)
			}
			return 451, "Address does not exist or cannot receive mail at this time, try again later", false
		}
		// If we got back nil without error, they really don't exist
		if user == nil {
			return 550, "User does not exist", false
		}
		// But if they do exist, check that their mailbox also exists
		maildir, err := dom.GetUserMaildir(recipient.User)
		if err != nil {
			if err := s.Println("User exists but GetUserMaildir errors: ", err); err != nil {
				s.Conn.Logger().Print(err)
			}
			return 451, "Address does not exist or cannot receive mail at this time, try again later", false
		}
		// If we got back nil without error, the maildir doesn't exist, but this is a temporary (hopefully) setup problem
		if maildir == nil {
			if err := s.Println("User exists but maildir is nil: ", err); err != nil {
				s.Conn.Logger().Print(err)
			}
			return 451, "Maildir does not exist; try again later", false
		}
	}

	// At this point, we are willing to accept this recipient
	s.Recipients = append(s.Recipients, recipient.String())
	if err := s.Println("Recipient accepted: ", *addr); err != nil {
		s.Conn.Logger().Print(err)
	}
	return 250, "OK", false
}

func (s *Session) processMAIL(line string) (int, string, bool) {
	if len(s.From) > 0 {
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

	s.From = *addr
	return 250, "OK", false
}

func (s *Session) processDATA(line string) (int, string, bool) {
	// Did the user specify an envelope?
	// Check if the sender has been set
	if len(s.From) == 0 {
		return 503, "need MAIL before DATA", false
	}
	// Check for number of recipients
	if len(s.Recipients) == 0 {
		return 503, "need RCPT before DATA", false
	}
	// Generate a received header
	rcv, err := s.createReceived()
	if err != nil {
		return 451, "message could not be accepted at this time, try again later", false
	}
	s.AddHeader(rcv)

	// Accept the start of message data
	err = s.SendCodeLine(354, "Send message content; end with <CRLF>.<CRLF>")
	if err != nil {
		return 451, "message could not be accepted at this time, try again later", false
	}
	for finished := false; !finished; {
		line, err := s.ReadLine()
		if err != nil {
			break
		}
		if err := s.Printf(">%v\n", line); err != nil {
			s.Conn.Logger().Print(err)
		}
		if strings.HasPrefix(line, ".") {
			if strings.HasPrefix(line, "..") {
				// Remove escaped period character
				line = line[1:]
			} else {
				// Check with spamc if needed
				if len(s.Config.Spamc) > 0 {
					err = s.Printf("session.Data is %v bytes", len(s.Data))
					if err != nil {
						return 451, "i/o error", false
					}
					err = s.Printf("session.Data:\n%v", s.Data)
					if err != nil {
						return 451, "i/o error", false
					}
					msg, err := s.checkSpam()
					if err != nil {
						return 451, "i/o error", false
					}
					// We don't block here; let the user use their filters
					s.Data = msg
				}
				err := s.enqueue()
				if err != nil {
					if err := s.Println("Unable to enqueue message!"); err != nil {
						s.Conn.Logger().Print(err)
					}
					return 451, "message could not be accepted at this time, try again later", false
				}
				return 250, "message accepted for delivery", false
			}
		}
		if err := s.Printf("Appended %v bytes to existing %v bytes in session.Data", len(line), len(s.Data)); err != nil {
			s.Conn.Logger().Print(err)
		}
		s.Data += line
		s.Data += "\n"
		if err := s.Printf("New session.Data is %v bytes", len(s.Data)); err != nil {
			s.Conn.Logger().Print(err)
		}
	}
	// If we somehow get here without the message being completed, return a temporary failure
	return 451, "message could not be accepted at this time, try again later", false
}

func (s *Session) createReceived() (string, error) {
	rcv := "Received: from "
	// remote server info
	rcv += s.Conn.GetTCPRemoteIP()
	rcv += " by "
	rcv += " with SMTP; "
	rcv += time.Now().String()
	rcv += "\n"
	return rcv, nil
}

func (s *Session) AddHeader(h string) {
	s.Headers = append(s.Headers, h)
}

// enqueue places the current message (as contained in the session) into the disk queue; ie accepting delivery
func (s *Session) enqueue() error {
	err := s.Config.MQueue.Enqueue(s.From, s.Recipients, []byte(s.Data))
	if err != nil {
		log.Fatal(err)
		return errors.New("queue attempt failed")
	}
	return nil
}

// CheckSpam runs spamc to see if a message is spam, and returns either an error, or the modified message
func (s *Session) checkSpam() (string, error) {
	if len(s.Config.Spamc) > 0 {
		cmd := exec.Command(s.Config.Spamc)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return "", err
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return "", err
		}

		if err := s.Printf("Executing spamc: %v", s.Config.Spamc); err != nil {
			s.Conn.Logger().Print(err)
		}
		if err := cmd.Start(); err != nil {
			return "", err
		}

		if err := s.Printf("Beginning output processing: %v", s.Config.Spamc); err != nil {
			s.Conn.Logger().Print(err)
		}
		result := ""

		// Create reader and writer
		spamwriter := bufio.NewWriter(stdin)
		spamwriter = bufio.NewWriterSize(spamwriter, len(s.Data))
		// spamreader := bufio.NewReader(stdout)
		// spamreader = bufio.NewReaderSize(spamreader, len(session.Data)+1024)

		// Write and flush
		l, err := spamwriter.WriteString(s.Data)
		if err != nil {
			return "", err
		}
		if err := s.Printf("Wrote %v bytes of %v to spamwriter", l, len(s.Data)); err != nil {
			s.Conn.Logger().Print(err)
		}
		if err := spamwriter.Flush(); err != nil {
			s.Conn.Logger().Print(err)
		}
		if err := stdin.Close(); err != nil {
			s.Conn.Logger().Print(err)
		}
		if err := s.Printf("Message written to spamc"); err != nil {
			s.Conn.Logger().Print(err)
		}

		// Create a reader at least as big as the original message with extra space for headers

		if err := s.Printf("Reading message back from spamc"); err != nil {
			s.Conn.Logger().Print(err)
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if err := s.Println(line); err != nil {
				s.Conn.Logger().Print(err)
			}
			result += line
			result += "\n"
		}

		if err := scanner.Err(); err != nil {
			return "", err
		}

		if err := s.Printf("Waiting for spamc to exit"); err != nil {
			s.Conn.Logger().Print(err)
		}
		err = cmd.Wait()
		if err != nil {
			return "", err
		}
		return result, nil
	}
	if err := s.Printf("spamc not configured!"); err != nil {
		s.Conn.Logger().Print(err)
	}
	return s.Data, nil
}

// extractAddress parses an SMTP command line for an @ address within <>
func extractAddressPart(line string) (*string, error) {
	begin := strings.Index(line, "<") + 1
	end := strings.LastIndex(line, ">")
	if begin == -1 || end == -1 {
		return nil, errors.New("address not found in command")
	}
	value := line[begin:end]
	// RFC 5321 https://tools.ietf.org/html/rfc5321#section-4.5.3
	if len(value) > 254 {
		return nil, errors.New("address exceeds maximum length of email address")
	}
	return &value, nil
}

// extractUsername decodes the username from the client's response
func extractUsername(resp string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(resp)
	if err != nil {
		return "", errors.New("base64 decode failed")
	}
	// s.Println("CRAM-MD5 response: ", decoded)
	return string(decoded), nil
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

// IsSuspiciousInput looks for input that contains filename elements
// This method should be used to check addresses or domain names coming from external sources
// It's not perfect, but it works for now
func IsSuspiciousAddress(input string) bool {
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
