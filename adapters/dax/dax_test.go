package dax

import (
	"testing"

	"github.com/prebid/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "daxtest", &DaxAdapter{endpoint: "https://mock.com?cid=123"})
	// the extra "" in adm are not reflected in real requests, probably a bug in the testing module
}
