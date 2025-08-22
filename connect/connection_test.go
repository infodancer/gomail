package connect

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStandardIOConnection(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err, "should create StandardIOConnection without error")
	assert.NotNil(t, conn, "connection should not be nil")
	assert.NotNil(t, conn.Logger(), "logger should not be nil")
}

func TestStandardIOConnection_IsEncrypted(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	// Currently stubbed to return true
	assert.True(t, conn.IsEncrypted(), "IsEncrypted should return true")
}

func TestStandardIOConnection_GetProto(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	// Test with environment variable set
	os.Setenv("PROTO", "tcp")
	defer os.Unsetenv("PROTO")
	
	assert.Equal(t, "tcp", conn.GetProto(), "should return PROTO environment variable")
}

func TestStandardIOConnection_GetTCPLocalIP(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testIP := "127.0.0.1"
	os.Setenv("TCPLOCALIP", testIP)
	defer os.Unsetenv("TCPLOCALIP")
	
	assert.Equal(t, testIP, conn.GetTCPLocalIP(), "should return TCPLOCALIP environment variable")
}

func TestStandardIOConnection_GetTCPLocalPort(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testPort := "25"
	os.Setenv("TCPLOCALPORT", testPort)
	defer os.Unsetenv("TCPLOCALPORT")
	
	assert.Equal(t, testPort, conn.GetTCPLocalPort(), "should return TCPLOCALPORT environment variable")
}

func TestStandardIOConnection_GetTCPLocalHost(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testHost := "localhost"
	os.Setenv("TCOLOCALHOST", testHost)
	defer os.Unsetenv("TCOLOCALHOST")
	
	assert.Equal(t, testHost, conn.GetTCPLocalHost(), "should return TCOLOCALHOST environment variable")
}

func TestStandardIOConnection_GetTCPRemotePort(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testPort := "12345"
	os.Setenv("TCPREMOTEPORT", testPort)
	defer os.Unsetenv("TCPREMOTEPORT")
	
	assert.Equal(t, testPort, conn.GetTCPRemotePort(), "should return TCPREMOTEPORT environment variable")
}

func TestStandardIOConnection_GetTCPRemoteIP(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testIP := "192.168.1.100"
	os.Setenv("TCPREMOTEIP", testIP)
	defer os.Unsetenv("TCPREMOTEIP")
	
	assert.Equal(t, testIP, conn.GetTCPRemoteIP(), "should return TCPREMOTEIP environment variable")
}

func TestStandardIOConnection_GetTCPRemoteHost(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testHost := "client.example.com"
	os.Setenv("TCPREMOTEHOST", testHost)
	defer os.Unsetenv("TCPREMOTEHOST")
	
	assert.Equal(t, testHost, conn.GetTCPRemoteHost(), "should return TCPREMOTEHOST environment variable")
}

func TestStandardIOConnection_Close(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	err = conn.Close()
	assert.NoError(t, err, "Close should not return an error")
}

func TestStandardIOConnection_EnvironmentVariables_Empty(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	// Clear all relevant environment variables
	envVars := []string{"PROTO", "TCPLOCALIP", "TCPLOCALPORT", "TCOLOCALHOST", 
		"TCPREMOTEPORT", "TCPREMOTEIP", "TCPREMOTEHOST"}
	
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
	
	// All should return empty strings when environment variables are not set
	assert.Equal(t, "", conn.GetProto(), "GetProto should return empty string when PROTO not set")
	assert.Equal(t, "", conn.GetTCPLocalIP(), "GetTCPLocalIP should return empty string when TCPLOCALIP not set")
	assert.Equal(t, "", conn.GetTCPLocalPort(), "GetTCPLocalPort should return empty string when TCPLOCALPORT not set")
	assert.Equal(t, "", conn.GetTCPLocalHost(), "GetTCPLocalHost should return empty string when TCOLOCALHOST not set")
	assert.Equal(t, "", conn.GetTCPRemotePort(), "GetTCPRemotePort should return empty string when TCPREMOTEPORT not set")
	assert.Equal(t, "", conn.GetTCPRemoteIP(), "GetTCPRemoteIP should return empty string when TCPREMOTEIP not set")
	assert.Equal(t, "", conn.GetTCPRemoteHost(), "GetTCPRemoteHost should return empty string when TCPREMOTEHOST not set")
}
