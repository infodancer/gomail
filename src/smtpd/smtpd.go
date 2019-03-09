package main

import 	"fmt"
import  "os"
import	"log"
import "bufio"
import "strings"
import "errors"
import "flag"

type Session struct {
	// Sender is the authenticated user sending the message; nil if not authenticated
	Sender string 
	// From is the claimed sender of the message
	From string
	// Recipients is the array of recipients
	Recipients []string
	// Data is the array of lines in the message itself
	Data []string
}

var session Session
var logger *log.Logger 
var reader *bufio.Reader
var helo *string

func main() {
	helo = flag.String("helo", "h", "The helo string to use when greeting clients")
	
	logger = log.New(os.Stderr, "", 0)
	logger.Println("gomail smtpd started")
	session = Session{}
	handleConnection()
}

func handleConnection() {
	fmt.Printf("220 " + *helo + " ESMTP\r\n")
	
	reader = bufio.NewReader(os.Stdin)	
	for finished := false; !finished; {
		line, err := reader.ReadString('\n')
		if err != nil {
			break;
		}
		logger.Println(">" + line)
		finished = handleInputLine(line)
		if finished {
			break;
		}
	}
	fmt.Printf("221 Bye\r\n")	
}

func handleInputLine(line string) bool {
	var result bool
	cmd := strings.Split(line," ")
	command := strings.ToUpper(strings.TrimSpace(cmd[0]))
	switch command {
		case "HELO": {
			result = processHELO(line)
		}
		case "EHLO": {
			result = processEHLO(line)
		}
		case "RCPT": {
			result = processRCPT(line)
		}
		case "MAIL": {
			result = processMAIL(line)
		}
		case "DATA": {
			result = processDATA(line)
		}
		
		// These commands are not vital
		case "RSET": {
			result = processRSET(line)
		}
		case "NOOP": {
			result = processNOOP(line)
		}
		case "VRFY": {
			result = processVRFY(line)
		}
		
		// QUIT terminates the session
		case "QUIT": {
			result = true
		}
		default: {
			result = false
		}
	}
	return result
}

// extractAddress parses an SMTP command line for an @ address within <>
func extractAddress(line string) (string, error) {
	begin := strings.Index(line, "<")
	end := strings.LastIndex(line, ">")
	if begin == -1 || end == -1 {
		return "", errors.New("Address not found in command")
	}
	return line[begin:end], nil
}

func processHELO(line string) bool {
	fmt.Println("250 Hello\r\n")
	return false
}

func processEHLO(line string) bool {
	fmt.Println("250 Hello\r\n")
	return false
}

func processRCPT(line string) bool {
	address, err := extractAddress(line)
	if err != nil {
		fmt.Print("250 OK\r\n")
		return false
	}
	session.Recipients = append(session.Recipients, address)
	fmt.Print("250 OK\r\n")
	return false
}

func processMAIL(line string) bool {
	if len(session.From) == 0 {
		address, err := extractAddress(line)
		if err != nil {
			fmt.Print("250 OK\r\n")
			return false
		}

		session.From = address
		fmt.Print("250 OK\r\n")
	} else {
		fmt.Print("400 MAIL FROM already sent\r\n")
	}
	return false
}

func processDATA(line string) bool {
	fmt.Print("354 Send message content; end with <CRLF>.<CRLF>\r\n")	
	for finished := false; !finished; {
		line, err := reader.ReadString('\n')
		if err != nil {
			break;
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
			break;
		}
	}	
	return false
}

// enqueue places the current message (as contained in the session) into the disk queue; ie accepting delivery
func enqueue() error {
	return errors.New("Queuing code not yet implemented")
}

// processQUIT simply terminates the session
func processQUIT(line string) bool {
	return true
}

// processRSET clears the session information
func processRSET(line string) bool {
	session = Session{} 
	return false
}

func processNOOP(line string) bool {
	return false
}

func processVRFY(line string) bool {
	return false
}