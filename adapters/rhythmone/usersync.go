package rhythmone

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewRhythmoneSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("rhythmone", 36, temp, adapters.SyncTypeRedirect)
}
