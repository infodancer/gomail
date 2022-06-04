package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/infodancer/gomail/smtp"
)

var logger *log.Logger

// SMTPConnection models an smtp network connection
type SMTPConnection struct {
	host *string
	port int
}

func main() {
	sender := flag.String("sender", "h", "The envelope sender to use")
	recipient := flag.String("recipient", "h", "The envelope recipient to use")
	hostname := flag.String("hostname", "", "Override the mx lookup and use the specified host")
	username := flag.String("username", "", "The username to use for smtp auth")
	password := flag.String("password", "", "The password to use for smtp auth")
	msgfile := flag.String("file", "", "The file to read the message from; stdin is used if left blank")
	logger := log.New(os.Stderr, "", 0)
	logger.Println("gomail smtps started")
	msg := ""
	if msgfile != nil {

	}
	sendMessageDirect(hostname, username, password, sender, recipient, &msg)
}

// OpenSMTPConnection opens a new SmtpConnection for mail; accepts hostnames or ip addresses with ports
func OpenSMTPConnection(host string) (*SMTPConnection, *smtp.Error) {
	if host == "" {
		return nil, smtp.NewError(500, "host must not be nil")
	}
	conn, err := net.Dial("tcp", host)
	if err != nil {
		logger.Println("gomail smtps started")
		return nil, smtp.NewError(451, "Connection failed")
	}

	result := SMTPConnection{}
	// Wait for banner
	banner, err := bufio.NewReader(conn).ReadString('\n')
	logger.Println("<" + banner)
	// Try EHLO
	fmt.Fprintf(conn, "EHLO \r\n")
	ehloresp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, smtp.NewError(451, "Connection dropped or I/O error")
	}
	// Fall back to HELO if it fails
	if !strings.HasPrefix(ehloresp, "250 OK") {
		logger.Println("EHLO failed")
		fmt.Fprintf(conn, "HELO \r\n")
		heloresp, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			logger.Println("HELO failed")
			return nil, smtp.NewError(451, ehloresp)
		}
		if strings.HasPrefix(heloresp, "250 OK") {
			logger.Println("HELO accepted")

		}
	}
	return &result, nil
}

// SendCommand sends a single comand and waits for the single-line response
func (c SMTPConnection) SendCommand(line string) *smtp.Error {
	c.sendLine(line)
	return smtp.NewError(500, "Not yet implemented!")
}

// SendLine sends a single line without waiting for a response
func (c SMTPConnection) readLine() (string, *smtp.Error) {
	return "", smtp.NewError(500, "Not yet implemented")
}

// SendLine sends a single line without waiting for a response
func (c SMTPConnection) sendLine(line string) *smtp.Error {

	return smtp.NewError(500, "Not yet implemented!")
}

// SetSender sets the smto sender address for this connection
func (c SMTPConnection) SetSender(sender string) *smtp.Error {
	return c.SendCommand("MAIL FROM:<" + sender + ">")
}

// AddRecipient adds a recipient for this connection
func (c SMTPConnection) AddRecipient(recipient string) *smtp.Error {
	return c.SendCommand("RCPT TO:<" + recipient + ">")
}

// SendMessage sends a message
func (c SMTPConnection) SendMessage(msg string) *smtp.Error {
	return smtp.NewError(500, "Not yet implemented!")
}

// sendMessageDirect sends a message directly to the specified host, with a single sender and recipient
// mxhost should be formatted as host:port
func sendMessageDirect(host *string, user *string, password *string, sender *string, recipient *string, msg *string) *smtp.Error {
	if recipient == nil {
		return smtp.NewError(500, "recipient must not be nil")
	}
	if msg == nil {
		return smtp.NewError(500, "msg must not be nil")
	}

	conn, err := OpenSMTPConnection(*host)
	if err != nil {
		return smtp.NewError(451, "Connection could not be opened, or HELO/EHLO negotiation failed")
	}

	conn.SetSender(*sender)
	conn.AddRecipient(*recipient)
	conn.SendMessage(*msg)
	return nil
}
