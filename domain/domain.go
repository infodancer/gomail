package domain

import (
	"errors"
	"fmt"

	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"infodancer.org/gomail/maildir"
)

// Domain holds information about a domain
type Domain struct {
	Name string
	Path string
}

// User is a user within a domain
type User struct {
	Name        string
	Path        string
	MaildirPath string
}

var logger *log.Logger
var domainRoot string

func init() {
	domainRoot = "/srv/domains"
	logger = log.New(os.Stderr, "", 0)
}

// SetDomainRoot sets the root directory for the domain heirarchy; by default, /srv/domains
func SetDomainRoot(path string) {
	domainRoot = path
}

// GetDomain provides a domain object based on the domain root and the provided name
func GetDomain(name string) (*Domain, error) {
	var result Domain
	result.Name = name
	result.Path = filepath.Join(domainRoot, name)
	logger.Println("Checking domain path " + result.Path)
	if _, err := os.Stat(result.Path); os.IsNotExist(err) {
		err := fmt.Errorf("requested domain %v does not exist or cannot be accessed: %v", result.Path, err)
		return nil, err
	}
	return &result, nil
}

// GetUser provides a user object based on the current domain and the provided user name
func (domain *Domain) GetUser(name string) (*User, error) {
	userpath := filepath.Join(domain.Path, name)
	logger.Println("Checking user path " + userpath)
	if _, err := os.Stat(userpath); os.IsNotExist(err) {
		err := fmt.Errorf("user does not exist: %v", err)
		return nil, err
	}
	var user User
	user = User{}
	return &user, nil
}

// GetUserMaildir retrieves the top-level maildir for a specified user
func (domain *Domain) GetUserMaildir(name *string) (*maildir.Maildir, error) {
	user, err := domain.GetUser(*name)
	if err != nil {
		err := fmt.Errorf("user does not exist: %v", err)
		return nil, err
	}
	return maildir.GetMaildir(user.MaildirPath)
}

// validateDomainName ensures a domain name is safe to use for a filename
func validateDomainName(name string) error {
	pattern := "[^a-zA-Z0-9_-]+"
	exp, err := regexp.Compile(pattern)
	if err != nil {
		err := fmt.Errorf("could not validate requested name; invalid regular expression: %v", err)
		return err
	}

	result := exp.ReplaceAllString(name, "")
	if len(result) != len(name) {
		err := fmt.Errorf("requested name contained illegal characters; should match %v", pattern)
		return err
	}

	return nil
}

// extractDomainPath transforms a domain into a filesystem path
func extractDomainPath(input string) (string, error) {
	var result string
	parts := strings.Split(input, ".")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return "", errors.New("extraneous dot detected")
		}
		result += "/" + part
	}

	return result, nil
}
