package vrtcal

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

func NewVrtcalSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("vrtcal", 0, temp, adapters.SyncTypeRedirect)
}
