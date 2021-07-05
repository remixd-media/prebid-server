package voxnest

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

func NewVoxnestSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("voxnest", 0, temp, adapters.SyncTypeRedirect)
}
