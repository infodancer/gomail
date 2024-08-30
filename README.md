# Gomail

Gomail is an implementation of internet mail protocols (SMTP, POP3, IMAP) in go.

The initial goal is to produce an efficient, secure, small mail system that is suitable for individuals 
to set up and run easily in a modern mail environment without a lot of administrative overhead; the system 
can be run in a normal environment or within a docker container with a mounted local volume for data persistence.

Configuration will be handled by manually modifying the volume initially.  

The author is <matthew@infodancer.org>.

Intended built-in features (in no particular order):

* Basic sending and receiving via SMTP
* Queueing for sent mail
* SMTP AUTH (CRAM-MD5)
* SMTP submission support (port 587, authenticated only)
* Maildir support
* Sending mail remotely (non-SSL)
* Collecting mail via POP3
* Support for passing messages through external program filters
* SpamAssassin support (via spamc)
* Debian packages
* SSL/TLS
* Integrated mailing list support (with VERP)
* milter support
* DKIM support (sending and receiving)
* SPF support (checking incoming messages)
* Greylisting support
* IMAP support
* (Partial) support for checkpassword
* (Maybe eventually) Administration via REST API

# Developer Requirements

Relies on [Task](https://taskfile.dev/) for building binaries and [nfpm](https://nfpm.goreleaser.com/) to build deb and rpm packages.  Changlelogs are maintained with [chglog](https://github.com/goreleaser/chglog).
