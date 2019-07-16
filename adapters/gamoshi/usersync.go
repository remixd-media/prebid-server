package gamoshi

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewGamoshiSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("gamoshi", 644, temp, adapters.SyncTypeRedirect)
}
