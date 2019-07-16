package yieldmo

import (
	"testing"

	"github.com/remixd-media/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "yieldmotest", NewYieldmoBidder("https://ads.yieldmo.com/openrtb2"))
}
