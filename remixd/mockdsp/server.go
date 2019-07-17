package mockdsp

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/remixd-media/prebid-server/remixd/mockdsp/handlers"
)

func StartServer() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8100"
	}

	var rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	app := &handlers.App{
		Rand: rand,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/openrtb2/auction", app.AuctionHandler)

	fmt.Printf("Listening on port %s\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
}
