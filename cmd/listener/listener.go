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
	// Command is the command to execute for each connection
	Command string `toml:"command"`
	// Args are the arguments to pass to the command
	Args []string `toml:"args"`
	// Server contains the server configuration
	Server config.ServerConfig `toml:"server"`
}

func main() {
	cfgfile := flag.String("cfg", "/opt/infodancer/gomail/etc/listener.toml", "The configuration file")
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if versionFlag != nil && *versionFlag {
		log.Println("Version: " + Version)
		os.Exit(0)
	}

	var cfg ListenerConfig
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
	address := fmt.Sprintf("%s:%d", cfg.Server.Listener.IPAddress, cfg.Server.Listener.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("error starting listener on %s: %v", address, err)
		os.Exit(1)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("error closing listener: %v", err)
		}
	}()

	log.Printf("listening on %s, running command: %s", address, cfg.Command)

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
		if cfg.Server.Listener.MaxConnections > 0 && connectionCount >= cfg.Server.Listener.MaxConnections {
			log.Printf("maximum connections (%d) reached, rejecting connection", cfg.Server.Listener.MaxConnections)
			if err := conn.Close(); err != nil {
				log.Printf("error closing rejected connection: %v", err)
			}
			continue
		}

		connectionCount++
		wg.Add(1)

		go func(c net.Conn) {
			defer wg.Done()
			defer func() {
				connectionCount--
				if err := c.Close(); err != nil {
					log.Printf("error closing connection: %v", err)
				}
			}()

			handleConnection(c, cfg)
		}(conn)
	}
}

func handleConnection(conn net.Conn, cfg ListenerConfig) {
	log.Printf("handling connection from %s", conn.RemoteAddr())

	// Start the configured command
	cmd := exec.Command(cfg.Command, cfg.Args...)
	
	// Get pipes for stdin/stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("error creating stdin pipe: %v", err)
		return
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("error creating stdout pipe: %v", err)
		if closeErr := stdin.Close(); closeErr != nil {
			log.Printf("error closing stdin pipe: %v", closeErr)
		}
		return
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		log.Printf("error starting command: %v", err)
		if closeErr := stdin.Close(); closeErr != nil {
			log.Printf("error closing stdin pipe: %v", closeErr)
		}
		if closeErr := stdout.Close(); closeErr != nil {
			log.Printf("error closing stdout pipe: %v", closeErr)
		}
		return
	}

	// Create a wait group for the goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Copy from network connection to command stdin
	go func() {
		defer wg.Done()
		defer func() {
			if err := stdin.Close(); err != nil {
				log.Printf("error closing stdin: %v", err)
			}
		}()
		
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
		defer func() {
			if err := conn.Close(); err != nil {
				log.Printf("error closing connection: %v", err)
			}
		}()
		
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
