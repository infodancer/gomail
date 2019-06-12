#!/bin/sh
echo "HELO hi"
echo "MAIL FROM:<test@example.com>"
echo "RCPT TO:<test@example.com>"
echo DATA
echo From: test@example.com
echo To: test@example.com
echo Subject: This is a test message
echo ""
echo Yes, this is an automated test.
echo .

