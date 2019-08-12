package tritondigital

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type TritonDigitalAdapter struct {
	URI string
}

func (adapter *TritonDigitalAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	numImps := len(request.Imp)

	requests := []*adapters.RequestData{}

	errors := []error{}

	for i := 0; i < numImps; i++ {
		imp := request.Imp[i]

		if imp.Video == nil {
			continue
		}

		impExt, err := parseExt(&imp)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		params := url.Values{}
		params.Add("version", "1.6.9")

		var adType string
		if imp.Video.StartDelay != nil && *imp.Video.StartDelay != 0 {
			adType = "midroll"

		} else {
			adType = "preroll"
		}
		params.Add("type", adType)

		if imp.Video.MinDuration > 0 {
			params.Add("mindur", fmt.Sprintf("%d", imp.Video.MinDuration))
		}
		if imp.Video.MaxDuration > 0 {
			params.Add("maxdur", fmt.Sprintf("%d", imp.Video.MaxDuration))
		}

		if imp.Video.MinBitRate > 0 {
			params.Add("minbr", fmt.Sprintf("%d", imp.Video.MinBitRate))
		}
		if imp.Video.MaxBitRate > 0 {
			params.Add("maxbr", fmt.Sprintf("%d", imp.Video.MaxBitRate))
		}

		if len(request.BCat) > 0 {
			params.Add("iab-categories-to-exclude", strings.Join(request.BCat, ","))
		}

		if request.User != nil {
			if request.User.Yob > 0 {
				params.Add("yob", fmt.Sprintf("%d", request.User.Yob))
			}

			if request.User.Gender == "M" {
				params.Add("gender", "m")
			} else if request.User.Gender == "F" {
				params.Add("gender", "f")
			}

			if request.User.Geo != nil {
				if request.User.Geo.Lat != 0 {
					params.Add("lat", fmt.Sprintf("%f", request.User.Geo.Lat))
				}
				if request.User.Geo.Lon != 0 {
					params.Add("long", fmt.Sprintf("%f", request.User.Geo.Lon))
				}
				if request.User.Geo.ZIP != "" {
					params.Add("postalcode", request.User.Geo.ZIP)
				}
				if request.User.Geo.Country != "" {
					params.Add("country", request.User.Geo.Country)
				}
			}
		}

		params.Add("at", "audio")
		params.Add("fmt", "vast")
		params.Add("banners", impExt.Banners)
		params.Add("stid", impExt.StID)

		headers := http.Header{}
		// set imp id to be able to match it against bid
		headers.Set("PBS-IMP-ID", imp.ID)

		if request.Device != nil && request.Device.UA != "" {
			headers.Set("User-Agent", request.Device.UA)
		}

		reqData := adapters.RequestData{
			Method:  http.MethodGet,
			Uri:     adapter.URI + "?" + params.Encode(),
			Headers: headers,
		}

		requests = append(requests, &reqData)
	}

	if len(errors) != 0 {
		return nil, errors
	}

	return requests, nil
}

func parseExt(imp *openrtb.Imp) (*openrtb_ext.ExtImpTritonDigital, error) {
	var bidderExt adapters.ExtImpBidder

	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding extImpBidder, err: %s", imp.ID, err),
		}
	}

	impExt := openrtb_ext.ExtImpTritonDigital{}
	err := json.Unmarshal(bidderExt.Bidder, &impExt)
	if err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding impExt, err: %s", imp.ID, err),
		}
	}

	return &impExt, nil
}

type VAST struct {
	Ads []Ad `xml:"Ad"`
}

type Ad struct {
	ID string `xml:"id,attr,omitempty"`
}

func (adapter *TritonDigitalAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
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

	var vast VAST
	err := xml.Unmarshal(response.Body, &vast)
	if err != nil {
		return nil, []error{err}
	}

	if len(vast.Ads) == 0 {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("No Ads in VAST response"),
		}}
	}

	bidderResponse := adapters.NewBidderResponseWithBidsCapacity(1)
	bidderResponse.Bids = append(bidderResponse.Bids, &adapters.TypedBid{
		Bid: &openrtb.Bid{
			ID:    vast.Ads[0].ID,
			ImpID: externalRequest.Headers.Get("PBS-IMP-ID"),
			Price: 1.00,
			AdM:   string(response.Body),
			CrID:  vast.Ads[0].ID,
		},
		BidType: openrtb_ext.BidTypeVideo,
	})

	return bidderResponse, nil
}

func NewTritonDigitalBidder(endpoint string) *TritonDigitalAdapter {
	return &TritonDigitalAdapter{
		URI: endpoint,
	}
}
