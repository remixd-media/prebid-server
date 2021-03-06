package bidswitch

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/prebid/prebid-server/config"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type BidSwitchAdapter struct {
	endpoint string
}

func (adapter *BidSwitchAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	requests := []*adapters.RequestData{}

	requestCopy := *request

	// copy impressions
	requestCopy.Imp = append([]openrtb.Imp{}, request.Imp...)
	//copy site
	if request.Site != nil {
		siteCopy := *request.Site
		requestCopy.Site = &siteCopy
	}

	for i := range requestCopy.Imp {
		requestCopy.Imp[i].Ext = nil // remove unused imp ext
		requestCopy.Imp[i].PMP = &openrtb.PMP{
			PrivateAuction: 1,
			Deals: []openrtb.Deal{
				{
					ID:          "Remixd-Premium-Audio-Blis",
					BidFloor:    0,
					BidFloorCur: "USD",
					AT:          1,
				},
			},
		}
	}
	if requestCopy.Site != nil {
		requestCopy.Site.Ext = nil // remove unused site ext
	}

	jsonBody, err := json.Marshal(requestCopy)
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
		Uri:     adapter.endpoint,
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

	var responseBody []byte
	if strings.Contains(response.Headers.Get("Content-Encoding"), "gzip") {
		// unzip response
		gzipReader, err := gzip.NewReader(bytes.NewReader(response.Body))
		if err != nil {
			return nil, []error{&errortypes.BadServerResponse{
				Message: fmt.Sprintf("Bad server response: gunzip err %d", response.StatusCode),
			}}
		}
		responseBody, err = ioutil.ReadAll(gzipReader)
		if err != nil {
			return nil, []error{&errortypes.BadServerResponse{
				Message: fmt.Sprintf("Bad server response: gunzip read err %d", response.StatusCode),
			}}
		}
	} else {
		responseBody = response.Body
	}

	fmt.Printf("bidswitch makeBids response body: %s\n", string(responseBody))

	var bidResponse openrtb.BidResponse
	err := json.Unmarshal(responseBody, &bidResponse)
	if err != nil {
		return nil, []error{fmt.Errorf("bid response decode: %v", err)}
	}

	fmt.Printf("bidswitch makeBids bidResponse: %+v\n", bidResponse)

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

func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &BidSwitchAdapter{
		endpoint: config.Endpoint,
	}
	return bidder, nil
}
