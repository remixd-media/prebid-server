package ttx

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func New33AcrossSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("33across", 58, temp, adapters.SyncTypeIframe)
}
