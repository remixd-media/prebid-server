package openx

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewOpenxSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("openx", 69, temp, adapters.SyncTypeRedirect)
}
