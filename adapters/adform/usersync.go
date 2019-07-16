package adform

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewAdformSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("adform", 50, temp, adapters.SyncTypeRedirect)
}
