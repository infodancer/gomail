package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/queue"
	"github.com/infodancer/gomail/smtpd"
)

func main() {
	helo := flag.String("helo", "localhost", "The helo string to use when greeting clients")
	cfg := smtpd.Config{
		ServerName: "localhost",
		Banner:     *helo,
		Spamc:      "",
		Maxsize:    0,
		MQueue:     &queue.Queue{},
	}

	var c connect.TCPConnection
	c, err := connect.NewStandardIOConnection()
	if err != nil {
		fmt.Println("error creating new StandardIOConnection")
		os.Exit(1)
	}
	s, err := cfg.Start(c)
	if err != nil {
		fmt.Println("error sending greeting")
		os.Exit(2)
	}
	err = s.HandleConnection()
	if err != nil {
		fmt.Println("error handling connection")
		os.Exit(3)
	}
}
