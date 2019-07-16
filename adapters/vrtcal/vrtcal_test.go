package vrtcal

import (
	"testing"

	"github.com/remixd-media/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "vrtcaltest", NewVrtcalBidder("http://rtb.vrtcal.com/bidder_prebid.vap?ssp=1804"))
}
