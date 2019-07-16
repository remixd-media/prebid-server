package yieldmo

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewYieldmoSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("yieldmo", 0, temp, adapters.SyncTypeRedirect)
}
