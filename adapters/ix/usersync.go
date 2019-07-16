package ix

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewIxSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("ix", 10, temp, adapters.SyncTypeRedirect)
}
