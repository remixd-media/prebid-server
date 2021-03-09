package adswizz

import (
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/prebid/prebid-server/config"
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
	endpoint string
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
		cb := fmt.Sprintf("%x", md5.Sum([]byte(request.ID+"-"+imp.ID)))[:8]
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

		if request.Device != nil && request.Device.Geo != nil {
			params.Add("lat", fmt.Sprintf("%.4f", request.Device.Geo.Lat))
			params.Add("lon", fmt.Sprintf("%.4f", request.Device.Geo.Lon))

			if request.Device.Geo.Country != "" {
				params.Add("aw_0_azn.pcountry", request.Device.Geo.Country)
			}
		}

		if impExt.PName != "" {
			params.Add("aw_0_azn.pname", impExt.PName)
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
			if request.Site.Content != nil && len(request.Site.Content.Cat) != 0 {
				params.Add("cat_include", strings.Join(request.Site.Content.Cat, ","))
			}
		}

		if imp.Audio != nil && imp.Audio.Feed == openrtb.FeedTypePodcast {
			params.Add("aw_0_azn.ptype", "Podcast")
		}

		reqURL := adapter.endpoint + "/" + impExt.Alias + "?" + params.Encode()
		fmt.Printf("adswizz makerequests reqUrl: %s\n", reqURL)
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

	price, err := strconv.ParseFloat(vast.Ads[0].InLine.Pricing, 64)
	if err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Couldn't parse CPM"),
		}}
	}
	if price == 0 {
		price = 0.1
	}

	var crID string
	var duration int

	if len(vast.Ads[0].InLine.Creatives.Creative) > 0 {
		creative := vast.Ads[0].InLine.Creatives.Creative[0]

		crID = creative.ID
		duration = adapters.ParseDuration(vast.Ads[0].InLine.Creatives.Creative[0].Linear.Duration)
	}

	bidderResponse := adapters.NewBidderResponseWithBidsCapacity(1)
	bidderResponse.Bids = append(bidderResponse.Bids, &adapters.TypedBid{
		Bid: &openrtb.Bid{
			ID:    vast.Ads[0].ID,
			ImpID: externalRequest.Headers.Get("PBS-IMP-ID"),
			Price: price,
			AdM:   string(response.Body),
			CrID:  crID,
		},
		BidType: openrtb_ext.BidTypeVideo,
		BidVideo: &openrtb_ext.ExtBidPrebidVideo{
			Duration: duration,
		},
	})

	return bidderResponse, nil
}

func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &AdsWizzAdapter{
		endpoint: config.Endpoint,
	}
	return bidder, nil
}
