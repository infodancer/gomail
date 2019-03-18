package domain

import (
	"errors"
	"fmt"
	"os/user"
	"regexp"
	"strings"
)

// Domain holds information about a domain
type Domain struct {
	Name string
	Path string
}

var domainRoot string

func init() {
	domainRoot = "/srv/domains"
}

// GetDomain provides a domain object based on the domain root and the provided name
func GetDomain(name string) (*Domain, error) {
	var result Domain
	result.Name = name
	result.Path = domainRoot + "/" + name
	return &result, nil
}

// GetUser provides a user object based on the current domain and the provided user name
func (domain *Domain) GetUser(name string) (*user.User, error) {
	return nil, nil
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
