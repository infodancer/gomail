package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"

	"github.com/infodancer/gomail/config"
)

var Version string

type ListenerConfig struct {
	config.ServerConfig
	// Command is the command to execute for each connection
	Command string `toml:"command"`
	// Args are the arguments to pass to the command
	Args []string `toml:"args"`
}

func main() {
	cfgfile := flag.String("cfg", "/opt/infodancer/gomail/etc/listener.toml", "The configuration file")
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

	// Validate configuration
	if cfg.Command == "" {
		log.Printf("error: command not specified in configuration")
		os.Exit(1)
	}

	// Start listening
	address := fmt.Sprintf("%s:%d", cfg.Listener.IPAddress, cfg.Listener.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("error starting listener on %s: %v", address, err)
		os.Exit(1)
	}
	defer listener.Close()

	log.Printf("listening on %s", address)

	// Handle connections
	var wg sync.WaitGroup
	connectionCount := 0

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}

		// Check max connections limit
		if cfg.Listener.MaxConnections > 0 && connectionCount >= cfg.Listener.MaxConnections {
			log.Printf("maximum connections (%d) reached, rejecting connection", cfg.Listener.MaxConnections)
			conn.Close()
			continue
		}

		connectionCount++
		wg.Add(1)

		go func(c net.Conn) {
			defer wg.Done()
			defer func() {
				connectionCount--
				c.Close()
			}()

			handleConnection(c, cfg)
		}(conn)
	}
}

func handleConnection(conn net.Conn, cfg smtpd.Config) {
	log.Printf("handling connection from %s", conn.RemoteAddr())

	// Start the smtpd command - we'll hardcode this for now
	// In the future this could be configurable
	cmd := exec.Command("./bin/smtpd", "-cfg", "/opt/infodancer/gomail/etc/smtpd.toml")
	
	// Get pipes for stdin/stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("error creating stdin pipe: %v", err)
		return
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("error creating stdout pipe: %v", err)
		stdin.Close()
		return
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		log.Printf("error starting command: %v", err)
		stdin.Close()
		stdout.Close()
		return
	}

	// Create a wait group for the goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Copy from network connection to command stdin
	go func() {
		defer wg.Done()
		defer stdin.Close()
		
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					log.Printf("error reading from connection: %v", err)
				}
				break
			}
			
			_, err = stdin.Write([]byte(line))
			if err != nil {
				log.Printf("error writing to command stdin: %v", err)
				break
			}
		}
	}()

	// Copy from command stdout to network connection
	go func() {
		defer wg.Done()
		defer conn.Close()
		
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text() + "\n"
			_, err := conn.Write([]byte(line))
			if err != nil {
				log.Printf("error writing to connection: %v", err)
				break
			}
		}
		
		if err := scanner.Err(); err != nil {
			log.Printf("error reading from command stdout: %v", err)
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		log.Printf("command exited with error: %v", err)
	} else {
		log.Printf("command completed successfully")
	}

	log.Printf("connection from %s closed", conn.RemoteAddr())
}
