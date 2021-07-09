package voxnest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"net/http"
	"net/url"
	"strings"
)

type VoxnestAdapter struct {
	endpoint string
}

func (adapter *VoxnestAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	numImps := len(request.Imp)
	requests := []*adapters.RequestData{}
	errors := []error{}

	for i := 0; i < numImps; i++ {
		imp := request.Imp[i]

		if imp.Audio == nil {
			continue
		}

		_, err := parseExt(&imp)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		params := url.Values{}

		delay := "preroll"
		if imp.Audio.StartDelay != nil {
			if *imp.Audio.StartDelay > 0 || *imp.Audio.StartDelay == openrtb.StartDelayGenericMidRoll {
				delay = "midroll"
			} else if *imp.Audio.StartDelay == openrtb.StartDelayGenericPostRoll {
				delay = "postroll"
			}
		}
		params.Add("ads_type", delay)

		showId, episodeId := adapter.getShowAndEpisodeId(request, imp)
		params.Add("content_show_id", showId)
		params.Add("content_episode_id", episodeId)

		var contentGenres []string
		if request.Site != nil && request.Site.Content != nil {
			contentGenres = append(contentGenres, request.Site.Content.Genre)
		}

		params.Add("content_categories", adapter.encodeStrSlice(contentGenres))
		params.Add("blocked_iab_categories", adapter.encodeStrSlice(request.BCat))

		headers := http.Header{}
		if request.Device != nil {
			if len(request.Device.UA) > 0 {
				headers.Add("User-Agent", request.Device.UA)
			}
			if len(request.Device.IP) > 0 {
				headers.Add("X-Forwarded-For", request.Device.IP)
			}
		}

		// set imp id to be able to match it against bid
		headers.Set("PBS-IMP-ID", imp.ID)

		reqURL := adapter.endpoint + "?" + params.Encode()

		pubId := ""
		if request.Site != nil && request.Site.Publisher != nil {
			pubId = request.Site.Publisher.ID
		}

		fmt.Printf("voxnest makerequests reqUrl: %s pubId=%s headers=%s\n", reqURL, pubId, headersToString(headers))
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

func (adapter *VoxnestAdapter) getShowAndEpisodeId(request *openrtb.BidRequest, imp openrtb.Imp) (string, string) {
	showId, episodeId := "remixd.com", "episode_id"
	if request.Site == nil {
		return showId, episodeId
	}

	entirePage := request.Site.Page
	u, err := url.Parse(request.Site.Page)
	if err == nil {
		showId, episodeId = u.Host, u.Path
		if episodeId == "" {
			episodeId = "/"
		}
		entirePage = u.Host + u.Path
	} else {
		showId = entirePage
	}

	if imp.Audio.Feed == openrtb.FeedTypePodcast && request.Site.Content != nil && len(request.Site.Content.ID) > 0 {
		return request.Site.Content.ID, entirePage
	}

	return showId, episodeId
}

func parseExt(imp *openrtb.Imp) (*openrtb_ext.ExtImpVoxnest, error) {
	var bidderExt adapters.ExtImpBidder

	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding extImpBidder, err: %s", imp.ID, err),
		}
	}

	impExt := openrtb_ext.ExtImpVoxnest{}
	err := json.Unmarshal(bidderExt.Bidder, &impExt)
	if err != nil {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding impExt, err: %s", imp.ID, err),
		}
	}

	return &impExt, nil
}

func (adapter *VoxnestAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	fmt.Printf("voxnest makebids start (id: %v | status: %v)\n", internalRequest.ID, response.StatusCode)
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

	fmt.Printf("voxnest makeBids response body (id: %v): %q\n", internalRequest.ID, string(response.Body))

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

	price := internalRequest.Imp[0].BidFloor
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

func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &VoxnestAdapter{
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

func (adapter *VoxnestAdapter) encodeStrSlice(strSlice []string) string {
	var builder strings.Builder
	builder.WriteString("[")
	for j, s := range strSlice {
		if j != 0 {
			builder.WriteString(",")
		}
		builder.WriteString("\"")
		builder.WriteString(s)
		builder.WriteString("\"")
	}
	builder.WriteString("]")
	return builder.String()
}
