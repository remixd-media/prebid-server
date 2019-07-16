package gumgum

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewGumGumSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("gumgum", 61, temp, adapters.SyncTypeIframe)
}
