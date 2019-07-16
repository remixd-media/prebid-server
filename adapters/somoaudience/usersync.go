package somoaudience

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewSomoaudienceSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("somoaudience", 341, temp, adapters.SyncTypeRedirect)
}
