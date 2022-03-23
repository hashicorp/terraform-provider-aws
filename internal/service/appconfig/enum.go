package appconfig

const (
	ConfigurationProfileTypeAWSAppConfigFeatureFlags = "AWS.AppConfig.FeatureFlags"
	ConfigurationProfileTypeAWSFreeform              = "AWS.Freeform"
)

func ConfigurationProfileType_Values() []string {
	return []string{
		ConfigurationProfileTypeAWSAppConfigFeatureFlags,
		ConfigurationProfileTypeAWSFreeform,
	}
}
