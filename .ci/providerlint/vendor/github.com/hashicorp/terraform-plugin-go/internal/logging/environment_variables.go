// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

// Environment variables.
const (
	// EnvTfLogProvider is the prefix of the environment variable that sets the
	// logging level of the root provider logger for the provider being served.
	// The suffix is an underscore and the parsed provider name. For example,
	// registry.terraform.io/hashicorp/example becomes TF_LOG_PROVIDER_EXAMPLE.
	EnvTfLogProvider = "TF_LOG_PROVIDER"

	// EnvTfLogSdk is an environment variable that sets the root logging level
	// of SDK loggers.
	EnvTfLogSdk = "TF_LOG_SDK"

	// EnvTfLogSdkProto is an environment variable that sets the logging level
	// of SDK protocol loggers. Infers root SDK logging level, if unset.
	EnvTfLogSdkProto = "TF_LOG_SDK_PROTO"

	// EnvTfLogSdkProtoDataDir is an environment variable that sets the
	// directory to write raw protocol data files for debugging purposes.
	EnvTfLogSdkProtoDataDir = "TF_LOG_SDK_PROTO_DATA_DIR"
)
