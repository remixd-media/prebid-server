package pulsepoint

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewPulsepointSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("pulsepoint", 81, temp, adapters.SyncTypeRedirect)
}
