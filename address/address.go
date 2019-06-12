package address

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Address represents an internet email address user@host
type Address struct {
	// User represents the username
	User *string
	// Host represnets the hostname
	Domain *string
	// Folder represents the subfolder portion of the address separated by a -
	Folder *string
}

// ToString produces a string version of the parsed address
func (address Address) ToString() (string, error) {
	if address.User != nil {
		result := *address.User
		if address.Folder != nil {
			result = result + "-" + *address.Folder
		}
		if address.Domain != nil {
			result = result + "@" + *address.Domain
		}
		return result, nil
	}
	return "", errors.New("invalid address; user component is required")
}

// CreateAddress creates an address structure from an input user@host
func CreateAddress(input string) (*Address, error) {
	result := Address{}
	if sep1 := strings.Index(input, "@"); sep1 != -1 {
		user := input[0:sep1]
		host := input[sep1+1 : len(input)]
		if sep2 := strings.Index(user, "-"); sep2 != -1 {
			user = user[0:sep2]
			folder := user[sep2:len(user)]
			result.Folder = &folder
		}
		result.User = &user
		result.Domain = &host
		return &result, nil
	}
	return nil, errors.New("Address not found in command")
}

// LookupMX returns the list of mxes for an address, sorted by priority
func LookupMX(address *Address) ([]*net.MX, error) {
	mxrecords, err := net.LookupMX(*address.Domain)
	if err != nil {

		return nil, err
	}
	for _, mx := range mxrecords {
		fmt.Println(mx.Host, mx.Pref)
	}
	return mxrecords, err
}
