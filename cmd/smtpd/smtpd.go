package main

import (
	"flag"
	"log"
	"os"

	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/smtpd"
)

func main() {
	cfgfile := flag.String("cfg", "/opt/infodancer/gomail/etc/smtpd.json", "The configuration file")
	flag.Parse()

	cfg, err := smtpd.ReadConfigFile(*cfgfile)
	if err != nil {
		log.Println("error reading configuration: %w", err)
		os.Exit(1)
	}
	var c connect.TCPConnection
	c, err = connect.NewStandardIOConnection()
	if err != nil {
		log.Println("error creating new StandardIOConnection")
		os.Exit(1)
	}
	s, err := cfg.Start(c)
	if err != nil {
		log.Println("error sending greeting")
		os.Exit(2)
	}
	err = s.HandleConnection()
	if err != nil {
		log.Println("error handling connection")
		os.Exit(3)
	}
}
