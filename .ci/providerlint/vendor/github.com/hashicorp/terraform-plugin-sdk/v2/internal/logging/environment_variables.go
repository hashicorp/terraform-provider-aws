// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

// Environment variables.
const (
	// EnvTfLogSdk is an environment variable that sets the logging level of
	// the root SDK logger, while the provider is under test. In "production"
	// usage, this environment variable is handled by terraform-plugin-go.
	//
	// Terraform CLI's logging must be explicitly turned on before this
	// environment varable can be used to reduce the SDK logging levels. It
	// cannot be used to show only SDK logging unless all other logging levels
	// are turned off.
	EnvTfLogSdk = "TF_LOG_SDK"

	// EnvTfLogSdkHelperResource is an environment variable that sets the logging
	// level of SDK helper/resource loggers. Infers root SDK logging level, if
	// unset.
	EnvTfLogSdkHelperResource = "TF_LOG_SDK_HELPER_RESOURCE"

	// EnvTfLogSdkHelperSchema is an environment variable that sets the logging
	// level of SDK helper/schema loggers. Infers root SDK logging level, if
	// unset.
	EnvTfLogSdkHelperSchema = "TF_LOG_SDK_HELPER_SCHEMA"
)
