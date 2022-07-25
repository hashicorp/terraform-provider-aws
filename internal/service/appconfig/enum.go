package appconfig

const (
	configurationProfileTypeFeatureFlags = "AWS.AppConfig.FeatureFlags"
	configurationProfileTypeFreeform     = "AWS.Freeform"
)

func ConfigurationProfileType_Values() []string {
	return []string{
		configurationProfileTypeFeatureFlags,
		configurationProfileTypeFreeform,
	}
}
