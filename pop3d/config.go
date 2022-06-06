package pop3d

import (
	"github.com/infodancer/gomail/connect"
)

type Config struct {
	ServerName string
}

// Start sends the banner for new connections
func (cfg *Config) Start(c *connect.TCPConnection) (*Session, error) {
	s := Session{}
	err := s.SendLine("+OK " + cfg.ServerName + " POP3 server ready")
	if err != nil {
		return nil, err
	}
	return &s, nil
}
