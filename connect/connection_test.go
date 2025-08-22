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
	err = os.Setenv("PROTO", "tcp")
	require.NoError(t, err, "should set PROTO environment variable")
	defer func() {
		err := os.Unsetenv("PROTO")
		require.NoError(t, err, "should unset PROTO environment variable")
	}()
	
	assert.Equal(t, "tcp", conn.GetProto(), "should return PROTO environment variable")
}

func TestStandardIOConnection_GetTCPLocalIP(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testIP := "127.0.0.1"
	err = os.Setenv("TCPLOCALIP", testIP)
	require.NoError(t, err, "should set TCPLOCALIP environment variable")
	defer func() {
		err := os.Unsetenv("TCPLOCALIP")
		require.NoError(t, err, "should unset TCPLOCALIP environment variable")
	}()
	
	assert.Equal(t, testIP, conn.GetTCPLocalIP(), "should return TCPLOCALIP environment variable")
}

func TestStandardIOConnection_GetTCPLocalPort(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testPort := "25"
	err = os.Setenv("TCPLOCALPORT", testPort)
	require.NoError(t, err, "should set TCPLOCALPORT environment variable")
	defer func() {
		err := os.Unsetenv("TCPLOCALPORT")
		require.NoError(t, err, "should unset TCPLOCALPORT environment variable")
	}()
	
	assert.Equal(t, testPort, conn.GetTCPLocalPort(), "should return TCPLOCALPORT environment variable")
}

func TestStandardIOConnection_GetTCPLocalHost(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testHost := "localhost"
	err = os.Setenv("TCOLOCALHOST", testHost)
	require.NoError(t, err, "should set TCOLOCALHOST environment variable")
	defer func() {
		err := os.Unsetenv("TCOLOCALHOST")
		require.NoError(t, err, "should unset TCOLOCALHOST environment variable")
	}()
	
	assert.Equal(t, testHost, conn.GetTCPLocalHost(), "should return TCOLOCALHOST environment variable")
}

func TestStandardIOConnection_GetTCPRemotePort(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testPort := "12345"
	err = os.Setenv("TCPREMOTEPORT", testPort)
	require.NoError(t, err, "should set TCPREMOTEPORT environment variable")
	defer func() {
		err := os.Unsetenv("TCPREMOTEPORT")
		require.NoError(t, err, "should unset TCPREMOTEPORT environment variable")
	}()
	
	assert.Equal(t, testPort, conn.GetTCPRemotePort(), "should return TCPREMOTEPORT environment variable")
}

func TestStandardIOConnection_GetTCPRemoteIP(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testIP := "192.168.1.100"
	err = os.Setenv("TCPREMOTEIP", testIP)
	require.NoError(t, err, "should set TCPREMOTEIP environment variable")
	defer func() {
		err := os.Unsetenv("TCPREMOTEIP")
		require.NoError(t, err, "should unset TCPREMOTEIP environment variable")
	}()
	
	assert.Equal(t, testIP, conn.GetTCPRemoteIP(), "should return TCPREMOTEIP environment variable")
}

func TestStandardIOConnection_GetTCPRemoteHost(t *testing.T) {
	conn, err := NewStandardIOConnection()
	require.NoError(t, err)
	
	testHost := "client.example.com"
	err = os.Setenv("TCPREMOTEHOST", testHost)
	require.NoError(t, err, "should set TCPREMOTEHOST environment variable")
	defer func() {
		err := os.Unsetenv("TCPREMOTEHOST")
		require.NoError(t, err, "should unset TCPREMOTEHOST environment variable")
	}()
	
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
		err := os.Unsetenv(envVar)
		require.NoError(t, err, "should unset %s environment variable", envVar)
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
