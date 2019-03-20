package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var logger *log.Logger

func main() {
	sender := flag.String("sender", "h", "The envelope sender to use")
	recipient := flag.String("recipient", "h", "The envelope recipient to use")
	mxhost := flag.String("mx", "", "Override the mx lookup and use the specified host")
	user := flag.String("user", "", "The username to use for smtp auth")
	password := flag.String("password", "", "The password to use for smtp auth")
	msgfile := flag.String("file", "", "The file to read the message from; stdin is used if left blank")
	logger := log.New(os.Stderr, "", 0)
	logger.Println("gomail smtps started")
	msg := ""
	if msgfile != nil {

	}
	sendMessageDirect(mxhost, user, password, sender, recipient, &msg)
}

// sendMessageDirect sends a message directly to the specified host, with a single sender and recipient
// mxhost should be formatted as host:port
func sendMessageDirect(mxhost *string, user *string, password *string, sender *string, recipient *string, msg *string) (int, string) {
	if mxhost == nil {
		return 500, "mxhost must not be nil"
	}
	if recipient == nil {
		return 500, "recipient must not be nil"
	}
	if msg == nil {
		return 500, "msg must not be nil"
	}
	conn, err := net.Dial("tcp", *mxhost)
	if err != nil {
		logger.Println("gomail smtps started")
		return 451, "Connection failed"
	}
	// Wait for banner
	banner, err := bufio.NewReader(conn).ReadString('\n')
	logger.Println("<" + banner)
	// Try EHLO
	fmt.Fprintf(conn, "EHLO \r\n")
	ehloresp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return 451, "Connection dropped or I/O error"
	}
	// Fall back to HELO if it fails
	if !strings.HasPrefix(ehloresp, "250 OK") {
		logger.Println("EHLO failed")
		fmt.Fprintf(conn, "HELO \r\n")
		heloresp, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			logger.Println("HELO failed")
			return 451, ehloresp
		}
		if strings.HasPrefix(heloresp, "250 OK") {
			logger.Println("HELO accepted")

		}
	}

	return 0, ""
}
