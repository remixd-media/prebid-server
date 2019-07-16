package remixd

import (
	"encoding/json"
	"fmt"

	"github.com/mxmCherry/openrtb"
)

func GetBiddersData(orig *openrtb.BidRequest) (json.RawMessage, error) {
	// placeholder
	return json.RawMessage(`{"appnexus":{"placementId":10433394}}`), nil
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
