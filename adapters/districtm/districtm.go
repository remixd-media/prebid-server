package districtm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type DistrictMAdapter struct {
	URI string
}

func (adapter *DistrictMAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	requests := []*adapters.RequestData{}

	errors := []error{}

	for i := range request.Imp {
		impExt, err := parseExt(&request.Imp[i])
		if err != nil {
			errors = append(errors, err)
			continue
		}

		request.Imp[i].Audio = nil // remove unused audio section
		request.Imp[i].Ext = nil   // remove unused imp ext
		if request.Site != nil {
			request.Site.Ext = nil // remove unused site ext

			if request.Site.Publisher == nil {
				request.Site.Publisher = &openrtb.Publisher{}
			}
			request.Site.Publisher.ID = impExt.PublisherID
		}

		/*	if request.Imp[i].Video != nil {
			request.Imp[i].Video.W = 1
			request.Imp[i].Video.H = 1
		}*/

		request.Imp[i].TagID = "DMX-Remixd"

		if request.User == nil || request.User.BuyerUID == "" {
			return nil, []error{fmt.Errorf("districtm no user buyer id")}
		}
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, []error{err}
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	reqData := adapters.RequestData{
		Method:  http.MethodPost,
		Uri:     adapter.URI,
		Headers: headers,
	}

	fmt.Printf("districtm makerequests jsonBody: %s\n", string(jsonBody))
	reqData.Body = jsonBody

	requests = append(requests, &reqData)

	if len(errors) != 0 {
		fmt.Printf("districtm errors: %+v\n", errors)
		return nil, errors
	}

	return requests, nil
}

func parseExt(imp *openrtb.Imp) (*openrtb_ext.ExtImpDistrictM, error) {
	var bidderExt adapters.ExtImpBidder

	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding extImpBidder, err: %s", imp.ID, err),
		}
	}

	impExt := openrtb_ext.ExtImpDistrictM{}
	err := json.Unmarshal(bidderExt.Bidder, &impExt)
	if err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding impExt, err: %s", imp.ID, err),
		}
	}

	return &impExt, nil
}

func (adapter *DistrictMAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	fmt.Printf("districtm makebids start: (ip: %v | page: %v | status: %d)\n", internalRequest.Device.IP, internalRequest.Site.Page, response.StatusCode)
	defer fmt.Printf("districtm makebids done: (ip: %v | page: %v)\n", internalRequest.Device.IP, internalRequest.Site.Page)

	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Bad user input: HTTP status %d", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Bad server response: HTTP status %d", response.StatusCode),
		}}
	}

	fmt.Printf("districtm makeBids response body: %s\n", string(response.Body))

	var bidResponse openrtb.BidResponse
	err := json.Unmarshal(response.Body, &bidResponse)
	if err != nil {
		return nil, []error{fmt.Errorf("bid response decode: %v", err)}
	}

	fmt.Printf("districtm makeBids bidResponse: %+v\n", bidResponse)

	if len(bidResponse.SeatBid) == 0 {
		// no bid
		return nil, nil
	}

	bidderResponse := adapters.NewBidderResponseWithBidsCapacity(len(bidResponse.SeatBid))
	for _, seatBid := range bidResponse.SeatBid {
		for i := range seatBid.Bid {
			bid := &seatBid.Bid[i]
			if bid.BURL == "" {
				// fix to move NURL to BURL, districtm specific issue
				bid.BURL = bid.NURL
				bid.NURL = ""
			}

			bidderResponse.Bids = append(bidderResponse.Bids, &adapters.TypedBid{
				Bid:     bid,
				BidType: openrtb_ext.BidTypeVideo,
			})
		}
	}

	return bidderResponse, nil
}

func NewDistrictMBidder(endpoint string) *DistrictMAdapter {
	return &DistrictMAdapter{
		URI: endpoint,
	}
}
