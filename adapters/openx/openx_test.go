package openx

import (
	"testing"

	"github.com/remixd-media/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "openxtest", NewOpenxBidder("http://rtb.openx.net/prebid"))
}
