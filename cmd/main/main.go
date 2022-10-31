package main

import (
	"github.com/nkvoll/innosonix-maxx/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
