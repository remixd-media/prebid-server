package grid

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewGridSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("grid", 0, temp, adapters.SyncTypeRedirect)
}
