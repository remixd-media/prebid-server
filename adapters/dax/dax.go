package dax

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"log"
	"net/http"
	"strconv"
)

type DaxAdapter struct {
	endpoint string
}

func (adapter *DaxAdapter) MakeRequests(request *openrtb.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {

	impressions := request.Imp
	adapterRequests := make([]*adapters.RequestData, 0, len(impressions))
	errs := make([]error, 0, len(impressions))

	if len(impressions) == 0 {
		errs = append(errs, &errortypes.BadInput{
			Message: fmt.Sprintf("Invalid BidRequest. No valid imp."),
		})
		return nil, errs
	}

	for _, impression := range impressions {
		headers := http.Header{}
		headers.Add("Content-Type", "application/json;charset=utf-8")
		headers.Add("Accept", "application/json")
		// Need to process audio impressions
		if impression.Audio == nil {
			continue
		}

		if impression.Video != nil || impression.Banner != nil {
			errs = append(errs, &errortypes.BadInput{
				Message: fmt.Sprintf("Ignoring impression id=%s, Only Audio is supported.", impression.ID),
			})
			continue
		}

		var bidderExt adapters.ExtImpBidder
		if err := json.Unmarshal(impression.Ext, &bidderExt); err != nil {
			errs = append(errs, &errortypes.BadInput{
				Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding extImpBidder, err: %s.", impression.ID, err),
			})
			continue
		}

		// update headers
		// set imp id to be able to match it against bid
		headers.Set("PBS-IMP-ID", impression.ID)
		if request.Device != nil {
			if request.Device.UA != "" {
				headers.Set("User-Agent", request.Device.UA)
			}
			if request.Device.IP != "" {
				headers.Set("X-Forwarded-For", request.Device.IP)
			}
		}
		if request.Site != nil {
			if request.Site.Page != "" {
				headers.Set("Referer", request.Site.Page)
			}
		}

		request.Imp = []openrtb.Imp{impression}
		body, err := json.Marshal(request)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		adapterRequests = append(adapterRequests, &adapters.RequestData{
			Method:  http.MethodPost,
			Uri:     adapter.endpoint,
			Body:    body,
			Headers: headers,
		})
		log.Printf("Dax req Body: %v\n", string(body))
	}

	request.Imp = impressions
	return adapterRequests, errs
}

func (adapter *DaxAdapter) MakeBids(request *openrtb.BidRequest, requestData *adapters.RequestData, responseData *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	var errs []error

	switch responseData.StatusCode {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusBadRequest:
		return nil, []error{&errortypes.BadInput{
			Message: "Unexpected status code: " + strconv.Itoa(responseData.StatusCode),
		}}
	case http.StatusOK:
		break
	default:
		return nil, []error{&errortypes.BadServerResponse{
			Message: "unexpected status code: " + strconv.Itoa(responseData.StatusCode),
		}}
	}

	var bidResponse openrtb.BidResponse
	err := json.Unmarshal(responseData.Body, &bidResponse)
	if err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: err.Error(),
		}}
	}

	response := adapters.NewBidderResponseWithBidsCapacity(len(request.Imp))
	impID := requestData.Headers.Get("PBS-IMP-ID")
	for _, seatBid := range bidResponse.SeatBid {
		for i := range seatBid.Bid {
			var vast adapters.VAST
			err = xml.Unmarshal([]byte(seatBid.Bid[i].AdM), &vast)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if len(vast.Ads) == 0 {
				errs = append(errs, &errortypes.BadServerResponse{
					Message: fmt.Sprintf("No Ads in VAST response"),
				})
				continue
			}

			if vast.Ads[0].InLine.Pricing == "" {
				vast.Ads[0].InLine.Pricing = "0"
			}
			var price float64
			price, err = strconv.ParseFloat(vast.Ads[0].InLine.Pricing, 64)
			if err != nil {
				errs = append(errs, &errortypes.BadServerResponse{
					Message: fmt.Sprintf("Couldn't parse CPM"),
				})
				continue
			}
			if price == 0 {
				price = 0.1
			}

			var crID string

			if len(vast.Ads[0].InLine.Creatives.Creative) > 0 {
				creative := vast.Ads[0].InLine.Creatives.Creative[0]

				crID = creative.ID
			}
			newBid := &adapters.TypedBid{
				Bid:     &seatBid.Bid[i],
				BidType: openrtb_ext.BidTypeAudio,
			}
			newBid.Bid.ID = vast.Ads[0].ID
			newBid.Bid.ImpID = impID
			newBid.Bid.Price = price
			newBid.Bid.CrID = crID
			response.Bids = append(response.Bids, newBid)
		}
	}

	return response, errs
}

// Builder builds a new instance of the DaxAdapter adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &DaxAdapter{
		endpoint: config.Endpoint,
	}
	return bidder, nil
}
