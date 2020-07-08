package tritondigital

import (
	"testing"

	"github.com/prebid/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "tritondigitaltest", NewTritonDigitalBidder("https://tritondigitalmock.example.com/ondemand/ars"))
	// the extra "" in adm are not reflected in real requests, probably a bug in the testing module
}
