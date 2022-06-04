package main

import (
	"bufio"
	"log"
	"os"

	"github.com/infodancer/gomail/connect"
)

var logger *log.Logger

func main() {
	logger = log.New(os.Stderr, "", 0)

	cfg := Config{
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
	s, err := cfg.Start(&c)
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
