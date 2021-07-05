package voxnest

import (
	"testing"

	"github.com/prebid/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "voxnesttest", &VoxnestAdapter{endpoint: "https://mock.voxnest.com"})
	// the extra "" in adm are not reflected in real requests, probably a bug in the testing module
}
