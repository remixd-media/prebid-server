package wideorbit

import (
	"testing"
	"text/template"

	"github.com/prebid/prebid-server/privacy"
	"github.com/stretchr/testify/assert"
)

func TestWideOrbitSyncer(t *testing.T) {
	temp := template.Must(template.New("sync-template").Parse("//not_localhost/synclocalhost%2Fsetuid%3Fbidder%3Dadswizz%26uid%3D@UUID@"))
	syncer := NewWideOrbitSyncer(temp)
	syncInfo, err := syncer.GetUsersyncInfo(privacy.Policies{})
	assert.NoError(t, err)
	assert.Equal(t, "//not_localhost/synclocalhost%2Fsetuid%3Fbidder%3Dadswizz%26uid%3D@UUID@", syncInfo.URL)
	assert.Equal(t, "redirect", syncInfo.Type)
	assert.EqualValues(t, 0, syncer.GDPRVendorID())
	assert.Equal(t, false, syncInfo.SupportCORS)
}
