package connect

import (
	"bufio"
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
}

// StandardIOConnection expects stdin, stdout, and TCP info in the environment
type StandardIOConnection struct {
	rw *bufio.ReadWriter
}

func NewStandardIOConnection(r *bufio.Reader, w *bufio.Writer) (TCPConnection, error) {
	stdcon := StandardIOConnection{
		rw: bufio.NewReadWriter(r, w),
	}
	return stdcon, nil
}

// Close currently just flushes the buffers...
func (c StandardIOConnection) Close() error {
	c.rw.Flush()
	return nil
}

// IsEncrypted indicates whethe the connection is encrypted (but not authenticated)
// Stubbed for now
func (c StandardIOConnection) IsEncrypted() bool {
	return true
}

func (c StandardIOConnection) ReadLine() (string, error) {
	s, err := c.rw.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Many early protocols specify CRLF, so for convenience...
	if s[len(s)] == '\r' {
		return s[0 : len(s)-1], nil
	}
	return s, nil
}

// WriteLine automatically appends a linefeed character
func (c StandardIOConnection) WriteLine(s string) error {
	_, err := c.rw.WriteString(s + `\n`)
	return err
}

func (c StandardIOConnection) GetProto() string {
	return os.Getenv("PROTO")
}

func (c StandardIOConnection) GetTCPLocalIP() string {
	return os.Getenv("TCPLOCALIP")
}

func (c StandardIOConnection) GetTCPLocalPort() string {
	return os.Getenv("TCPLOCALPORT")
}

func (c StandardIOConnection) GetTCPLocalHost() string {
	return os.Getenv("TCOLOCALHOST")
}

func (c StandardIOConnection) GetTCPRemotePort() string {
	return os.Getenv("TCPREMOTEPORT")
}

func (c StandardIOConnection) GetTCPRemoteIP() string {
	return os.Getenv("TCPREMOTEIP")
}

func (c StandardIOConnection) GetTCPRemoteHost() string {
	return os.Getenv("TCPREMOTEHOST")
}

func (c StandardIOConnection) GetTCPRemoteInfo() string {
	return os.Getenv("TCPREMOTEINFO")
}
