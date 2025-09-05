package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/infodancer/gomail/config"
)

var Version string

// isConnectionClosed checks if an error is due to a closed network connection
func isConnectionClosed(err error) bool {
	return err != nil && strings.Contains(err.Error(), "use of closed network connection")
}

type GenericConfig struct {
	// Command is the command to execute for each connection (optional, for listener configs)
	Command string `toml:"command"`
	// Args are the arguments to pass to the command (optional, for listener configs)
	Args []string `toml:"args"`
	// Server contains the server configuration
	Server config.ServerConfig `toml:"server"`
	// Legacy fields for configs that don't use nested server structure
	ServerName string                  `toml:"server_name"`
	Listener   config.Listener         `toml:"listener"`
	TLS        config.SecureConnection `toml:"tls"`
}

func main() {
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if versionFlag != nil && *versionFlag {
		log.Println("Version: " + Version)
		os.Exit(0)
	}

	configFiles := flag.Args()
	if len(configFiles) == 0 {
		log.Printf("error: no configuration files specified")
		log.Printf("usage: %s [options] config1.toml [config2.toml ...]", os.Args[0])
		os.Exit(1)
	}

	var wg sync.WaitGroup

	// Start a listener for each configuration file
	for _, cfgfile := range configFiles {
		wg.Add(1)
		go func(configFile string) {
			defer wg.Done()
			startListener(configFile)
		}(cfgfile)
	}

	// Wait for all listeners to finish
	wg.Wait()
}

func startListener(cfgfile string) {
	var cfg GenericConfig
	err := config.LoadTOMLConfig(cfgfile, &cfg)
	if err != nil {
		log.Printf("error reading configuration from %s: %v", cfgfile, err)
		return
	}

	// Normalize configuration - handle both nested and legacy formats
	var serverConfig config.ServerConfig
	if cfg.Server.ServerName != "" {
		// Use nested server configuration
		serverConfig = cfg.Server
	}

	var command string
	var args []string
	if cfg.Command != "" {
		// Explicit command specified (top-level listener config)
		command = cfg.Command
		args = cfg.Args
	} else if serverConfig.Listener.Command != "" {
		// Command specified in nested listener configuration
		command = serverConfig.Listener.Command
		args = serverConfig.Listener.Args
	}

	// Require a command to be configured
	if command == "" {
		log.Printf("error: no command configured in %s", cfgfile)
		return
	}

	// Start listening
	address := fmt.Sprintf("%s:%d", serverConfig.Listener.IPAddress, serverConfig.Listener.Port)
	var listener net.Listener
	
	if serverConfig.TLS.Enabled {
		// Load TLS certificate and key
		cert, err := tls.LoadX509KeyPair(serverConfig.TLS.CertFile, serverConfig.TLS.KeyFile)
		if err != nil {
			log.Printf("error loading TLS certificate for %s: %v", cfgfile, err)
			return
		}
		
		// Configure TLS
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		
		// Set minimum TLS version if specified
		switch serverConfig.TLS.MinTLSVersion {
		case "1.0":
			tlsConfig.MinVersion = tls.VersionTLS10
		case "1.1":
			tlsConfig.MinVersion = tls.VersionTLS11
		case "1.2":
			tlsConfig.MinVersion = tls.VersionTLS12
		case "1.3":
			tlsConfig.MinVersion = tls.VersionTLS13
		default:
			tlsConfig.MinVersion = tls.VersionTLS12 // Default to TLS 1.2
		}
		
		// Set client certificate requirements
		if serverConfig.TLS.RequireClientCert {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
		
		listener, err = tls.Listen("tcp", address, tlsConfig)
		if err != nil {
			log.Printf("error starting TLS listener on %s for config %s: %v", address, cfgfile, err)
			return
		}
		log.Printf("listening on %s with TLS (config: %s), running command: %s %v", address, cfgfile, command, args)
	} else {
		listener, err = net.Listen("tcp", address)
		if err != nil {
			log.Printf("error starting listener on %s for config %s: %v", address, cfgfile, err)
			return
		}
		log.Printf("listening on %s (config: %s), running command: %s %v", address, cfgfile, command, args)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("error closing listener for %s: %v", cfgfile, err)
		}
	}()


	// Handle connections
	var connWg sync.WaitGroup
	connectionCount := 0
	var connCountMutex sync.Mutex

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection on %s: %v", address, err)
			continue
		}

		// Check max connections limit
		connCountMutex.Lock()
		if serverConfig.Listener.MaxConnections > 0 && connectionCount >= serverConfig.Listener.MaxConnections {
			log.Printf("maximum connections (%d) reached for %s, rejecting connection", serverConfig.Listener.MaxConnections, address)
			if err := conn.Close(); err != nil {
				log.Printf("error closing rejected connection: %v", err)
			}
			connCountMutex.Unlock()
			continue
		}
		connectionCount++
		connCountMutex.Unlock()

		// Output connection info to stdout
		localAddr := conn.LocalAddr().(*net.TCPAddr)
		remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
		maxConns := serverConfig.Listener.MaxConnections
		if maxConns == 0 {
			maxConns = -1 // Indicate unlimited
		}
		if maxConns > 0 {
			fmt.Printf("Connection accepted: server=%s local_port=%d remote_ip=%s remote_port=%d [%d/%d]\n",
				serverConfig.ServerName, localAddr.Port, remoteAddr.IP.String(), remoteAddr.Port, connectionCount, maxConns)
		} else {
			fmt.Printf("Connection accepted: server=%s local_port=%d remote_ip=%s remote_port=%d [%d/unlimited]\n",
				serverConfig.ServerName, localAddr.Port, remoteAddr.IP.String(), remoteAddr.Port, connectionCount)
		}

		connWg.Add(1)

		go func(c net.Conn) {
			defer connWg.Done()
			defer func() {
				connCountMutex.Lock()
				connectionCount--
				connCountMutex.Unlock()
				if err := c.Close(); err != nil && !isConnectionClosed(err) {
					log.Printf("error closing connection: %v", err)
				}
			}()

			handleConnection(c, command, args, serverConfig.Listener.IdleTimeout)
		}(conn)
	}
}

func handleConnection(conn net.Conn, command string, args []string, idleTimeoutSeconds int) {
	log.Printf("handling connection from %s", conn.RemoteAddr())

	// Set up idle timeout if configured
	var idleTimeout time.Duration
	if idleTimeoutSeconds > 0 {
		idleTimeout = time.Duration(idleTimeoutSeconds) * time.Second
		if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
			log.Printf("error setting initial connection deadline: %v", err)
		}
	}

	// Start the configured command
	cmd := exec.Command(command, args...)

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

	// Create channels to signal when goroutines should stop
	stopReading := make(chan struct{})
	stopWriting := make(chan struct{})
	
	// Create a wait group for the goroutines
	var wg sync.WaitGroup
	wg.Add(3) // Add one more for the command monitor

	// Monitor the command and signal when it exits
	go func() {
		defer wg.Done()
		err := cmd.Wait()
		if err != nil {
			log.Printf("command exited with error: %v", err)
		} else {
			log.Printf("command completed successfully")
		}
		// Close the connection immediately when command exits
		if closeErr := conn.Close(); closeErr != nil && !isConnectionClosed(closeErr) {
			log.Printf("error closing connection after command exit: %v", closeErr)
		}
		// Signal both goroutines to stop
		close(stopReading)
		close(stopWriting)
	}()

	// Copy from network connection to command stdin
	go func() {
		defer wg.Done()
		defer func() {
			if err := stdin.Close(); err != nil && !isConnectionClosed(err) {
				log.Printf("error closing stdin: %v", err)
			}
		}()

		reader := bufio.NewReader(conn)
		for {
			select {
			case <-stopReading:
				return
			default:
				// Update deadline before each read if timeout is configured
				if idleTimeoutSeconds > 0 {
					if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
						log.Printf("error updating connection deadline: %v", err)
					}
				}

				line, err := reader.ReadString('\n')
				if err != nil {
					if err != io.EOF && !isConnectionClosed(err) {
						// Check if this is a timeout error
						if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
							log.Printf("connection from %s timed out after %d seconds", conn.RemoteAddr(), idleTimeoutSeconds)
						} else {
							log.Printf("error reading from connection: %v", err)
						}
					}
					return
				}

				_, err = stdin.Write([]byte(line))
				if err != nil {
					if !isConnectionClosed(err) {
						log.Printf("error writing to command stdin: %v", err)
					}
					return
				}
			}
		}
	}()

	// Copy from command stdout to network connection
	go func() {
		defer wg.Done()

		scanner := bufio.NewScanner(stdout)
		for {
			select {
			case <-stopWriting:
				return
			default:
				if !scanner.Scan() {
					if err := scanner.Err(); err != nil {
						log.Printf("error reading from command stdout: %v", err)
					}
					return
				}
				line := scanner.Text() + "\n"

				// Update deadline before each write if timeout is configured
				if idleTimeoutSeconds > 0 {
					if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
						log.Printf("error updating connection deadline: %v", err)
					}
				}

				_, err := conn.Write([]byte(line))
				if err != nil {
					if !isConnectionClosed(err) {
						log.Printf("error writing to connection: %v", err)
					}
					return
				}
			}
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	log.Printf("connection from %s closed", conn.RemoteAddr())
}
