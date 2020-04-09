package bidswitch

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

func NewBidSwitchSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("bidswitch", 0, temp, adapters.SyncTypeRedirect)
}
