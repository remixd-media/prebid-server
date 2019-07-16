package mgid

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewMgidSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("mgid", 358, temp, adapters.SyncTypeRedirect)
}
