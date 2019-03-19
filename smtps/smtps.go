package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	sender := flag.String("helo", "h", "The helo string to use when greeting clients")
	recipient := flag.String("helo", "h", "The helo string to use when greeting clients")
	mxhost := flag.String("mx", "", "Override the mx lookup and use the specified host")
	user := flag.String("user", "", "")
	password := flag.String("password", "", "")

	logger := log.New(os.Stderr, "", 0)
	logger.Println("gomail smtps started")
}
