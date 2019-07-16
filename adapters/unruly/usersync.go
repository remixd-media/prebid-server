package unruly

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewUnrulySyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("unruly", 162, temp, adapters.SyncTypeRedirect)
}
