package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
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

	b := bytes.Buffer{}
	tr := io.TeeReader(r.Body, &b)

	bidRequest := &openrtb.BidRequest{}
	err := json.NewDecoder(tr).Decode(bidRequest)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("bidRequest: %v\n", b.String())

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
		bid.Price = math.Round(imp.BidFloor*(1+app.Rand.Float64())*100) / 100
		bid.NURL = fmt.Sprintf("http://localhost:%s/winnotice?impbidid=%s", os.Getenv("PORT"), impBidID)
		bid.CrID = "creative-1"

		var impExt map[string]interface{}
		err = json.Unmarshal(imp.Ext, &impExt)
		if err != nil {
			JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, ok := impExt["appnexus"]; ok {
			bid.Ext = json.RawMessage(`{
				"appnexus": {
				  "brand_id": 1,
				  "brand_category_id": 1,
				  "auction_id": 8189378542222915032,
				  "bid_ad_type": 0,
				  "bidder_id": 2,
				  "ranking_price": 0.000000
				}
			  }`)
		} else if _, ok := impExt["pubmatic"]; ok {
		}

		seatBid.Bid = append(seatBid.Bid, bid)
	}

	bidResponse.SeatBid = append(bidResponse.SeatBid, seatBid)

	JSONOk(w, bidResponse)
}
