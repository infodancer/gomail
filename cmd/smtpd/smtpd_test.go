package main

import (
	"strings"
	"testing"
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
	if isSuspiciousAddress(example1) {
		t.Error("IsSuspiciousAddress reported valid address as suspicious: ", example1)
	}
	if isSuspiciousAddress(example2) {
		t.Error("IsSuspiciousAddress reported valid address as suspicious: ", example2)
	}
	if !isSuspiciousAddress(example3) {
		t.Error("IsSuspiciousAddress reported suspicious address as valid: ", example3)
	}

}

func TestHandleInputLine(t *testing.T) {
	s := Session{}

	if code, result, finished := s.HandleInputLine("HELO hi"); code != 250 || !strings.HasSuffix(result, "Hello") || finished {
		t.Error("Invalid response to HELO: ", result)
	}
	if code, result, finished := s.HandleInputLine("EHLO hi"); code != 250 || !strings.HasSuffix(result, "Hello") || finished {
		t.Error("Invalid response to EHLO: ", result)
	}
	if code, result, finished := s.HandleInputLine("NOOP"); code != 250 || !strings.HasPrefix(result, "OK") || finished {
		t.Error("Invalid response to NOOP: ", result)
	}
	if code, result, finished := s.HandleInputLine("RSET"); code != 250 || !strings.HasPrefix(result, "OK") || finished {
		t.Error("Invalid response to RSET: ", result)
	}
	if code, result, finished := s.HandleInputLine("VRFY"); code != 500 || !strings.HasPrefix(result, "VRFY not supported") || finished {
		t.Error("Invalid response to VRFY: ", result)
	}
	if code, result, finished := s.HandleInputLine("MAIL FROM:<test@example.com>"); code != 250 || !strings.HasPrefix(result, "OK") || finished {
		t.Error("Invalid response to MAIL FROM: ", code, result, finished)
	}
	if code, result, finished := s.HandleInputLine("QUIT"); code != 221 || strings.HasPrefix(result, " goodbye") || !finished {
		t.Error("Invalid response to QUIT: ", result)
	}
}
