package tritondigital

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

func NewTritonDigitalSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("tritondigital", 0, temp, adapters.SyncTypeRedirect)
}
