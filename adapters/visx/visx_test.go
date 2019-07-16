package visx

import (
	"testing"

	"github.com/remixd-media/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "visxtest", NewVisxBidder("http://localhost/prebid"))
}
