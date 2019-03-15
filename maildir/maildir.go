package maildir

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

var deliveryCounter int64

// CreateMaildir creates a maildir directory structure
func CreateMaildir(name string) {

}

func createUniqueName() string {
	date := time.Now()
	left := date.Nanosecond()
	center := rand.Int63()
	right, err := os.Hostname()
	if err != nil {

	}

	result := fmt.Sprintf("%v.%v.%v", left, center, right)
	return result
}
