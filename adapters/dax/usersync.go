package dax

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

func NewDaxSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("dax", 0, temp, adapters.SyncTypeRedirect)
}
