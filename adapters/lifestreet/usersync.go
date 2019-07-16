package lifestreet

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewLifestreetSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("lifestreet", 67, temp, adapters.SyncTypeRedirect)
}
