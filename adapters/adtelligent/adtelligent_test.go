package adtelligent

import (
	"testing"

	"github.com/remixd-media/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "adtelligenttest", NewAdtelligentBidder("http://hb.adtelligent.com/auction"))
}
