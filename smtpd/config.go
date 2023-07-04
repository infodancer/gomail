package smtpd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/infodancer/gomail/connect"
	"github.com/infodancer/gomail/queue"
)

type Config struct {
	ServerName string `json:"servername"`
	Banner     string `json:"banner"`
	Spamc      string `json:"spamc"`
	Maxsize    int64  `json:"maxsize"`
	MQueue     *queue.Queue
}

func ReadConfigFile(file string) (*Config, error) {
	c := Config{}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading smtpd config from %s: %w", file, err)
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling smtpd config from %s: %w", file, err)
	}
	return &c, nil
}

// Connection accepts a connection and sends the configured banner
func (cfg *Config) Start(c connect.TCPConnection) (*Session, error) {
	s := &Session{
		Conn: c,
	}
	err := s.SendCodeLine(220, cfg.ServerName+" "+cfg.Banner)
	if err != nil {
		return nil, err
	}
	return s, nil
}
