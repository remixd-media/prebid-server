package remixd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mxmCherry/openrtb"
)

func TestSelectWinningBids(t *testing.T) {
	req := openrtb.BidRequest{
		Imp: []openrtb.Imp{
			openrtb.Imp{
				ID:       "1",
				BidFloor: 2.15,
			},
			openrtb.Imp{
				ID:       "2",
				BidFloor: 1.2,
			},
		},
	}

	resp := openrtb.BidResponse{
		SeatBid: []openrtb.SeatBid{
			openrtb.SeatBid{
				Seat: "dsp-1",
				Bid: []openrtb.Bid{
					openrtb.Bid{
						ImpID: "1",
						Price: 2.2,
					},
					openrtb.Bid{
						ImpID: "2",
						Price: 1.54,
					},
				},
			},
			openrtb.SeatBid{
				Seat: "dsp-2",
				Bid: []openrtb.Bid{
					openrtb.Bid{
						ImpID: "1",
						Price: 2.3,
					},
					openrtb.Bid{
						ImpID: "2",
						Price: 1.4,
					},
				},
			},
		},
	}

	SelectWinningBids(&req, &resp)

	assert.Equal(t, []openrtb.SeatBid{
		openrtb.SeatBid{
			Seat: "remixd",
			Bid: []openrtb.Bid{
				openrtb.Bid{
					ImpID: "1",
					Price: 2.3,
				},
				openrtb.Bid{
					ImpID: "2",
					Price: 1.54,
				},
			},
		},
	}, resp.SeatBid)
}
