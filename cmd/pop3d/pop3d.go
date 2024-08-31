package main

import (
	"flag"
	"log"
	"os"

	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/pop3d"
)

var Version string

func main() {
	cfgfile := flag.String("cfg", "/opt/infodancer/gomail/etc/pop3d.json", "The configuration file")
	flag.Parse()

	cfg, err := pop3d.ReadConfigFile(*cfgfile)
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
	var s *pop3d.Session
	s, err = cfg.Start(&c)
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
