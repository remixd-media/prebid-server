package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gofrs/uuid"

	"github.com/mxmCherry/openrtb"
)

func genID() (string, error) {
	uuidv4, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return uuidv4.String(), nil
}

func (app *App) AuctionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%v] New Auction Request\n", time.Now())

	bidRequest := &openrtb.BidRequest{}
	err := readJSON(r, bidRequest)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	bidResponse := &openrtb.BidResponse{}
	bidResponse.ID = bidRequest.ID
	bidID, err := genID()
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	bidResponse.BidID = bidID
	bidResponse.Cur = "USD"

	seatBid := openrtb.SeatBid{
		Seat: "mockdsp-" + os.Getenv("PORT"),
	}

	for _, imp := range bidRequest.Imp {
		bid := openrtb.Bid{}

		impBidID, err := genID()
		if err != nil {
			JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		bid.ID = impBidID

		bid.ImpID = imp.ID
		bid.Price = imp.BidFloor * 1.1
		bid.NURL = fmt.Sprintf("http://localhost:%s/winnotice?impbidid=%s", os.Getenv("PORT"), impBidID)
		bid.CrID = "creative-1"

		seatBid.Bid = append(seatBid.Bid, bid)
	}

	bidResponse.SeatBid = append(bidResponse.SeatBid, seatBid)

	JSONOk(w, bidResponse)
}
