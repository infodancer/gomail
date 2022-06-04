package main

import (
	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/queue"
)

type Config struct {
	ServerName string
	Banner     string
	Spamc      string
	maxsize    int64
	MQueue     *queue.Queue
}

// Connection accepts a connection and sends the configured banner
func (cfg *Config) Start(c *connect.TCPConnection) (*Session, error) {
	s := Session{}
	err := s.SendCodeLine(220, cfg.ServerName+" "+cfg.Banner)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
