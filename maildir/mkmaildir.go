package maildir

import (
	"log"
	"os"
)

func Main() {
	args := os.Args[1:]

	for _, arg := range args {
		_, err := Create(arg)
		if err != nil {
			log.Fatalf("could not create maildir in %v: %v\n", arg, err)
		}
	}
}
