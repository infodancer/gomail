package pop3d

import (
	"github.com/infodancer/gomail/config"
	"github.com/infodancer/gomail/connect"
)

type Config struct {
	// Embed the common server configuration
	config.ServerConfig `toml:"server"`
	// POP3-specific configuration
	Banner string `toml:"banner"`
}

// Start sends the banner for new connections
func (cfg *Config) Start(c connect.TCPConnection) (*Session, error) {
	s := Session{
		Config: *cfg,
		Conn:   c,
	}
	banner := cfg.Banner
	if banner == "" {
		banner = cfg.ServerName + " POP3 server ready"
	}
	err := s.SendLine("+OK " + banner)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
