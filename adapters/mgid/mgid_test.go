package mgid

import (
	"github.com/remixd-media/prebid-server/adapters/adapterstest"
	"testing"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "mgidtest", NewMgidBidder("https://prebid.mgid.com/prebid/"))
}
