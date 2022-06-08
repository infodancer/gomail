package main

import (
	"flag"
	"log"
	"os"

	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/queue"
	"github.com/infodancer/gomail/smtpd"
)

var recipientLimit *int
var logger *log.Logger

func main() {
	helo := flag.String("helo", "h", "The helo string to use when greeting clients")

	logger = log.New(os.Stderr, "", 0)

	cfg := smtpd.Config{
		ServerName: "",
		Banner:     *helo,
		Spamc:      "",
		Maxsize:    0,
		MQueue:     &queue.Queue{},
	}

	var c connect.TCPConnection
	c, err := connect.NewStandardIOConnection()
	if err != nil {
		logger.Println("error creating new StandardIOConnection")
		os.Exit(1)
	}
	s, err := cfg.Start(c)
	if err != nil {
		logger.Println("error sending greeting")
		os.Exit(2)
	}
	err = s.HandleConnection()
	if err != nil {
		logger.Println("error handling connection")
		os.Exit(3)
	}
}
