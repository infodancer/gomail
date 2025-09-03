package main

import (
	"flag"
	"log"
	"os"

	"github.com/infodancer/gomail/config"
	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/queue"

	"github.com/infodancer/gomail/smtpd"
)

var Version string

func main() {
	cfgfile := flag.String("cfg", "/opt/infodancer/gomail/etc/smtpd.toml", "The configuration file")
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if versionFlag != nil && *versionFlag {
		log.Println("Version: " + Version)
		os.Exit(0)
	}

	var cfg smtpd.Config
	err := config.LoadTOMLConfig(*cfgfile, &cfg)
	if err != nil {
		log.Printf("error reading configuration: %v", err)
		os.Exit(1)
	}

	// Initialize the mail queue if not set
	if cfg.MQueue == nil {
		queueDir := os.Getenv("QUEUE_DIR")
		if queueDir == "" {
			queueDir = "/tmp/test-queue"
		}
		cfg.MQueue, err = queue.GetQueue(queueDir)
		if err != nil {
			log.Printf("error initializing queue: %v", err)
			os.Exit(1)
		}
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
