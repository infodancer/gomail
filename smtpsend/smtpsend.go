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

	if con.IsESMTPSupported {

	}
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
