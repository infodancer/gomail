package address

import "testing"

// TestCreateAddress
func TestCreateAddress(t *testing.T) {
	addr1 := "test@example.com"
	if addr, err := CreateAddress(addr1); err == nil {
		if *addr.Domain != "example.com" {
			t.Error("domain was not example.com:", addr1)
		}
		if *addr.User != "test" {
			t.Error("user was not test: ", addr1)
		}
		if addr.Folder != nil {
			t.Error("folder was returned when not present", addr1)
		}
	} else {
		t.Error("Could not create address from ", addr1, err)
	}
}
