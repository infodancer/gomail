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

func TestGetUser(t *testing.T) {
	if GetUser("test@example.com") != "test" {
		t.Error("expected \"test\"")
	}
	if GetUser("test-folder@example.com") != "test" {
		t.Error("expected \"test\"")
	}
	if GetUser("test") != "test" {
		t.Error("expected \"test\"")
	}
	if GetUser("test@") != "test" {
		t.Error("expected \"test\"")
	}
}

func TestGetHost(t *testing.T) {
	if GetHost("test@example.com") != "example" {
		t.Error("expected \"example\"")
	}
	if GetHost("test-folder@example.com") != "example" {
		t.Error("expected \"example\"")
	}
	if GetHost("test") != "" {
		t.Error("expected \"\"")
	}
	if GetHost("test@") != "" {
		t.Error("expected \"\"")
	}
}

func TestGetFolder(t *testing.T) {
	if GetFolder("test@example.com") != "" {
		t.Error("expected \"\"")
	}
	if GetFolder("test-folder@example.com") != "folder" {
		t.Error("expected \"folder\"")
	}
	if GetFolder("test-folder") != "folder" {
		t.Error("expected \"folder\"")
	}
	if GetFolder("test") != "" {
		t.Error("expected \"\"")
	}
	if GetFolder("test@") != "" {
		t.Error("expected \"\"")
	}
}
