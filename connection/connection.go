package connection

import (
	"bufio"
	"os"
)

// This type holds information about a tcp connection
type Connection interface {
	ReadLine() (string, error)
	WriteLine(s string) error
	Close() error

	GetProto() string
	GetTCPLocalIP() string
	GetTCPLocalPort() string
	GetTCPLocalHost() string
	GetTCPRemotePort() string
	GetTCPRemoteIP() string
	GetTCPRemoteHost() string
}

// StandardIOConnection expects stdin, stdout, and TCP info in the environment
type StandardIOConnection struct {
	rw *bufio.ReadWriter
}

// Close currently just flushes the buffers...
func (c StandardIOConnection) Close() error {
	c.rw.Flush()
	return nil
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
