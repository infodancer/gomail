package smtpd

import (
	"github.com/infodancer/gomail/config"
	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/queue"
)

type Config struct {
	// Embed the common server configuration
	config.ServerConfig `toml:"server"`
	// SMTP-specific configuration
	Banner  string `toml:"banner"`
	Spamc   string `toml:"spamc"`
	Maxsize int64  `toml:"maxsize"`
	MQueue  *queue.Queue
}

// Start accepts a connection and sends the configured banner
func (cfg *Config) Start(c connect.TCPConnection) (*Session, error) {
	s := &Session{
		Config: *cfg,
		Conn:   c,
	}
	err := s.SendCodeLine(220, cfg.ServerName+" "+cfg.Banner)
	if err != nil {
		return nil, err
	}
	return s, nil
}
