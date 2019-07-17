package main

import (
	"log"

	"github.com/remixd-media/prebid-server/remixd/mockdsp"
)

func main() {
	err := mockdsp.StartServer()
	if err != nil {
		log.Fatal(err)
	}
}
