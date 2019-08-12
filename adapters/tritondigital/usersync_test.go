package tritondigital

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestTritonDigitalSyncer(t *testing.T) {
	temp := template.Must(template.New("sync-template").Parse("//not_localhost/synclocalhost%2Fsetuid%3Fbidder%3Dtritondigital%26uid%3D@UUID@"))
	syncer := NewTritonDigitalSyncer(temp)
	syncInfo, err := syncer.GetUsersyncInfo("0", "")
	assert.NoError(t, err)
	assert.Equal(t, "//not_localhost/synclocalhost%2Fsetuid%3Fbidder%3Dtritondigital%26uid%3D@UUID@", syncInfo.URL)
	assert.Equal(t, "redirect", syncInfo.Type)
	assert.EqualValues(t, 0, syncer.GDPRVendorID())
	assert.Equal(t, false, syncInfo.SupportCORS)
}
