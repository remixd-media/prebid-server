package adkernelAdn

import (
	"text/template"

	"github.com/remixd-media/prebid-server/adapters"
	"github.com/remixd-media/prebid-server/usersync"
)

const adkernelGDPRVendorID = uint16(14)

func NewAdkernelAdnSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("adkernelAdn", 14, temp, adapters.SyncTypeRedirect)
}
