package smtpd

import (
	"strings"
	"testing"
)

// TestextractAddress extracts an address from a RCPT TO or MAIL FROM line
func TestExtractAddress(t *testing.T) {
	expected := "test@example.com"
	testMailFrom1 := "MAIL FROM:<test@example.com>"
	if mf1, _ := extractAddress(testMailFrom1); mf1 != expected {
		t.Error("extractAddress failed for ", testMailFrom1, mf1)
	}
	testRcptTo1 := "RCPT TO:<test@example.com>"
	if rcpt1, _ := extractAddress(testMailFrom1); rcpt1 != expected {
		t.Error("extractAddress failed for ", testRcptTo1, rcpt1)
	}
}

func TestHandleInputLine(t *testing.T) {
	var session Session
	if code, result, finished := handleInputLine(session, "HELO hi"); code != 250 || !strings.HasSuffix(result, "Hello") || finished {
		t.Error("Invalid response to HELO: ", result)
	}
	if code, result, finished := handleInputLine(session, "EHLO hi"); code != 250 || !strings.HasSuffix(result, " Hello") || finished {
		t.Error("Invalid response to EHLO: ", result)
	}
	if code, result, finished := handleInputLine(session, "NOOP"); code != 250 || !strings.HasPrefix(result, "OK") || finished {
		t.Error("Invalid response to NOOP: ", result)
	}
	if code, result, finished := handleInputLine(session, "RSET"); code != 250 || !strings.HasPrefix(result, "OK") || finished {
		t.Error("Invalid response to RSET: ", result)
	}
	if code, result, finished := handleInputLine(session, "VRFY"); code != 500 || !strings.HasPrefix(result, "VRFY not supported") || finished {
		t.Error("Invalid response to VRFY: ", result)
	}
	if code, result, finished := handleInputLine(session, "MAIL FROM:<test@example.com>"); code != 250 || !strings.HasPrefix(result, "OK") || finished {
		t.Error("Invalid response to MAIL FROM: ", code, result, finished)
	}
	if code, result, finished := handleInputLine(session, "QUIT"); code != 221 || strings.HasPrefix(result, " goodbye") || !finished {
		t.Error("Invalid response to QUIT: ", result)
	}
}
