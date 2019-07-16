package brightroll

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewBrightrollSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("brightroll", 25, temp, adapters.SyncTypeRedirect)
}
