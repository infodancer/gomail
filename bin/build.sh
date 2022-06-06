#!/bin/sh
go build -o bin/smtpd cmd/smtpd/smtpd.go
go build -o bin/pop3d cmd/pop3d/pop3d.go
