package smtpd

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// MockConnection implements connect.TCPConnection for testing
type MockConnection struct {
	readLines  []string
	writeLines []string
	readIndex  int
}

func (m *MockConnection) ReadLine() (string, error) {
	if m.readIndex >= len(m.readLines) {
		return "", nil
	}
	line := m.readLines[m.readIndex]
	m.readIndex++
	return line, nil
}

func (m *MockConnection) WriteLine(s string) error {
	m.writeLines = append(m.writeLines, s)
	return nil
}

func (m *MockConnection) Close() error {
	return nil
}

func (m *MockConnection) GetProto() string {
	return "tcp"
}

func (m *MockConnection) GetTCPLocalIP() string {
	return "127.0.0.1"
}

func (m *MockConnection) GetTCPLocalPort() string {
	return "25"
}

func (m *MockConnection) GetTCPLocalHost() string {
	return "localhost"
}

func (m *MockConnection) GetTCPRemotePort() string {
	return "12345"
}

func (m *MockConnection) GetTCPRemoteIP() string {
	return "192.168.1.100"
}

func (m *MockConnection) GetTCPRemoteHost() string {
	return "client.example.com"
}

func (m *MockConnection) Logger() interface{} {
	return nil
}

func createTestSession() *Session {
	mockConn := &MockConnection{}
	config := Config{
		Spamc: "",
	}
	session := Create(config, mockConn)
	session.Data = "Subject: Test Message\n\nThis is a test message body."
	return session
}

func TestCheckSpam_NoSpamcConfigured(t *testing.T) {
	session := createTestSession()
	session.Config.Spamc = ""

	result, err := session.checkSpam()

	if err != nil {
		t.Errorf("Expected no error when spamc not configured, got: %v", err)
	}

	if result != session.Data {
		t.Errorf("Expected original data to be returned when spamc not configured")
	}
}

func TestCheckSpam_WithSpamcConfigured(t *testing.T) {
	// Skip this test if we can't find a suitable mock command
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("cat command not available for testing")
	}

	session := createTestSession()
	session.Config.Spamc = "cat" // Use cat as a mock spamc that just echoes input

	result, err := session.checkSpam()

	if err != nil {
		t.Errorf("Expected no error with valid spamc command, got: %v", err)
	}

	// cat should return the same content with newlines
	expectedLines := strings.Split(session.Data, "\n")
	resultLines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")

	if len(resultLines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d lines", len(expectedLines), len(resultLines))
	}

	for i, expectedLine := range expectedLines {
		if i < len(resultLines) && resultLines[i] != expectedLine {
			t.Errorf("Line %d: expected %q, got %q", i, expectedLine, resultLines[i])
		}
	}
}

func TestCheckSpam_WithSpamcAddingHeaders(t *testing.T) {
	// Create a temporary script that adds spam headers
	tmpScript := `#!/bin/bash
echo "X-Spam-Status: Yes, score=10.0"
echo "X-Spam-Level: **********"
cat
`
	
	tmpFile, err := os.CreateTemp("", "mock_spamc_*.sh")
	if err != nil {
		t.Skip("Cannot create temporary file for test")
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(tmpScript); err != nil {
		t.Skip("Cannot write to temporary file for test")
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		t.Skip("Cannot make temporary file executable")
	}

	session := createTestSession()
	session.Config.Spamc = tmpFile.Name()

	result, err := session.checkSpam()

	if err != nil {
		t.Errorf("Expected no error with mock spamc script, got: %v", err)
	}

	if !strings.Contains(result, "X-Spam-Status: Yes, score=10.0") {
		t.Errorf("Expected spam headers to be added to result")
	}

	if !strings.Contains(result, "X-Spam-Level: **********") {
		t.Errorf("Expected spam level header to be added to result")
	}

	if !strings.Contains(result, "This is a test message body.") {
		t.Errorf("Expected original message body to be preserved")
	}
}

func TestCheckSpam_WithInvalidSpamcCommand(t *testing.T) {
	session := createTestSession()
	session.Config.Spamc = "/nonexistent/command"

	// This should cause a fatal error in the current implementation
	// We can't easily test log.Fatal, but we can verify the behavior
	// In a real implementation, this might be refactored to return an error instead
	defer func() {
		if r := recover(); r != nil {
			// Expected behavior due to log.Fatal in the code
		}
	}()

	_, err := session.checkSpam()
	
	// If we get here without a panic, the error handling might have been improved
	if err == nil {
		t.Errorf("Expected error with invalid spamc command")
	}
}

func TestCheckSpam_EmptyMessage(t *testing.T) {
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("cat command not available for testing")
	}

	session := createTestSession()
	session.Config.Spamc = "cat"
	session.Data = ""

	result, err := session.checkSpam()

	if err != nil {
		t.Errorf("Expected no error with empty message, got: %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result for empty message, got: %q", result)
	}
}

func TestCheckSpam_LargeMessage(t *testing.T) {
	if _, err := exec.LookPath("cat"); err != nil {
		t.Skip("cat command not available for testing")
	}

	session := createTestSession()
	session.Config.Spamc = "cat"
	
	// Create a large message
	largeBody := strings.Repeat("This is a line of text in a large message.\n", 1000)
	session.Data = "Subject: Large Test Message\n\n" + largeBody

	result, err := session.checkSpam()

	if err != nil {
		t.Errorf("Expected no error with large message, got: %v", err)
	}

	if !strings.Contains(result, "Large Test Message") {
		t.Errorf("Expected subject to be preserved in large message")
	}

	if !strings.Contains(result, "This is a line of text in a large message.") {
		t.Errorf("Expected body content to be preserved in large message")
	}
}
