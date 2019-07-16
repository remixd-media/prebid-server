package beachfront

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewBeachfrontSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("beachfront", 0, temp, adapters.SyncTypeIframe)
}
