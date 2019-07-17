package mockdsp

import (
	"fmt"
	"net/http"
	"os"

	"github.com/remixd-media/prebid-server/remixd/mockdsp/handlers"
)

func StartServer() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8100"
	}

	app := &handlers.App{}

	mux := http.NewServeMux()
	mux.HandleFunc("/openrtb2/auction", app.AuctionHandler)

	fmt.Printf("Listening on port %s\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
}
