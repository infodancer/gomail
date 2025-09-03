package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Listener contains configuration for a TCP listener
type Listener struct {
	// IPAddress to listen on, empty means all interfaces
	IPAddress string `toml:"ip_address"`
	// Port to listen on
	Port int `toml:"port"`
	// MaxConnections is the maximum number of concurrent connections
	MaxConnections int `toml:"max_connections"`
	// Timeout in seconds for idle connections
	IdleTimeout int `toml:"idle_timeout"`
	// Command is the command to execute for each connection
	Command string `toml:"command"`
	// Args are the arguments to pass to the command
	Args []string `toml:"args"`
}

// SecureConnection contains TLS/SSL configuration
type SecureConnection struct {
	// Enabled indicates whether TLS/SSL is enabled
	Enabled bool `toml:"enabled"`
	// CertFile is the path to the certificate file
	CertFile string `toml:"cert_file"`
	// KeyFile is the path to the private key file
	KeyFile string `toml:"key_file"`
	// RequireClientCert indicates whether client certificates are required
	RequireClientCert bool `toml:"require_client_cert"`
	// MinTLSVersion is the minimum TLS version to support (e.g., "1.2")
	MinTLSVersion string `toml:"min_tls_version"`
}

// ServerConfig contains common server configuration
type ServerConfig struct {
	// ServerName is the name of the server
	ServerName string `toml:"server_name"`
	// Listener contains the TCP listener configuration
	Listener Listener `toml:"listener"`
	// TLS contains the TLS configuration
	TLS SecureConnection `toml:"tls"`
}

// LoadTOMLConfig loads configuration from a TOML file into the provided config struct
func LoadTOMLConfig(filePath string, config interface{}) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", filePath)
	}

	// Read and parse the TOML file
	_, err := toml.DecodeFile(filePath, config)
	if err != nil {
		return fmt.Errorf("error parsing TOML configuration from %s: %w", filePath, err)
	}

	return nil
}
