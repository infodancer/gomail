package main

import (
	"bufio"
	"log"
	"os"

	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/pop3d"
)

var logger *log.Logger

func main() {
	logger = log.New(os.Stderr, "", 0)

	cfg := pop3d.Config{
		ServerName: "",
	}

	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	var c connect.TCPConnection
	c, err := connect.NewStandardIOConnection(r, w)
	if err != nil {
		logger.Println("error creating new StandardIOConnection")
		os.Exit(1)
	}
	var s *pop3d.Session
	s, err = cfg.Start(&c)
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
