package ttx

import (
	"testing"

	"github.com/remixd-media/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "33across", New33AcrossBidder("http://ssc.33across.com"))
}
