package main

import (
	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/pop3d"
)

type Config struct {
	ServerName string
}

// Start sends the banner for new connections
func (cfg *Config) Start(c *connect.TCPConnection) (*pop3d.Session, error) {
	s := pop3d.Session{}
	err := s.SendLine("+OK " + cfg.ServerName + " POP3 server ready")
	if err != nil {
		return nil, err
	}
	return &s, nil
}
