package rubicon

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewRubiconSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("rubicon", 52, temp, adapters.SyncTypeRedirect)
}
