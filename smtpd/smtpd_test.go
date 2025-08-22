package smtpd

import (
	"testing"

	"github.com/infodancer/gomail/config"
	"github.com/infodancer/gomail/connect"
	"github.com/stretchr/testify/assert"
)

// TestextractAddress extracts an address from a RCPT TO or MAIL FROM line
func TestExtractAddressPart(t *testing.T) {
	expected := "test@example.com"
	testMailFrom1 := "MAIL FROM:<test@example.com>"
	if mf1, _ := extractAddressPart(testMailFrom1); *mf1 != expected {
		t.Error("extractAddress failed for ", testMailFrom1, mf1)
	}
	testRcptTo1 := "RCPT TO:<test@example.com>"
	if rcpt1, _ := extractAddressPart(testMailFrom1); *rcpt1 != expected {
		t.Error("extractAddress failed for ", testRcptTo1, rcpt1)
	}
}

func TestIsSuspiciousAddress(t *testing.T) {
	example1 := "test@example.com"
	example2 := "test-folder@example.com"
	example3 := "../../../test@example.com"
	if IsSuspiciousAddress(example1) {
		t.Error("IsSuspiciousAddress reported valid address as suspicious: ", example1)
	}
	if IsSuspiciousAddress(example2) {
		t.Error("IsSuspiciousAddress reported valid address as suspicious: ", example2)
	}
	if !IsSuspiciousAddress(example3) {
		t.Error("IsSuspiciousAddress reported suspicious address as valid: ", example3)
	}

}

func TestHandleInputLine(t *testing.T) {
	cfg := Config{ServerConfig: config.ServerConfig{ServerName: "testserver"}}
	c, err := connect.NewStandardIOConnection()
	assert.NoError(t, err, "unable to create connection")
	s := &Session{Conn: c, Config: cfg}
	success := 250
	endsession := 221
	failure := 500
	code, result, finished := s.HandleInputLine("HELO hi")
	assert.Equal(t, success, code, "result code was not 250")
	assert.Contains(t, result, "testserver")
	assert.False(t, finished)
	code, result, finished = s.HandleInputLine("EHLO hi")
	assert.Equal(t, success, code, "result code was not 250")
	assert.Contains(t, result, "testserver")
	assert.False(t, finished)
	code, result, finished = s.HandleInputLine("NOOP")
	assert.Equal(t, success, code, "result code was not 250")
	assert.Contains(t, result, "OK")
	assert.False(t, finished)
	code, result, finished = s.HandleInputLine("RSET")
	assert.Equal(t, success, code, "result code was not 250")
	assert.Contains(t, result, "OK")
	assert.False(t, finished)
	code, result, finished = s.HandleInputLine("VRFY")
	assert.Equal(t, failure, code, "VRFY command did not return 500")
	assert.Contains(t, result, "VRFY not supported")
	assert.False(t, finished)
	code, result, finished = s.HandleInputLine("VRFY test@example.com")
	assert.Equal(t, failure, code, "VRFY command did not return 500")
	assert.Contains(t, result, "VRFY not supported")
	assert.False(t, finished)
	code, result, finished = s.HandleInputLine("MAIL FROM:<test@example.com>")
	assert.Equal(t, success, code, "MAIL FROM command did not return 250")
	assert.Contains(t, result, "OK")
	assert.False(t, finished)
	code, result, finished = s.HandleInputLine("QUIT")
	assert.Equal(t, endsession, code, "QUIT command did not return 221")
	assert.Contains(t, result, "goodbye")
	assert.True(t, finished)
}
