package logging

// Environment variables.
const (
	// EnvTfLogSdkHelperResource is an environment variable that sets the logging
	// level of SDK helper/resource loggers. Infers root SDK logging level, if
	// unset.
	EnvTfLogSdkHelperResource = "TF_LOG_SDK_HELPER_RESOURCE"

	// EnvTfLogSdkHelperSchema is an environment variable that sets the logging
	// level of SDK helper/schema loggers. Infers root SDK logging level, if
	// unset.
	EnvTfLogSdkHelperSchema = "TF_LOG_SDK_HELPER_SCHEMA"
)
