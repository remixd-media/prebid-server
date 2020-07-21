package adswizz

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

func NewAdsWizzSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("adswizz", 0, temp, adapters.SyncTypeRedirect)
}
