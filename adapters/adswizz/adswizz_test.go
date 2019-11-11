package adswizz

import (
	"testing"

	"github.com/prebid/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "adswizztest", NewAdsWizzBidder("https://mock.deliveryengine.adswizz.com/vast/4.0/request/alias"))
	// the extra "" in adm are not reflected in real requests, probably a bug in the testing module
}
