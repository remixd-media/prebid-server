package wideorbit

import (
	"testing"

	"github.com/prebid/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "wideorbittest", NewWideOrbitBidder("https://mock.wideorbit.com?pbId=123"))
	// the extra "" in adm are not reflected in real requests, probably a bug in the testing module
}
