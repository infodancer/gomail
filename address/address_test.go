package address

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateAddress
func TestCreateAddress(t *testing.T) {
	addr1 := "test@example.com"
	addr, err := CreateAddress(addr1)
	assert.NoError(t, err)
	assert.Equal(t, "example.com", addr.Domain)
	assert.Equal(t, "test", addr.User)
	assert.Equal(t, "", addr.Folder)
}

func TestGetUser(t *testing.T) {
	assert.Equal(t, "test", GetUser("test@example.com"))
	assert.Equal(t, "test", GetUser("test-folder@example.com"))
	assert.Equal(t, "test", GetUser("test"))
	assert.Equal(t, "test", GetUser("test@"))
}

func TestGetHost(t *testing.T) {
	assert.Equal(t, "example.com", GetHost("test@example.com"))
	assert.Equal(t, "example.com", GetHost("test-folder@example.com"))
	assert.Equal(t, "", GetHost("test"))
	assert.Equal(t, "", GetHost("test@"))
}

func TestGetFolder(t *testing.T) {
	assert.Equal(t, "", GetFolder("test@example.com"))
	assert.Equal(t, "folder", GetFolder("test-folder@example.com"))
	assert.Equal(t, "folder", GetFolder("test-folder"))
	assert.Equal(t, "", GetFolder("test"))
	assert.Equal(t, "", GetFolder("test@"))
}
