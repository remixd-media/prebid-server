package openrtb_ext

// ExtImpAdsWizz defines the contract for bidrequest.imp[i].ext.adswizz
type ExtImpAdsWizz struct {
	Alias string `json:"alias"`
	PName string `json:"pname"` //remixd custom field
}
