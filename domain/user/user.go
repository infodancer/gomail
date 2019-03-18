package user

// GetMaildir retrieves the top-level maildir
func GetMaildir() *maildir.Maildir {
	maildir := createMaildir()
	return maildir
}
