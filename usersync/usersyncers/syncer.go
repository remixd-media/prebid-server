package usersyncers

import (
	"strings"
	"text/template"

	"github.com/golang/glog"
	ttx "github.com/remixd-media/prebid-server/adapters/33across"
	"github.com/remixd-media/prebid-server/adapters/adform"
	"github.com/remixd-media/prebid-server/adapters/adkernelAdn"
	"github.com/remixd-media/prebid-server/adapters/adtelligent"
	"github.com/remixd-media/prebid-server/adapters/appnexus"
	"github.com/remixd-media/prebid-server/adapters/audienceNetwork"
	"github.com/remixd-media/prebid-server/adapters/beachfront"
	"github.com/remixd-media/prebid-server/adapters/brightroll"
	"github.com/remixd-media/prebid-server/adapters/consumable"
	"github.com/remixd-media/prebid-server/adapters/conversant"
	"github.com/remixd-media/prebid-server/adapters/eplanning"
	"github.com/remixd-media/prebid-server/adapters/gamoshi"
	"github.com/remixd-media/prebid-server/adapters/grid"
	"github.com/remixd-media/prebid-server/adapters/gumgum"
	"github.com/remixd-media/prebid-server/adapters/improvedigital"
	"github.com/remixd-media/prebid-server/adapters/ix"
	"github.com/remixd-media/prebid-server/adapters/lifestreet"
	"github.com/remixd-media/prebid-server/adapters/mgid"
	"github.com/remixd-media/prebid-server/adapters/openx"
	"github.com/remixd-media/prebid-server/adapters/pubmatic"
	"github.com/remixd-media/prebid-server/adapters/pulsepoint"
	"github.com/remixd-media/prebid-server/adapters/rhythmone"
	"github.com/remixd-media/prebid-server/adapters/rtbhouse"
	"github.com/remixd-media/prebid-server/adapters/rubicon"
	"github.com/remixd-media/prebid-server/adapters/sharethrough"
	"github.com/remixd-media/prebid-server/adapters/somoaudience"
	"github.com/remixd-media/prebid-server/adapters/sonobi"
	"github.com/remixd-media/prebid-server/adapters/sovrn"
	"github.com/remixd-media/prebid-server/adapters/unruly"
	"github.com/remixd-media/prebid-server/adapters/visx"
	"github.com/remixd-media/prebid-server/adapters/vrtcal"
	"github.com/remixd-media/prebid-server/adapters/yieldmo"
	"github.com/remixd-media/prebid-server/config"
	"github.com/remixd-media/prebid-server/openrtb_ext"
	"github.com/remixd-media/prebid-server/usersync"
)

// NewSyncerMap returns a map of all the usersyncer objects.
// The same keys should exist in this map as in the exchanges map.
// Static syncer map will be removed when adapter isolation is complete.
func NewSyncerMap(cfg *config.Configuration) map[openrtb_ext.BidderName]usersync.Usersyncer {
	syncers := make(map[openrtb_ext.BidderName]usersync.Usersyncer, len(cfg.Adapters))

	insertIntoMap(cfg, syncers, openrtb_ext.Bidder33Across, ttx.New33AcrossSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderAdform, adform.NewAdformSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderAdkernelAdn, adkernelAdn.NewAdkernelAdnSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderAdtelligent, adtelligent.NewAdtelligentSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderAppnexus, appnexus.NewAppnexusSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderBeachfront, beachfront.NewBeachfrontSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderBrightroll, brightroll.NewBrightrollSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderConsumable, consumable.NewConsumableSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderConversant, conversant.NewConversantSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderEPlanning, eplanning.NewEPlanningSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderFacebook, audienceNetwork.NewFacebookSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderGrid, grid.NewGridSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderGumGum, gumgum.NewGumGumSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderImprovedigital, improvedigital.NewImprovedigitalSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderIx, ix.NewIxSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderLifestreet, lifestreet.NewLifestreetSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderOpenx, openx.NewOpenxSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderPubmatic, pubmatic.NewPubmaticSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderPulsepoint, pulsepoint.NewPulsepointSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderRhythmone, rhythmone.NewRhythmoneSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderRTBHouse, rtbhouse.NewRTBHouseSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderRubicon, rubicon.NewRubiconSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderSharethrough, sharethrough.NewSharethroughSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderSomoaudience, somoaudience.NewSomoaudienceSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderSovrn, sovrn.NewSovrnSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderSonobi, sonobi.NewSonobiSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderVrtcal, vrtcal.NewVrtcalSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderYieldmo, yieldmo.NewYieldmoSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderVisx, visx.NewVisxSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderGamoshi, gamoshi.NewGamoshiSyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderUnruly, unruly.NewUnrulySyncer)
	insertIntoMap(cfg, syncers, openrtb_ext.BidderMgid, mgid.NewMgidSyncer)

	return syncers
}

func insertIntoMap(cfg *config.Configuration, syncers map[openrtb_ext.BidderName]usersync.Usersyncer, bidder openrtb_ext.BidderName, syncerFactory func(*template.Template) usersync.Usersyncer) {
	lowercased := strings.ToLower(string(bidder))
	urlString := cfg.Adapters[lowercased].UserSyncURL
	if urlString == "" {
		glog.Warningf("adapters." + string(bidder) + ".usersync_url was not defined, and their usersync API isn't flexible enough for Prebid Server to choose a good default. No usersyncs will be performed with " + string(bidder))
		return
	}
	syncers[bidder] = syncerFactory(template.Must(template.New(lowercased + "_usersync_url").Parse(urlString)))
}
