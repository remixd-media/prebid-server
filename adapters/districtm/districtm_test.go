package districtm

import (
	"testing"

	"github.com/prebid/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "districtmtest", &DistrictMAdapter{endpoint: "https://xtest.districtm.com/"})
	// the extra "" in adm are not reflected in real requests, probably a bug in the testing module
}
