package adtelligent

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewAdtelligentSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("adtelligent", 0, temp, adapters.SyncTypeRedirect)
}
