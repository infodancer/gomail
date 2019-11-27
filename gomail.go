package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"

	"infodancer.org/gomail/smtpd"
)

// Connection represents a network connection from a client
type Connection interface {
	Close()
}

// ClientManager manages Clients
type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// Client is a network connection from a client
type Client struct {
	port       int
	clientType int
	socket     net.Conn
	reader     bufio.Reader
	data       chan []byte
	session    smtpd.Session
}

const (
	smtp = iota
	pop3 = iota
	imap = iota
)

// This entry point is for Docker containers; it starts everything else with minimal arguments
func main() {
	test := flag.Bool("test", false, "Start in test mode (add 1000 to all port numbers)")
	confdir := flag.String("conf", "/srv/gomail", "The directory containing configuration and runtime data (must be persistent); defaults to /srv/gomail")
	flag.Parse()
	if !*test {
		// Start SMTP listener on port 25 (incoming mail)
		startSMTPListener(25, false, false)
		// Start SMTP listener on port 587 (clients sending authenticated mail only)
		startSMTPListener(587, false, true)
		// Start SMTP listener on port 587 (clients sending authenticated mail over ssl only)
		startSMTPListener(465, true, true)
		// Start SMTP listener on port ??? (SSL only)
		// startSMTPListener(587, true, false)
		// Start POP3 listener
		startPOP3Listener(110, false)
		startPOP3Listener(995, true)
		// Start IMAP listener
		startIMAPListener(143, false)
		startIMAPListener(993, true)
	} else {
		// Start SMTP listener on port 25 (incoming mail)
		go startSMTPListener(1025, false, false)
		// Start SMTP listener on port 587 (clients sending authenticated mail only)
		go startSMTPListener(1587, false, true)
		// Start SMTP listener on port 587 (clients sending authenticated mail over ssl only)
		go startSMTPListener(1465, true, true)
		// Start SMTP listener on port ??? (SSL only)
		// go startSMTPListener(1587, true, false)
		// Start POP3 listener
		startPOP3Listener(1110, false)
		startPOP3Listener(1995, true)
		// Start IMAP listener
		startIMAPListener(1143, false)
		startIMAPListener(1993, true)
	}
	// Start outbound mail processing
	for {
		// Loop endlessly to process mail until interrupted
		time.Sleep(10 * time.Second)
	}
}

// Close will close the connection to a client
func (client *Client) Close() {
	manager.unregister <- client
	client.socket.Close()
}

// Close will close the connection to a client
func (client *Client) String() {
	return fmt.Printf("Connection type %v with client from %v to port %v", client.ClientType, client.socket.RemoteAddr, client.socket.LocalAddr)
}

func (manager *ClientManager) start() {
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
			fmt.Printf("Added new connection to %v speaking %v\n", connection.socket.LocalAddr(), connection.clientType)
		case connection := <-manager.unregister:
			if _, ok := manager.clients[connection]; ok {
				close(connection.data)
				delete(manager.clients, connection)
				fmt.Println("A connection has terminated!")
			}
		case message := <-manager.broadcast:
			for connection := range manager.clients {
				select {
				case connection.data <- message:
				default:
					close(connection.data)
					delete(manager.clients, connection)
				}
			}
		}
	}
}

func (manager *ClientManager) receive(client *Client) {
	for {
		line, isPrefix, err := client.reader.ReadLine()
		// Lines that exceed our buffer are considered errors
		if err != nil || isPrefix {
			manager.unregister <- client
			client.socket.Close()
			break
		}
		if len(line) > 0 {
			fmt.Println("RECEIVED: " + string(line))
			// Handle the line here
			smtpd.HandleInputLine(client.session, line)
		}
	}
}

func (manager *ClientManager) send(client *Client) {
	defer client.socket.Close()
	for {
		select {
		case message, ok := <-client.data:
			if !ok {
				return
			}
			client.socket.Write(message)
		}
	}
}

func startIMAPListener(port int, ssl bool) error {
	fmt.Println("IMAP protocol listener not yet implemented")
}

func startPOP3Listener(port int, ssl bool) error {
	fmt.Println("POP3 protocol listener not yet implemented")
}

func startSMTPListener(port int, sslRequired bool, authRequired bool) error {
	sport := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", sport)
	if err != nil {
		fmt.Println("Can't listen on port:", port)
		return err
	}
	fmt.Println("Listening on port:", port)

	manager := ClientManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	go manager.start()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection", port)
		}
		client := &Client{socket: connection, clientType: smtp, port: port, data: make(chan []byte)}
		manager.register <- client
		go manager.receive(client)
		go manager.send(client)
	}
}
