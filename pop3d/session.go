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
		response, finished, err := s.HandleInputLine(line)
		if err != nil {
			if err := s.Println("error handling input line"); err != nil {
				return err
			}
			return err
		}
		err = s.SendLine(response)
		if err != nil {
			if err := s.Println("io error sending response"); err != nil {
				return err
			}
			return err
		}
		if finished {
			break
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
func (s Session) HandleInputLine(line string) (string, bool, error) {
	cmd := strings.Split(line, " ")
	command := strings.ToUpper(strings.TrimSpace(cmd[0]))
	switch command {
	// These commands are valid only in the AUTH state
	case "USER":
		response, err := s.processUSER(line)
		return response, false, err
	case "PASS":
		response, err := s.processPASS(line)
		return response, false, err

	// These commands are valid only in the TRANSACTION state
	case "STAT":
		response, err := s.processSTAT(line)
		return response, false, err
	case "LIST":
		response, err := s.processLIST(line)
		return response, false, err
	case "RETR":
		response, err := s.processRETR(line)
		return response, false, err
	case "DELE":
		response, err := s.processDELE(line)
		return response, false, err

	// These commands are not vital
	case "NOOP":
		response, err := s.processNOOP(line)
		return response, false, err

	// QUIT terminates the session
	case "QUIT":
		response, err := s.processQUIT(line)
		return response, true, err
	default:
		return "-ERR", false, errors.New("unrecognized command")
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
