package improvedigital

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewImprovedigitalSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("improvedigital", 253, temp, adapters.SyncTypeRedirect)
}
