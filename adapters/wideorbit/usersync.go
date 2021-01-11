package wideorbit

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

func NewWideOrbitSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("wideorbit", 0, temp, adapters.SyncTypeRedirect)
}
