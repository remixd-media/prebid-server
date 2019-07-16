package sonobi

import (
	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
	"text/template"
)

func NewSonobiSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("sonobi", 104, temp, adapters.SyncTypeRedirect)
}
