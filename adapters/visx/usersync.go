package visx

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewVisxSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("visx", 0, temp, adapters.SyncTypeRedirect)
}
