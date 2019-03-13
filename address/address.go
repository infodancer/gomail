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
	Host *string
}

// CreateAddress creates an address structure from an input user@host
func CreateAddress(input string) (*Address, error) {
	sep := strings.Index(input, "@")
	if sep == -1 {
		return nil, errors.New("Address not found in command")
	}
	user := input[0:sep]
	host := input[sep+1 : len(input)]
	result := Address{User: &user, Host: &host}
	return &result, nil
}

// LookupMX returns the list of mxes for an address, sorted by priority
func LookupMX(address *Address) ([]*net.MX, error) {
	mxrecords, err := net.LookupMX(*address.Host)
	if err != nil {

		return nil, err
	}
	for _, mx := range mxrecords {
		fmt.Println(mx.Host, mx.Pref)
	}
	return mxrecords, err
}
