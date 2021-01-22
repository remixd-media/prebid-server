package wideorbit

import (
	"bytes"
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

type WideOrbitAdapter struct {
	endpoint string
}

func (adapter *WideOrbitAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	numImps := len(request.Imp)

	requests := []*adapters.RequestData{}

	errors := []error{}

	for i := 0; i < numImps; i++ {
		imp := request.Imp[i]

		if imp.Audio == nil {
			continue
		}

		impExt, err := parseExt(&imp)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		params := url.Values{}

		isDeviceValid := request.Device != nil && request.Device.UA != "" && request.Device.IP != ""
		isPlacementIdValid := impExt.PlacementId != ""
		isAudioValid := imp.Audio != nil &&
			imp.Audio.MinBitrate > 0 && imp.Audio.MaxBitrate > 0 &&
			imp.Audio.MIMEs != nil && len(imp.Audio.MIMEs) > 0 &&
			imp.Audio.Protocols != nil && len(imp.Audio.Protocols) > 0

		if !isDeviceValid {
			errors = append(errors, fmt.Errorf("invalid input parameters - device not valid"))
			continue
		}
		if !isPlacementIdValid {
			errors = append(errors, fmt.Errorf("invalid input parameters - placementId not valid"))
			continue
		}
		if !isAudioValid {
			errors = append(errors, fmt.Errorf("invalid input parameters - audio not valid"))
			continue
		}

		//required
		params.Add("uas", request.Device.UA)
		params.Add("uip", request.Device.IP)
		params.Add("pId", impExt.PlacementId)
		params.Add("minbr", fmt.Sprintf("%d", imp.Audio.MinBitrate))
		params.Add("maxbr", fmt.Sprintf("%d", imp.Audio.MaxBitrate))
		params.Add("mimes", strings.Join(imp.Audio.MIMEs, ","))
		params.Add("spc", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(imp.Audio.Protocols)), ","), "[]"))

		//URL of the webpage/station the ad is played on. For in app requests, this should be aliased to the top level domain of the station’s website.
		params.Add("url", request.Site.Page)

		//optional
		params.Add("delay", "0") //0 is preroll - todo pass as param
		params.Add("lmt", "0")   //limit ad tracking
		if imp.Audio.MaxDuration > 0 {
			params.Add("maxdur", fmt.Sprintf("%d", imp.Audio.MaxDuration))
		}
		if imp.Audio.MinDuration > 0 {
			params.Add("mindur", fmt.Sprintf("%d", imp.Audio.MinDuration))
		}
		if len(request.BCat) > 0 {
			params.Add("bcat", strings.Join(request.BCat, ","))
		}

		if request.User != nil {

			if request.User.BuyerUID != "" {
				params.Add("uid", request.User.BuyerUID)
			}

			if request.User.Gender != "" {
				params.Add("a_gender", request.User.Gender) //M or F
			}

			if request.User.Yob > 0 {
				params.Add("a_age", fmt.Sprintf("%d", time.Now().Year()-int(request.User.Yob)))
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

		if request.Device != nil {
			if request.Device.Geo != nil {
				params.Add("lat", fmt.Sprintf("%.4f", request.Device.Geo.Lat))
				params.Add("lon", fmt.Sprintf("%.4f", request.Device.Geo.Lon))
			}
		}

		if request.Site != nil {
			if request.Site.Ref != "" {
				params.Add("ref", request.Site.Ref)
			}
		}

		headers := http.Header{}

		//wideorbit requires accept language parameter to be set in the following way (required)
		if request.Device != nil && request.Device.Language != "" {
			headers.Set("Accept-Language", "HTTP_ACCEPT_LANGUAGE="+request.Device.Language)
		} else {
			headers.Set("Accept-Language", "HTTP_ACCEPT_LANGUAGE=en")
		}

		// set imp id to be able to match it against bid
		headers.Set("PBS-IMP-ID", imp.ID)

		reqURL := adapter.endpoint + "&" + params.Encode()
		fmt.Printf("wideorbit makerequests reqUrl: %s headers=%s\n", reqURL, headersToString(headers))
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

func parseExt(imp *openrtb.Imp) (*openrtb_ext.ExtImpWideOrbit, error) {
	var bidderExt adapters.ExtImpBidder

	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding extImpBidder, err: %s", imp.ID, err),
		}
	}

	impExt := openrtb_ext.ExtImpWideOrbit{}
	err := json.Unmarshal(bidderExt.Bidder, &impExt)
	if err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding impExt, err: %s", imp.ID, err),
		}
	}

	return &impExt, nil
}

func (adapter *WideOrbitAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
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
		BidType: openrtb_ext.BidTypeAudio,
	})

	return bidderResponse, nil
}

// Adding header fields to request header
func addHeaderIfNonEmpty(headers http.Header, headerName string, headerValue string) {
	if len(headerValue) > 0 {
		headers.Add(headerName, headerValue)
	}
}

func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &WideOrbitAdapter{
		endpoint: config.Endpoint,
	}
	return bidder, nil
}


func headersToString(m map[string][]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=%s", key, strings.Join(value, ","))
	}
	return b.String()
}
