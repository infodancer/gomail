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
	User string
	// Host represnets the hostname
	Domain string
	// Folder represents the subfolder portion of the address separated by a -
	Folder string
}

func Parse(input string) (user string, folder string, host string) {
	if sep1 := strings.Index(input, "@"); sep1 != -1 {
		user = input[0:sep1]
		host = input[sep1+1:]
	} else {
		user = input
	}
	if sep2 := strings.Index(user, `-`); sep2 != -1 {
		folder = user[sep2+1:]
		user = user[0:sep2]
	}
	return
}

func GetUser(s string) string {
	user, _, _ := Parse(s)
	return user
}

func GetFolder(s string) string {
	_, folder, _ := Parse(s)
	return folder
}

func GetHost(s string) string {
	_, _, host := Parse(s)
	return host
}

// String produces a string version of the parsed address
func (address Address) String() string {
	result := address.User
	if len(address.Folder) > 0 {
		result = result + "-" + address.Folder
	}
	if len(address.Domain) > 0 {
		result = result + "@" + address.Domain
	}
	return result
}

// CreateAddress creates an address structure from an input user@host
func CreateAddress(input string) (*Address, error) {
	result := Address{}
	if sep1 := strings.Index(input, "@"); sep1 != -1 {
		user := input[0:sep1]
		host := input[sep1+1:]
		if sep2 := strings.Index(user, "-"); sep2 != -1 {
			user = user[0:sep2]
			folder := user[sep2:]
			result.Folder = folder
		}
		result.User = user
		result.Domain = host
		return &result, nil
	}
	return nil, errors.New("Address not found in command")
}

// LookupMX returns the list of mxes for an address, sorted by priority
func LookupMX(address *Address) ([]*net.MX, error) {
	mxrecords, err := net.LookupMX(address.Domain)
	if err != nil {

		return nil, err
	}
	for _, mx := range mxrecords {
		fmt.Println(mx.Host, mx.Pref)
	}
	return mxrecords, err
}
