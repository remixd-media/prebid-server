package tritondigital

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
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
		params.Add("version", "1.7.4")

		var adType string
		/*	if imp.Video.StartDelay != nil && *imp.Video.StartDelay != 0 {
			adType = "midroll"
		} else {*/
		adType = "preroll"
		//		}
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

		if request.Site != nil {
			if request.Site.Page != "" {
				params.Add("referrer", request.Site.Page)
				params.Add("site-url", request.Site.Page)
			}
		}

		if request.User != nil {
			if request.User.BuyerUID != "" {
				tritonParterUIDMap := map[string]string{}

				err := json.Unmarshal([]byte(request.User.BuyerUID), &tritonParterUIDMap)
				if err != nil {
					fmt.Printf("tritondigital makerequests parse buyeruid: %s err: %v\n", request.User.BuyerUID, err)
				} else {
					for k, v := range tritonParterUIDMap {
						if strings.HasSuffix(k, "-uid") {
							params.Add(k, v)
						}
					}
				}

				// old usage from cookie sync in 1.6.9
				//params.Add("lsid", request.User.BuyerUID)
			}

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

		if request.Device != nil {
			if request.Device.UA != "" {
				params.Add("ua", request.Device.UA)
			}
			if request.Device.IP != "" {
				params.Add("ip", request.Device.IP)
			}
		}

		params.Add("at", "audio")
		params.Add("cntnr", "mp3")
		params.Add("acodec", "mp3")
		params.Add("fmt", "vast")
		params.Add("banners", impExt.Banners)
		params.Add("stid", impExt.StID)

		headers := http.Header{}
		// set imp id to be able to match it against bid
		headers.Set("PBS-IMP-ID", imp.ID)

		reqURL := adapter.URI + "?" + params.Encode()
		fmt.Printf("tritondigital makerequests reqUrl: %s\n", reqURL)

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

func (adapter *TritonDigitalAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	fmt.Printf("tritondigital makebids start (id: %v)\n", internalRequest.ID)
	defer fmt.Printf("tritondigital makebids (id: %v):\n", internalRequest.ID)

	if response.StatusCode == http.StatusNoContent {
		fmt.Printf("tritondigital makeBids no content (id: %v):\n", internalRequest.ID)
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		fmt.Printf("tritondigital makeBids err bad input (id: %v): http status %v\n", internalRequest.ID, response.StatusCode)
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Bad user input (id: %v): HTTP status %v", internalRequest.ID, response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		fmt.Printf("tritondigital makeBids err bad server response (id: %v): http status %v\n", internalRequest.ID, response.StatusCode)
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Bad server response: HTTP status %v", response.StatusCode),
		}}
	}

	fmt.Printf("tritondigital makeBids response body (id: %v): %q\n", internalRequest.ID, string(response.Body))

	var vast adapters.VAST
	err := xml.Unmarshal(response.Body, &vast)
	if err != nil {
		fmt.Printf("tritondigital makeBids vast parse error (id: %v): %v\n", internalRequest.ID, err)
		return nil, []error{err}
	}

	if len(vast.Ads) == 0 {
		fmt.Printf("tritondigital makeBids no ad (id: %v)\n", internalRequest.ID)
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("No Ads in VAST response"),
		}}
	}

	price, err := strconv.ParseFloat(vast.Ads[0].InLine.Pricing, 64)
	if err != nil {
		fmt.Printf("tritondigital makeBids couldn't parse CPM (id: %v)\n", internalRequest.ID)
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Couldn't parse CPM"),
		}}
	}

	// adjust to net
	price = price * 0.8
	// round to 4 decimals
	price = math.Floor(price*10000) / 10000

	var crID string
	var duration int

	if len(vast.Ads[0].InLine.Creatives.Creative) > 0 {
		creative := vast.Ads[0].InLine.Creatives.Creative[0]

		crID = creative.ID
		duration = adapters.ParseDuration(creative.Linear.Duration)
	}

	adID := vast.Ads[0].ID

	if crID == "" {
		crID = adID // avoid having empty crID
	}

	adm := string(response.Body)
	adm = adapters.HidePricing(adm)

	bidderResponse := adapters.NewBidderResponseWithBidsCapacity(1)
	bidderResponse.Bids = append(bidderResponse.Bids, &adapters.TypedBid{
		Bid: &openrtb.Bid{
			ID:    vast.Ads[0].ID,
			ImpID: externalRequest.Headers.Get("PBS-IMP-ID"),
			Price: price,
			AdM:   adm,
			CrID:  crID,
		},
		BidType: openrtb_ext.BidTypeVideo,
		BidVideo: &openrtb_ext.ExtBidPrebidVideo{
			Duration: duration,
		},
	})

	return bidderResponse, nil
}

func NewTritonDigitalBidder(endpoint string) *TritonDigitalAdapter {
	return &TritonDigitalAdapter{
		URI: endpoint,
	}
}
