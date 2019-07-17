package remixd

import (
	"encoding/json"
	"fmt"

	"github.com/mxmCherry/openrtb"
)

func GetBiddersData(orig *openrtb.BidRequest) (json.RawMessage, error) {
	// placeholder
	return json.RawMessage(`{
		"appnexus":{
			"placementId":10433394
		},
		"pubmatic": {
			"adSlot": "AdTag_Div1@300x250",
			"publisherId": "999",
			"keywords": [{
				"key": "pmZoneID",
				"value": ["Zone1", "Zone2"]
			  },
			  {
				 "key": "preference",
				 "value": ["sports", "movies"]
			  }
			],
			"wrapper": {
				"version": 1,
				"profile": 5123
			}
		}	
	}`), nil
}

func ParseImpExts(orig *openrtb.BidRequest) ([]map[string]json.RawMessage, error) {
	imps := orig.Imp

	biddersData, err := GetBiddersData(orig)
	if err != nil {
		return nil, fmt.Errorf("Error getting raw bidders config: %v", err)
	}

	exts := make([]map[string]json.RawMessage, len(imps))
	// Loop over every impression in the request
	for i := 0; i < len(imps); i++ {
		// Unpack each set of extensions found in the Imp array
		err := json.Unmarshal(biddersData, &exts[i])
		if err != nil {
			return nil, fmt.Errorf("Error unpacking extensions for Imp[%d]: %s", i, err.Error())
		}
	}
	return exts, nil
}

func SelectWinningBids(req *openrtb.BidRequest, resp *openrtb.BidResponse) {
	winningBids := []openrtb.Bid{}

	maxBids := make(map[string]*openrtb.Bid)
	imps := make(map[string]*openrtb.Imp)

	for _, imp := range req.Imp {
		imps[imp.ID] = &imp
	}

	for _, seatBid := range resp.SeatBid {
		for i := range seatBid.Bid {
			// using bid from loop would cause loop assignment err
			bid := seatBid.Bid[i]
			if imp := imps[bid.ImpID]; imp != nil {
				if bid.Price < imp.BidFloor {
					// price too low
					continue
				}

				if maxBids[bid.ImpID] == nil || bid.Price > maxBids[bid.ImpID].Price {
					// using bid here would cause err
					maxBids[bid.ImpID] = &bid
				}
			}
		}
	}

	for _, imp := range req.Imp {
		if bid := maxBids[imp.ID]; bid != nil {
			winningBids = append(winningBids, *bid)
		}
	}

	// empty seatbids
	resp.SeatBid = []openrtb.SeatBid{}

	if len(winningBids) > 0 {
		seatBid := openrtb.SeatBid{
			Seat: "remixd",
			Bid:  winningBids,
		}

		resp.SeatBid = append(resp.SeatBid, seatBid)
	}
}
