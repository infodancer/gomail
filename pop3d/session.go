package pop3d

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/infodancer/gomail/connect"
)

// Session describes the current session
type Session struct {
	// Config holds the server configuration
	Config Config
	// Connection holds the client connection information
	Conn connect.TCPConnection
	// State holds the state of the session
	State SessionState
}

type SessionState int

const (
	STATE_AUTHORIZATION SessionState = iota + 1
	STATE_TRANSACTION
	STATE_UPDATE
)

func (s Session) HandleConnection() error {
	for {
		line, err := s.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			if err := s.Println("io error reading from connection"); err != nil {
				return err
			}
		}
		_, err = s.HandleInputLine(line)
		if err != nil {
			if err := s.Println("error handling input line"); err != nil {
				return err
			}
			return err
		}
	}
	return nil
}

// SendLine accepts a line without linefeeds and sends it with a CRLF and the provided response code
func (s Session) SendLine(line string) error {
	if err := s.Println("S:" + line); err != nil {
		return err
	}
	return s.Conn.WriteLine(line)
}

func (s *Session) Printf(v ...any) error {
	_, err := fmt.Fprintf(os.Stderr, v[0].(string), v[1:])
	return err
}

func (s *Session) Println(v ...any) error {
	_, err := fmt.Fprintln(os.Stderr, v...)
	return err
}

// ReadLine reads a line
func (s Session) ReadLine() (string, error) {
	return s.Conn.ReadLine()
}

// HandleInputLine accepts a line and handles it
func (s Session) HandleInputLine(line string) (string, error) {
	cmd := strings.Split(line, " ")
	command := strings.ToUpper(strings.TrimSpace(cmd[0]))
	switch command {
	// These commands are valid only in the AUTH state
	case "USER":
		return s.processUSER(line)
	case "PASS":
		return s.processPASS(line)

	// These commands are valid only in the TRANSACTION state
	case "STAT":
		return s.processSTAT(line)
	case "LIST":
		return s.processLIST(line)
	case "RETR":
		return s.processRETR(line)
	case "DELE":
		return s.processDELE(line)

	// These commands are not vital
	case "NOOP":
		return s.processNOOP(line)

	// QUIT terminates the session
	case "QUIT":
		return s.processQUIT(line)
	default:
		return "-ERR", errors.New("unrecognized command")
	}
}

// processUSER does nothing
func (s Session) processDELE(line string) (string, error) {

	return "-ERR", errors.New("not yet implemented")
}

// processRETR does nothing
func (s Session) processRETR(line string) (string, error) {

	return "-ERR", errors.New("not yet implemented")
}

// processLIST does nothing
func (s Session) processLIST(line string) (string, error) {

	return "-ERR", errors.New("not yet implemented")
}

// processSTAT does nothing
func (s Session) processSTAT(line string) (string, error) {

	return "-ERR", errors.New("not yet implemented")
}

// processUSER does nothing
func (s Session) processUSER(line string) (string, error) {

	return "-ERR", errors.New("not yet implemented")
}

// processPASS does nothing
func (s Session) processPASS(line string) (string, error) {

	return "-ERR", errors.New("not yet implemented")
}

// processNOOP does nothing
func (s Session) processNOOP(line string) (string, error) {
	return "+OK", nil
}

// processQUIT simply terminates the session
func (s Session) processQUIT(line string) (string, error) {
	return "+OK goodbye", nil
}
