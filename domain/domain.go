package domain

import (
	"errors"
	"fmt"

	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/infodancer/gomail/maildir"
)

const validateDomainPattern string = "[^a-zA-Z0-9_-]+"

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
var domainPattern *regexp.Regexp

func init() {
	var err error
	domainRoot = os.Getenv("DOMAIN_ROOT")
	if domainRoot == "" {
		domainRoot = "/srv/domains"
	}
	logger = log.New(os.Stderr, "", 0)
	domainPattern, err = regexp.Compile(validateDomainPattern)
	if err != nil {
		err := fmt.Errorf("could not validate requested name; invalid regular expression: %w", err)
		logger.Println(err)
	}
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
	userpath := filepath.Join(domain.Path, "users", name)
	logger.Println("Checking user path " + userpath)
	if _, err := os.Stat(userpath); os.IsNotExist(err) {
		err := fmt.Errorf("user does not exist: %v", err)
		return nil, err
	}
	user := User{}
	user.Name = name
	user.Path = filepath.Join(domain.Path, "users", name)
	user.MaildirPath = filepath.Join(user.Path, "Maildir")
	return &user, nil
}

// GetUserMaildir retrieves the top-level maildir for a specified user
func (domain *Domain) GetUserMaildir(name string) (*maildir.Maildir, error) {
	user, err := domain.GetUser(name)
	if err != nil {
		err := fmt.Errorf("user does not exist: %v", err)
		return nil, err
	}
	logger.Println("Checking for maildir at " + user.MaildirPath)
	result, err := maildir.New(user.MaildirPath)
	if err != nil {
		err := fmt.Errorf("could not load maildir from %v: %v", user.MaildirPath, err)
		return nil, err
	}
	if result == nil {
		err := fmt.Errorf("maildir does not exist: %v", user.MaildirPath)
		return nil, err
	}
	return result, nil
}

// validateDomainName ensures a domain name is safe to use for a filename
func ValidateDomainName(name string) error {
	result := domainPattern.ReplaceAllString(name, "")
	if len(result) != len(name) {
		err := fmt.Errorf("requested name contained illegal characters")
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
