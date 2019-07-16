package rtbhouse

import (
	"testing"

	"github.com/remixd-media/prebid-server/adapters/adapterstest"
)

const testsDir = "rtbhousetest"
const testsBidderEndpoint = "http://localhost/prebid_server"

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, testsDir, NewRTBHouseBidder(testsBidderEndpoint))
}
