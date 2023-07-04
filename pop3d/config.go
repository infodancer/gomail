package pop3d

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/infodancer/gomail/connect"
)

type Config struct {
	ServerName string `json:"servername"`
}

func ReadConfigFile(file string) (*Config, error) {
	c := Config{}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading pop3d config from %s: %w", file, err)
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling pop3d config from %s: %w", file, err)
	}
	return &c, nil
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
