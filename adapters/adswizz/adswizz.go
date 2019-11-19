package adswizz

import (
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type AdsWizzAdapter struct {
	URI string
}

func (adapter *AdsWizzAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
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

		// id md5 as cachebuster
		cb := fmt.Sprintf("%x", md5.Sum([]byte(imp.ID)))[:8]
		params.Add("cb", cb)

		if imp.Video.MaxDuration > 0 {
			params.Add("duration", fmt.Sprintf("%d", imp.Video.MaxDuration*1000))
		}

		if len(request.BCat) > 0 {
			params.Add("cat_exclude", strings.Join(request.BCat, ","))
		}

		if request.User != nil {

			if request.User.BuyerUID != "" {
				params.Add("listenerId", request.User.BuyerUID)
			}

			if request.User.Gender == "M" {
				params.Add("aw_0_1st.gender", "male")
			} else if request.User.Gender == "F" {
				params.Add("aw_0_1st.gender", "female")
			}

			if request.User.Yob > 0 {
				params.Add("aw_0_1st.age", fmt.Sprintf("%d", time.Now().Year()-int(request.User.Yob)))
			}

			if request.User.Geo != nil {
				if request.User.Geo.Lat != 0 {
					params.Add("lat", fmt.Sprintf("%f", request.User.Geo.Lat))
				}
				if request.User.Geo.Lon != 0 {
					params.Add("lon", fmt.Sprintf("%f", request.User.Geo.Lon))
				}
			}
		}

		headers := http.Header{}
		// set imp id to be able to match it against bid
		headers.Set("PBS-IMP-ID", imp.ID)

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

		reqData := adapters.RequestData{
			Method:  http.MethodGet,
			Uri:     adapter.URI + "/" + impExt.Alias + "?" + params.Encode(),
			Headers: headers,
		}

		requests = append(requests, &reqData)
	}

	if len(errors) != 0 {
		return nil, errors
	}

	return requests, nil
}

func parseExt(imp *openrtb.Imp) (*openrtb_ext.ExtImpAdsWizz, error) {
	var bidderExt adapters.ExtImpBidder

	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding extImpBidder, err: %s", imp.ID, err),
		}
	}

	impExt := openrtb_ext.ExtImpAdsWizz{}
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
	ID     string `xml:"id,attr"`
	InLine InLine `xml:"InLine"`
}

type Creative struct {
	ID string `xml:"id,attr"`
}

type Creatives struct {
	Creative []Creative `xml:"Creative"`
}

type InLine struct {
	Pricing   string    `xml:"Pricing"`
	Creatives Creatives `xml:"Creatives"`
}

func (adapter *AdsWizzAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
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

	price, err := strconv.ParseFloat(vast.Ads[0].InLine.Pricing, 64)
	if err != nil {
		price = 0
	}

	if internalRequest.Test == 1 {
		price = 4.5 // testing value
	}

	bidderResponse := adapters.NewBidderResponseWithBidsCapacity(1)
	bidderResponse.Bids = append(bidderResponse.Bids, &adapters.TypedBid{
		Bid: &openrtb.Bid{
			ID:    vast.Ads[0].ID,
			ImpID: externalRequest.Headers.Get("PBS-IMP-ID"),
			Price: price,
			AdM:   string(response.Body),
			CrID:  vast.Ads[0].InLine.Creatives.Creative[0].ID,
		},
		BidType: openrtb_ext.BidTypeVideo,
	})

	return bidderResponse, nil
}

func NewAdsWizzBidder(endpoint string) *AdsWizzAdapter {
	return &AdsWizzAdapter{
		URI: endpoint,
	}
}
