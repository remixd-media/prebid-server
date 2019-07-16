package eplanning

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewEPlanningSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("eplanning", 0, temp, adapters.SyncTypeIframe)
}
