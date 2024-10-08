package connect

import (
	"bufio"
	"log"
	"os"
)

// TCPConnection holds information about a tcp connection
type TCPConnection interface {
	ReadLine() (string, error)
	WriteLine(string) error
	Close() error
	GetProto() string
	GetTCPLocalIP() string
	GetTCPLocalPort() string
	GetTCPLocalHost() string
	GetTCPRemotePort() string
	GetTCPRemoteIP() string
	GetTCPRemoteHost() string
	// IsSecure returns true if the connection is encrypted
	IsEncrypted() bool
	Logger() *log.Logger
}

// StandardIOConnection expects stdin, stdout, and TCP info in the environment
type StandardIOConnection struct {
	rw     *bufio.ReadWriter
	logger *log.Logger
}

func NewStandardIOConnection() (TCPConnection, error) {
	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)

	logger := log.New(os.Stderr, "", 1|2|6)
	stdcon := StandardIOConnection{
		rw:     bufio.NewReadWriter(r, w),
		logger: logger,
	}
	return &stdcon, nil
}

// Close currently just flushes the buffers...
func (c *StandardIOConnection) Close() error {
	c.rw.Flush()
	return nil
}

// Logger returns a pointer to the logger for this connection
// Usually this logs to stderr
func (c *StandardIOConnection) Logger() *log.Logger {
	return c.logger
}

// IsEncrypted indicates whethe the connection is encrypted (but not authenticated)
// Stubbed for now
func (c *StandardIOConnection) IsEncrypted() bool {
	return true
}

func (c *StandardIOConnection) ReadLine() (string, error) {
	s, err := c.rw.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Many early protocols specify CRLF, so for convenience...
	if s[len(s)-1] == '\r' {
		return s[0 : len(s)-2], nil
	}
	return s, nil
}

// WriteLine automatically appends a linefeed character
func (c *StandardIOConnection) WriteLine(s string) error {
	_, err := c.rw.WriteString(s + "\n")
	if err != nil {
		return err
	}
	err = c.rw.Flush()
	return err
}

func (c *StandardIOConnection) GetProto() string {
	return os.Getenv("PROTO")
}

func (c *StandardIOConnection) GetTCPLocalIP() string {
	return os.Getenv("TCPLOCALIP")
}

func (c *StandardIOConnection) GetTCPLocalPort() string {
	return os.Getenv("TCPLOCALPORT")
}

func (c *StandardIOConnection) GetTCPLocalHost() string {
	return os.Getenv("TCOLOCALHOST")
}

func (c *StandardIOConnection) GetTCPRemotePort() string {
	return os.Getenv("TCPREMOTEPORT")
}

func (c *StandardIOConnection) GetTCPRemoteIP() string {
	return os.Getenv("TCPREMOTEIP")
}

func (c *StandardIOConnection) GetTCPRemoteHost() string {
	return os.Getenv("TCPREMOTEHOST")
}

func (c *StandardIOConnection) GetTCPRemoteInfo() string {
	return os.Getenv("TCPREMOTEINFO")
}
