package bidswitch

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type BidSwitchAdapter struct {
	URI string
}

func (adapter *BidSwitchAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	requests := []*adapters.RequestData{}

	for i := range request.Imp {
		request.Imp[i].Video = nil // remove unused video section
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, []error{err}
	}

	gzipBuf := bytes.Buffer{}
	gzipWriter := gzip.NewWriter(&gzipBuf)
	_, err = gzipWriter.Write(jsonBody)
	if err != nil {
		return nil, []error{err}
	}
	gzipWriter.Close()

	headers := http.Header{}
	headers.Set("Content-Encoding", "gzip")
	headers.Set("Accept-Encoding", "gzip")

	reqData := adapters.RequestData{
		Method:  http.MethodPost,
		Uri:     adapter.URI,
		Headers: headers,
	}

	fmt.Printf("bidswitch makerequests jsonBody: %s\n", string(jsonBody))

	if os.Getenv("PBS_BIDSWITCH_NO_GZIP") == "1" {
		reqData.Body = jsonBody
	} else {
		reqData.Body = gzipBuf.Bytes()
	}

	requests = append(requests, &reqData)

	return requests, nil
}

func (adapter *BidSwitchAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	fmt.Printf("bidswitch makebids start: (ip: %v | page: %v | status: %d)\n", internalRequest.Device.IP, internalRequest.Site.Page, response.StatusCode)
	defer fmt.Printf("bidswitch makebids done: (ip: %v | page: %v)\n", internalRequest.Device.IP, internalRequest.Site.Page)

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

	var bidResponse openrtb.BidResponse
	err := json.Unmarshal(response.Body, &bidResponse)
	if err != nil {
		return nil, []error{fmt.Errorf("bid response decode: %v", err)}
	}

	fmt.Printf("bidswitch MakeBids: bidResponse: %+v\n", bidResponse)

	if len(bidResponse.SeatBid) == 0 {
		// no bid
		return nil, nil
	}

	bidderResponse := adapters.NewBidderResponseWithBidsCapacity(len(bidResponse.SeatBid))
	for _, seatBid := range bidResponse.SeatBid {
		for i := range seatBid.Bid {
			bidderResponse.Bids = append(bidderResponse.Bids, &adapters.TypedBid{
				Bid:     &seatBid.Bid[i],
				BidType: openrtb_ext.BidTypeVideo,
			})
		}
	}

	return bidderResponse, nil
}

func NewBidSwitchBidder(endpoint string) *BidSwitchAdapter {
	return &BidSwitchAdapter{
		URI: endpoint,
	}
}
