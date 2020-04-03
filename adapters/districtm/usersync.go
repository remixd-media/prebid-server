package districtm

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

func NewDistrictMSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("districtm", 0, temp, adapters.SyncTypeRedirect)
}
