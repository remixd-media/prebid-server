package pubmatic

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewPubmaticSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("pubmatic", 76, temp, adapters.SyncTypeIframe)
}
