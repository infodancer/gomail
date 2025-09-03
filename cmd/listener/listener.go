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
	} else {
		// Use legacy flat configuration
		serverConfig.ServerName = cfg.ServerName
		serverConfig.Listener = cfg.Listener
		serverConfig.TLS = cfg.TLS
	}

	// Determine command to run
	var command string
	var args []string
	if cfg.Command != "" {
		// Explicit command specified (listener config)
		command = cfg.Command
		args = cfg.Args
	}

	// Start listening
	address := fmt.Sprintf("%s:%d", serverConfig.Listener.IPAddress, serverConfig.Listener.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("error starting listener on %s for config %s: %v", address, cfgfile, err)
		return
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("error closing listener for %s: %v", cfgfile, err)
		}
	}()

	log.Printf("listening on %s (config: %s), running command: %s", address, cfgfile, command)

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
		localAddr := c.LocalAddr().(*net.TCPAddr)
		remoteAddr := c.RemoteAddr().(*net.TCPAddr)
		fmt.Printf("Connection accepted: server=%s local_port=%d remote_ip=%s remote_port=%d\n",
			serverConfig.ServerName, localAddr.Port, remoteAddr.IP.String(), remoteAddr.Port)

		connWg.Add(1)

		go func(c net.Conn) {
			defer connWg.Done()
			defer func() {
				connCountMutex.Lock()
				connectionCount--
				connCountMutex.Unlock()
				if err := c.Close(); err != nil {
					log.Printf("error closing connection: %v", err)
				}
			}()

			handleConnection(c, command, args)
		}(conn)
	}
}

func handleConnection(conn net.Conn, command string, args []string) {
	log.Printf("handling connection from %s", conn.RemoteAddr())

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
