package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// SMTPConnection describes an outgoing SMTP connection
type SMTPConnection struct {
	// Started indicates the time the connection was opened
	Started time.Time
	// Activity indicates the last activity on the connection
	Activity time.Time
	// RemoteHost indicates the hostname and port
	RemoteHost string
	// IsESMTPSupported indicates whether expanded SMTP is supported
	IsESMTPSupported bool
	// Reader contains a read buffer
	Reader *bufio.Reader
	// Conn contains the opened connection
	Conn net.Conn
}

var sender *string
var recipient *string
var hostname *string
var username *string
var password *string
var authmethod *string

func main() {
	hostname = flag.String("hostname", "", "The hostname or ip address to connect to, with optional port number")
	username = flag.String("username", "", "The user to authenticate as")
	password = flag.String("password", "", "The password to authenticate with")
	authmethod = flag.String("authmethod", "", "The authentication method to use")
	sender = flag.String("sender", "", "The envelope sender")
	recipient = flag.String("recipient", "", "The envelope recipient")
	flag.Parse()

	if sender == nil {
		fmt.Println("sender is a required parameter")
		os.Exit(1)
	}

	if recipient == nil {
		fmt.Println("recipient is a required parameter")
		os.Exit(1)
	}

	con, err := createSMTPConnection(hostname)
	if err != nil {
		fmt.Println("Connection failed!")
		os.Exit(1)
	}

	heloname := "localhost"
	if con.IsESMTPSupported {
		err = con.SendEHLO(&heloname)
	} else {
		err = con.SendHELO(&heloname)
	}

	if err != nil {
		fmt.Println("Remote server not ready; greeting failed")
		os.Exit(1)
	}

	// Handle authentication here

	// Actually send the message
	con.SendMailFrom(sender)
	if err != nil {
		fmt.Println("Remote server not ready; greeting failed")
		os.Exit(1)
	}

	con.SendRcptTo(recipient)
	if err != nil {
		fmt.Println("Remote server not ready; greeting failed")
		os.Exit(1)
	}
	/*
		con.SendData()
		if err != nil {
			fmt.Println("Remote server not ready; greeting failed")
			os.Exit(1)
		}
	*/
}

func createSMTPConnection(hostname *string) (*SMTPConnection, error) {
	host := hostname
	if !isPortSpecified(hostname) {
		// Default to port 25
		*host += ":25"
	}

	conn, err := net.Dial("tcp", *host)
	if err != nil {
		return nil, err
	}

	result := SMTPConnection{}
	result.Conn = conn
	result.Reader = bufio.NewReader(conn)

	// Check that the server is accepting connections
	line, err := result.ReadLine()
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(*line, "220") {
		if strings.Contains(*line, "ESMTP") {
			result.IsESMTPSupported = true
		}
		return &result, nil
	}
	return nil, errors.New(*line)
}

// SendMail sends the actual message
func (c SMTPConnection) SendMail(from *string, to *string, data *string) error {
	var err error
	err = c.SendMailFrom(from)
	if err != nil {
		return err
	}

	err = c.SendRcptTo(from)
	if err != nil {
		return err
	}

	return nil
}

// SendEHLO sends the greeting
func (c SMTPConnection) SendEHLO(clienthost *string) error {
	line := "EHLO " + *clienthost
	err := c.SendLine(&line)
	if err != nil {
		return err
	}
	return nil
}

// SendHELO sends the greeting
func (c SMTPConnection) SendHELO(clienthost *string) error {
	line := "HELO " + *clienthost
	err := c.SendLine(&line)
	if err != nil {
		return err
	}
	return nil
}

// SendMailFrom sends the greeting
func (c SMTPConnection) SendMailFrom(sender *string) error {
	line := "MAIL FROM: <" + *sender + ">"
	err := c.SendLine(&line)
	if err != nil {
		return err
	}
	return nil
}

// SendRcptTo sends the greeting
func (c SMTPConnection) SendRcptTo(sender *string) error {
	line := "RCPT TO: <" + *sender + ">"
	err := c.SendLine(&line)
	if err != nil {
		return err
	}
	return nil
}

func isPortSpecified(hostname *string) bool {
	return strings.Contains(*hostname, ":")
}

// SendLine sends a line of data (adding line terminators)
func (c SMTPConnection) SendLine(line *string) error {
	_, err := fmt.Fprintf(c.Conn, "%v\r\n", line)
	fmt.Println(">" + *line + "\r\n")
	return err
}

func (c SMTPConnection) sendLine(line *string) error {
	_, err := fmt.Fprintf(c.Conn, "%v\r\n", line)
	return err
}

// ReadLine reads a line from the connection
func (c SMTPConnection) ReadLine() (*string, error) {
	line, err := c.Reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	fmt.Println("<" + line)
	return &line, nil
}

// Close closes the connection
func (c SMTPConnection) Close() error {
	return nil
}
