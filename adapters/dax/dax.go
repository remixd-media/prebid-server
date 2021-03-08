package dax

import (
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

type DaxAdapter struct {
	endpoint string
}

func (adapter *DaxAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	numImps := len(request.Imp)

	requests := []*adapters.RequestData{}

	errors := []error{}

	for i := 0; i < numImps; i++ {
		imp := request.Imp[i]

		if imp.Audio == nil {
			continue
		}

		_, err := parseExt(&imp) //it may be used in future
		if err != nil {
			errors = append(errors, err)
			continue
		}

		params := url.Values{}

		//preroll
		sd := "0"
		if imp.Audio.StartDelay != nil {
			delay := *imp.Audio.StartDelay
			if delay == openrtb.StartDelayGenericMidRoll || delay > 0 {
				//midroll
				sd = "-1"
			} else if delay == openrtb.StartDelayGenericPostRoll {
				//postroll
				sd = "-2"
			}
		}

		params.Add("sd", sd)

		if imp.Audio.Feed == openrtb.FeedTypePodcast {
			params.Add("feed_type", "3")
		}

		if imp.Audio.MaxDuration > 0 {
			params.Add("dur_max", fmt.Sprintf("%d", imp.Audio.MaxDuration*1000))
		}

		if imp.Audio.MinDuration > 0 {
			params.Add("dur_min", fmt.Sprintf("%d", imp.Audio.MinBitrate*1000))
		}

		if len(request.BCat) > 0 {
			params.Add("bcat", strings.Join(request.BCat, ";"))
		}

		if request.User != nil {

			if request.User.BuyerUID != "" {
				params.Add("dax_listenerId", request.User.BuyerUID)
			}

			if request.User.Gender == "M" {
				params.Add("gender", "male")
			} else if request.User.Gender == "F" {
				params.Add("gender", "female")
			}

			if request.User.Yob > 0 {
				params.Add("age", fmt.Sprintf("%d", time.Now().Year()-int(request.User.Yob)))
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
			if request.Site.Content != nil {
				content := request.Site.Content
				if len(content.Cat) != 0 {
					params.Add("cat", strings.Join(content.Cat, ";"))
				}
				if content.Genre != "" {
					params.Add("genre", content.Genre)
				}
			}

		}

		reqURL := adapter.endpoint + "&" + params.Encode()
		fmt.Printf("dax makerequests reqUrl: %s\n", reqURL)
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

func parseExt(imp *openrtb.Imp) (*openrtb_ext.ExtImpDax, error) {
	var bidderExt adapters.ExtImpBidder

	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding extImpBidder, err: %s", imp.ID, err),
		}
	}

	impExt := openrtb_ext.ExtImpDax{}
	err := json.Unmarshal(bidderExt.Bidder, &impExt)
	if err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding impExt, err: %s", imp.ID, err),
		}
	}

	return &impExt, nil
}

func (adapter *DaxAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
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

	if len(vast.Ads[0].InLine.Creatives.Creative) > 0 {
		creative := vast.Ads[0].InLine.Creatives.Creative[0]

		crID = creative.ID
		//duration = adapters.ParseDuration(vast.Ads[0].InLine.Creatives.Creative[0].Linear.Duration)
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
		BidType: openrtb_ext.BidTypeAudio, //not sure why video is used by previous dev, probably due to limitation in pbs
	})

	return bidderResponse, nil
}

func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &DaxAdapter{
		endpoint: config.Endpoint,
	}
	return bidder, nil
}
