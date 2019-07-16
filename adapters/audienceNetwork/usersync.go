package audienceNetwork

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewFacebookSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("audienceNetwork", 0, temp, adapters.SyncTypeRedirect)
}
