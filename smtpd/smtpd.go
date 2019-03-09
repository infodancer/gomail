package smtpd

import "fmt"
import "os"
import "log"
import "bufio"
import "strings"
import "errors"
import "flag"
import "time"
import "gomail/address"

// Session describes the current session
type Session struct {
	// Started indicates the time the session began
	Started time.Time
	// Activity indicates the time of last activity
	Activity time.Time

	// Sender is the authenticated user sending the message; nil if not authenticated
	Sender address.Address
	// From is the claimed sender of the message
	From address.Address
	// Recipients is the array of recipients
	Recipients []address.Address
	// Data is the array of lines in the message itself
	Data []string
}

var recipientLimit *int
var logger *log.Logger
var reader *bufio.Reader
var helo *string

func main() {
	helo = flag.String("helo", "h", "The helo string to use when greeting clients")
	recipientLimit = flag.Int("maxrcpt", 100, "The maximum number of recipients on a single message")

	logger = log.New(os.Stderr, "", 0)
	logger.Println("gomail smtpd started")
	session := Session{}
	handleConnection(session)
}

// sendLine accepts a line without linefeeds and sends it with network linefeeds
func sendCodeLine(code int, line string) {
	fmt.Print(code, " ", line, "\r\n")
}

// sendLine accepts a line without linefeeds and sends it with network linefeeds
func sendLine(line string) {
	fmt.Print(line, "\r\n")
}

func handleConnection(session Session) {
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

func handleInputLine(session Session, line string) (int, string, bool) {
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
			return processEHLO(session, line)
		}
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
func extractAddress(line string) (string, error) {
	begin := strings.Index(line, "<") + 1
	end := strings.LastIndex(line, ">")
	if begin == -1 || end == -1 {
		return "", errors.New("Address not found in command")
	}
	if strings.Index(address, "@") == -1 {
		return "", errors.New("Address not found in command")
	}
	return line[begin:end], nil
}

// processHELO handles the standard SMTP helo
func processHELO(session Session, line string) (int, string, bool) {
	return 250, "Hello", false
}

// processEHLO handles the extended EHLO command, but the extensions are listed elsewhere
func processEHLO(session Session, line string) (int, string, bool) {
	return 250, "Hello", false
}

func processRCPT(session Session, line string) (int, string, bool) {
	address, err := extractAddress(line)
	if err != nil {
		return 550, "Invalid address", false
	}
	// Check if the sender has been set
	if len(session.From) == 0 {
		return 503, "need MAIL before RCPT", false
	}
	// Check for number of recipients
	if len(session.Recipients) >= *recipientLimit {
		return 452, "Too many recipients", false
	}
	// Check if this is being sent to a bounce address
	if len(address) == 0 {
		return 503, "We don't accept mail to that address", false
	}
	// Check for relay and allow only if sender has authenticated
	if !isLocalAddress(address) && len(session.Sender) == 0 {
		return 553, "We don't relay mail to remote addresses", false
	}

	session.Recipients = append(session.Recipients, address)
	return 250, "OK", false
}

func processMAIL(session Session, line string) (int, string, bool) {
	if len(session.From) > 0 {
		return 400, "MAIL FROM already sent", false
	}
	address, err := extractAddress(line)
	if err != nil {
		return 451, "Invalid address", false
	}
	// Check if this is a bounce message
	if len(address) == 0 {
		return 551, "We don't accept mail to that address", false
	}

	session.From = address
	return 250, "OK", false
}

func processDATA(session Session, line string) (int, string, bool) {
	fmt.Print("354 Send message content; end with <CRLF>.<CRLF>\r\n")
	for finished := false; !finished; {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		logger.Println(">" + line)
		session.Data = append(session.Data, line)
		if strings.HasPrefix(line, ".") && !strings.HasPrefix(line, "..") {
			err := enqueue()
			if err != nil {
				logger.Println("Unable to enqueue message!")
				fmt.Print("451 message could not be accepted at this time, try again later\r\n")
			} else {
				fmt.Print("250 message accepted for delivery\r\n")
			}
			break
		}
	}
	return 0, "", false
}

// enqueue places the current message (as contained in the session) into the disk queue; ie accepting delivery
func enqueue() error {
	return errors.New("Queuing code not yet implemented")
}

// processQUIT simply terminates the session
func processQUIT(session Session, line string) (int, string, bool) {
	return 221, "goodbye", true
}

// processRSET clears the session information
func processRSET(session Session, line string) (int, string, bool) {
	session.Sender = ""
	session.From = ""
	session.Recipients = make([]string, 0)
	session.Data = make([]string, 0)
	return 250, "OK", false
}

func processNOOP(session Session, line string) (int, string, bool) {
	return 250, "OK", false
}

func processVRFY(session Session, line string) (int, string, bool) {
	return 500, "VRFY not supported", false
}

func isLocalAddress(address string) bool {

}
