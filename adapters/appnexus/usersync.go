package appnexus

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewAppnexusSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("adnxs", 32, temp, adapters.SyncTypeRedirect)
}
