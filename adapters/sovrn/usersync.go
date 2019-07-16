package sovrn

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewSovrnSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("sovrn", 13, temp, adapters.SyncTypeRedirect)
}
