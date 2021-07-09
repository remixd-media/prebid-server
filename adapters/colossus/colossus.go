package colossus

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type ColossusAdapter struct {
	URI string
}

// Builder builds a new instance of the Colossus adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &ColossusAdapter{
		URI: config.Endpoint,
	}
	return bidder, nil
}

// MakeRequests create bid request for colossus demand
func (adapter *ColossusAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	var errors []error
	var requests []*adapters.RequestData

	for _, imp := range request.Imp {
		if imp.Audio == nil {
			continue
		}

		ext, err := parseExt(&imp)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		params := url.Values{}
		params.Add("dnt", "0")
		params.Add("placementId", ext.PlacementId)
		if request.Site != nil {
			params.Add("domain", request.Site.Domain)
			params.Add("page", request.Site.Page)
		}
		params.Add("bidfloor", fmt.Sprintf("%v", imp.BidFloor))
		if request.Device != nil {
			if len(request.Device.IP) > 0 {
				params.Add("ip", request.Device.IP)
			}
			if len(request.Device.UA) > 0 {
				params.Add("ua", request.Device.UA)
			}
		}

		headers := http.Header{}
		// set imp id to be able to match it against bid
		headers.Set("PBS-IMP-ID", imp.ID)

		reqURL := adapter.URI + "&" + params.Encode()

		pubId := ""
		if request.Site != nil && request.Site.Publisher != nil {
			pubId = request.Site.Publisher.ID
		}

		fmt.Printf("colossus makerequests reqUrl: %s pubId=%s headers=%s\n", reqURL, pubId, headersToString(headers))
		reqData := adapters.RequestData{
			Method:  http.MethodGet,
			Uri:     reqURL,
			Headers: headers,
		}

		requests = append(requests, &reqData)
	}
	if len(errors) != 0 {
		return nil, errors
	}

	return requests, nil
}

func parseExt(imp *openrtb.Imp) (*openrtb_ext.ExtImpColossus, error) {
	var bidderExt adapters.ExtImpBidder

	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding extImpBidder, err: %s", imp.ID, err),
		}
	}

	impExt := openrtb_ext.ExtImpColossus{}
	err := json.Unmarshal(bidderExt.Bidder, &impExt)
	if err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding impExt, err: %s", imp.ID, err),
		}
	}

	return &impExt, nil
}

// MakeBids makes the bids
func (adapter *ColossusAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	fmt.Printf("colossus makebids start (id: %v | status: %v)\n", internalRequest.ID, response.StatusCode)
	var errs []error

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

	fmt.Printf("colossus makeBids response body (id: %v): %q\n", internalRequest.ID, string(response.Body))

	var vast adapters.VAST
	err := xml.Unmarshal(response.Body, &vast)
	if err != nil {
		return nil, []error{err}
	}

	if len(vast.Ads) == 0 {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("No Ads in VAST response"),
		}}
	}

	price, err := vast.Ads[0].GetPricing()
	if err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Couldn't parse CPM"),
		}}
	}
	if price == 0 {
		price = internalRequest.Imp[0].BidFloor
	}

	crID := vast.Ads[0].GetCreativeId()

	bidderResponse := adapters.NewBidderResponseWithBidsCapacity(1)
	bidderResponse.Bids = append(bidderResponse.Bids, &adapters.TypedBid{
		Bid: &openrtb.Bid{
			ID:    vast.Ads[0].ID,
			ImpID: externalRequest.Headers.Get("PBS-IMP-ID"),
			Price: price,
			AdM:   string(response.Body),
			CrID:  crID,
		},
		BidType: openrtb_ext.BidTypeAudio,
	})

	return bidderResponse, nil
}

func headersToString(m map[string][]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=%s", key, strings.Join(value, ","))
	}
	return b.String()
}
