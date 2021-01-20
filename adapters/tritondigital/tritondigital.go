package tritondigital

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/prebid/prebid-server/config"
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
	endpoint string
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

			catsV1 := request.Site.Cat
			// convert iab categories to v2 ids
			catsV2 := ""
			for _, cat := range catsV1 {
				if catV2, ok := iabConvMap[cat]; ok {
					catsV2 += catV2 + ","
				}
			}
			catsV2 = strings.TrimSuffix(catsV2, ",")
			if catsV2 != "" {
				params.Add("iab-v2-cat", catsV2)
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
		params.Add("acodec", "mp3,aac_hev1,aac_hev2,aac_lc")
		params.Add("fmt", "vast")
		params.Add("banners", impExt.Banners)
		params.Add("stid", impExt.StID)

		headers := http.Header{}
		// set imp id to be able to match it against bid
		headers.Set("PBS-IMP-ID", imp.ID)

		reqURL := adapter.endpoint + "?" + params.Encode()
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

func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &TritonDigitalAdapter{
		endpoint:    config.Endpoint,
	}
	return bidder, nil
}


// iab conv map v1 => v2
var iabConvMap = map[string]string{
	`IAB19-1`:  `603`,
	`IAB6-1`:   `193`,
	`IAB5-2`:   `133`,
	`IAB20-1`:  `665`,
	`IAB20-2`:  `656`,
	`IAB23-2`:  `454`,
	`IAB3-2`:   `102`,
	`IAB20-3`:  `672`,
	`IAB8-5`:   `211`,
	`IAB8-18`:  `211`,
	`IAB7-4`:   `288`,
	`IAB7-5`:   `233`,
	`IAB17-12`: `484`,
	`IAB19-3`:  `608`,
	`IAB21-1`:  `442`,
	`IAB9-2`:   `248`,
	`IAB15-1`:  `456`,
	`IAB9-17`:  `265`,
	`IAB20-4`:  `658`,
	`IAB2-3`:   `30`,
	`IAB2-1`:   `32`,
	`IAB17-1`:  `518`,
	`IAB2-2`:   `34`,
	`IAB2`:     `1`,
	`IAB8-2`:   `215`,
	`IAB17-2`:  `545`,
	`IAB17-26`: `547`,
	`IAB9-3`:   `249`,
	`IAB18-1`:  `553`,
	`IAB20-5`:  `674`,
	`IAB15-2`:  `465`,
	`IAB3-3`:   `119`,
	`IAB16-2`:  `423`,
	`IAB9-4`:   `259`,
	`IAB9-5`:   `270`,
	`IAB18-2`:  `574`,
	`IAB17-4`:  `549`,
	`IAB7-33`:  `312`,
	`IAB1-1`:   `42`,
	`IAB17-5`:  `485`,
	`IAB23-3`:  `458`,
	`IAB20-6`:  `675`,
	`IAB3`:     `53`,
	`IAB20-7`:  `676`,
	`IAB19-5`:  `633`,
	`IAB20-9`:  `677`,
	`IAB9-6`:   `250`,
	`IAB17-6`:  `499`,
	`IAB2-4`:   `25`,
	`IAB9-7`:   `271`,
	`IAB4-11`:  `125`,
	`IAB4-1`:   `126`,
	`IAB4`:     `123`,
	`IAB16-3`:  `424`,
	`IAB2-5`:   `18`,
	`IAB17-7`:  `486`,
	`IAB15-3`:  `466`,
	`IAB23-5`:  `459`,
	`IAB9-9`:   `260`,
	`IAB2-22`:  `19`,
	`IAB17-8`:  `500`,
	`IAB9-10`:  `261`,
	`IAB5-5`:   `137`,
	`IAB9-11`:  `262`,
	`IAB2-21`:  `3`,
	`IAB19-2`:  `610`,
	`IAB19-8`:  `600`,
	`IAB19-9`:  `601`,
	`IAB3-5`:   `121`,
	`IAB9-15`:  `264`,
	`IAB2-6`:   `8`,
	`IAB2-7`:   `9`,
	`IAB22-2`:  `474`,
	`IAB17-9`:  `491`,
	`IAB2-8`:   `10`,
	`IAB20-12`: `678`,
	`IAB17-3`:  `492`,
	`IAB19-12`: `611`,
	`IAB14-1`:  `188`,
	`IAB6-3`:   `194`,
	`IAB7-17`:  `316`,
	`IAB19-13`: `612`,
	`IAB8-8`:   `217`,
	`IAB9-1`:   `205`,
	`IAB19-22`: `613`,
	`IAB8-9`:   `218`,
	`IAB14-2`:  `189`,
	`IAB16-4`:  `425`,
	`IAB9-12`:  `251`,
	`IAB5`:     `132`,
	`IAB6-9`:   `190`,
	`IAB19-15`: `623`,
	`IAB17-17`: `496`,
	`IAB20-14`: `659`,
	`IAB6`:     `186`,
	`IAB20-26`: `666`,
	`IAB17-10`: `510`,
	`IAB13-4`:  `396`,
	`IAB1-3`:   `201`,
	`IAB16-1`:  `426`,
	`IAB17-11`: `511`,
	`IAB17-13`: `511`,
	`IAB17-32`: `511`,
	`IAB7-1`:   `225`,
	`IAB8`:     `210`,
	`IAB8-10`:  `219`,
	`IAB9-13`:  `266`,
	`IAB10-4`:  `275`,
	`IAB9-14`:  `273`,
	`IAB15-8`:  `469`,
	`IAB15-4`:  `470`,
	`IAB17-15`: `512`,
	`IAB3-7`:   `77`,
	`IAB19-16`: `614`,
	`IAB3-8`:   `78`,
	`IAB2-10`:  `22`,
	`IAB2-12`:  `22`,
	`IAB2-11`:  `11`,
	`IAB8-12`:  `221`,
	`IAB7-35`:  `223`,
	`IAB7-38`:  `223`,
	`IAB7-24`:  `296`,
	`IAB13-5`:  `411`,
	`IAB7-25`:  `234`,
	`IAB23-6`:  `460`,
	`IAB9`:     `239`,
	`IAB7-26`:  `235`,
	`IAB10`:    `274`,
	`IAB10-1`:  `278`,
	`IAB10-2`:  `279`,
	`IAB19-17`: `634`,
	`IAB10-5`:  `280`,
	`IAB5-10`:  `145`,
	`IAB5-11`:  `146`,
	`IAB20-17`: `667`,
	`IAB17-16`: `497`,
	`IAB20-18`: `668`,
	`IAB3-9`:   `55`,
	`IAB1-4`:   `440`,
	`IAB17-18`: `514`,
	`IAB17-27`: `515`,
	`IAB10-3`:  `282`,
	`IAB19-25`: `618`,
	`IAB17-19`: `516`,
	`IAB13-6`:  `398`,
	`IAB10-7`:  `283`,
	`IAB12-1`:  `382`,
	`IAB19-24`: `624`,
	`IAB6-4`:   `195`,
	`IAB23-7`:  `461`,
	`IAB9-19`:  `252`,
	`IAB4-4`:   `128`,
	`IAB4-5`:   `127`,
	`IAB23-8`:  `462`,
	`IAB10-8`:  `284`,
	`IAB5-8`:   `147`,
	`IAB16-5`:  `427`,
	`IAB11-2`:  `383`,
	`IAB12-3`:  `384`,
	`IAB3-10`:  `57`,
	`IAB2-13`:  `23`,
	`IAB9-20`:  `241`,
	`IAB3-1`:   `58`,
	`IAB3-11`:  `58`,
	`IAB14-4`:  `191`,
	`IAB17-20`: `520`,
	`IAB7-31`:  `228`,
	`IAB7-37`:  `301`,
	`IAB3-12`:  `107`,
	`IAB2-14`:  `13`,
	`IAB17-25`: `519`,
	`IAB2-15`:  `27`,
	`IAB1-5`:   `324`,
	`IAB1-6`:   `338`,
	`IAB9-16`:  `243`,
	`IAB13-8`:  `412`,
	`IAB12-2`:  `385`,
	`IAB9-21`:  `253`,
	`IAB8-6`:   `222`,
	`IAB7-32`:  `229`,
	`IAB2-16`:  `14`,
	`IAB17-23`: `521`,
	`IAB13-9`:  `413`,
	`IAB17-24`: `501`,
	`IAB9-22`:  `254`,
	`IAB15-5`:  `244`,
	`IAB6-2`:   `196`,
	`IAB6-5`:   `197`,
	`IAB6-6`:   `198`,
	`IAB2-17`:  `24`,
	`IAB13-2`:  `405`,
	`IAB13`:    `391`,
	`IAB13-7`:  `410`,
	`IAB13-12`: `415`,
	`IAB16`:    `422`,
	`IAB9-23`:  `255`,
	`IAB7-36`:  `236`,
	`IAB15-6`:  `471`,
	`IAB2-18`:  `15`,
	`IAB11-4`:  `386`,
	`IAB1-2`:   `432`,
	`IAB5-9`:   `139`,
	`IAB6-7`:   `305`,
	`IAB5-12`:  `149`,
	`IAB5-13`:  `134`,
	`IAB19-4`:  `631`,
	`IAB19-19`: `631`,
	`IAB19-20`: `631`,
	`IAB19-32`: `631`,
	`IAB9-24`:  `245`,
	`IAB21`:    `441`,
	`IAB21-3`:  `451`,
	`IAB23`:    `453`,
	`IAB10-9`:  `276`,
	`IAB4-9`:   `130`,
	`IAB16-6`:  `429`,
	`IAB4-6`:   `129`,
	`IAB13-10`: `416`,
	`IAB2-19`:  `28`,
	`IAB17-28`: `525`,
	`IAB9-25`:  `272`,
	`IAB17-29`: `527`,
	`IAB17-30`: `227`,
	`IAB17-31`: `530`,
	`IAB22-1`:  `481`,
	`IAB15`:    `464`,
	`IAB9-26`:  `246`,
	`IAB9-27`:  `256`,
	`IAB9-28`:  `267`,
	`IAB17-33`: `502`,
	`IAB19-35`: `627`,
	`IAB2-20`:  `4`,
	`IAB7-39`:  `307`,
	`IAB19-30`: `605`,
	`IAB22`:    `473`,
	`IAB17-34`: `503`,
	`IAB17-35`: `531`,
	`IAB7-19`:  `309`,
	`IAB7-40`:  `310`,
	`IAB19-6`:  `635`,
	`IAB7-41`:  `237`,
	`IAB17-36`: `504`,
	`IAB17-44`: `533`,
	`IAB20-23`: `662`,
	`IAB15-7`:  `472`,
	`IAB20-24`: `671`,
	`IAB5-14`:  `136`,
	`IAB6-8`:   `199`,
	`IAB17`:    `483`,
	`IAB9-29`:  `263`,
	`IAB2-23`:  `5`,
	`IAB13-11`: `414`,
	`IAB4-3`:   `395`,
	`IAB18`:    `552`,
	`IAB7-42`:  `311`,
	`IAB17-37`: `505`,
	`IAB17-38`: `537`,
	`IAB17-39`: `538`,
	`IAB19-26`: `636`,
	`IAB19`:    `596`,
	`IAB1-7`:   `640`,
	`IAB17-40`: `539`,
	`IAB7-43`:  `293`,
	`IAB20`:    `653`,
	`IAB8-16`:  `212`,
	`IAB8-17`:  `213`,
	`IAB16-7`:  `430`,
	`IAB9-30`:  `680`,
	`IAB17-41`: `541`,
	`IAB17-42`: `542`,
	`IAB17-43`: `506`,
	`IAB15-10`: `390`,
	`IAB19-23`: `607`,
	`IAB19-34`: `629`,
	`IAB14-7`:  `165`,
	`IAB7-44`:  `231`,
	`IAB7-45`:  `238`,
	`IAB9-31`:  `257`,
	`IAB5-1`:   `135`,
	`IAB7-2`:   `301`,
	`IAB18-6`:  `561,580`,
	`IAB7-3`:   `297`,
	`IAB23-1`:  `453`,
	`IAB8-1`:   `214`,
	`IAB7-6`:   `312`,
	`IAB7-7`:   `300`,
	`IAB7-8`:   `301`,
	`IAB13-1`:  `410`,
	`IAB7-9`:   `301`,
	`IAB15-9`:  `465`,
	`IAB7-10`:  `313`,
	`IAB3-4`:   `602`,
	`IAB20-8`:  `660`,
	`IAB8-3`:   `214`,
	`IAB20-10`: `660`,
	`IAB7-11`:  `314`,
	`IAB20-11`: `660`,
	`IAB23-4`:  `459`,
	`IAB9-8`:   `270`,
	`IAB8-4`:   `214`,
	`IAB7-12`:  `306`,
	`IAB7-13`:  `313`,
	`IAB7-14`:  `287`,
	`IAB18-5`:  `566,582,576,575`,
	`IAB7-15`:  `315`,
	`IAB8-7`:   `214`,
	`IAB19-11`: `616`,
	`IAB7-16`:  `289`,
	`IAB7-18`:  `301`,
	`IAB7-20`:  `290`,
	`IAB20-13`: `659`,
	`IAB7-21`:  `313`,
	`IAB18-3`:  `560,579`,
	`IAB13-3`:  `52`,
	`IAB20-15`: `659`,
	`IAB8-11`:  `214`,
	`IAB7-22`:  `318`,
	`IAB20-16`: `659`,
	`IAB7-23`:  `313`,
	`IAB7`:     `286,223`,
	`IAB10-6`:  `634`,
	`IAB7-27`:  `318`,
	`IAB11-1`:  `388`,
	`IAB7-29`:  `318`,
	`IAB7-30`:  `304`,
	`IAB19-18`: `619`,
	`IAB8-13`:  `214`,
	`IAB20-19`: `659`,
	`IAB20-20`: `657`,
	`IAB8-14`:  `214`,
	`IAB18-4`:  `581,565`,
	`IAB23-9`:  `459`,
	`IAB11`:    `379`,
	`IAB8-15`:  `214`,
	`IAB20-21`: `662`,
	`IAB17-21`: `492`,
	`IAB17-22`: `518`,
	`IAB12`:    `379`,
	`IAB23-10`: `453`,
	`IAB7-34`:  `301`,
	`IAB4-8`:   `395`,
	`IAB26-3`:  `608`,
	`IAB20-25`: `151`,
	`IAB20-27`: `659`,
}
