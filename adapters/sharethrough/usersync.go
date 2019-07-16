package sharethrough

import (
	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
	"text/template"
)

func NewSharethroughSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("sharethrough", 80, temp, adapters.SyncTypeRedirect)
}
