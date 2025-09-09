FROM scratch

# Add CA certificates for TLS
ADD ca-certificates.crt /etc/ssl/certs/

# Add binaries
ADD listener /usr/local/bin/listener
ADD smtpd /usr/local/bin/smtpd
ADD pop3d /usr/local/bin/pop3d

# Create configuration directory
ADD configs/ /etc/gomail/

# Set up volumes for persistent storage
VOLUME ["/var/mail", "/var/spool/mail", "/etc/ssl/private"]

# Expose standard mail ports
EXPOSE 25    # SMTP
EXPOSE 110   # POP3
EXPOSE 143   # IMAP (for future use)
EXPOSE 465   # SMTPS (SMTP over SSL)
EXPOSE 587   # SMTP Submission
EXPOSE 993   # IMAPS (for future use)
EXPOSE 995   # POP3S (POP3 over SSL)

LABEL maintainer="matthew@infodancer.org"
LABEL description="Multi-protocol mail server with TLS support"

# Run listener with all configuration files
CMD ["/usr/local/bin/listener", \
     "/etc/gomail/smtp.toml", \
     "/etc/gomail/smtps.toml", \
     "/etc/gomail/submission.toml", \
     "/etc/gomail/pop3.toml", \
     "/etc/gomail/pop3s.toml"]
