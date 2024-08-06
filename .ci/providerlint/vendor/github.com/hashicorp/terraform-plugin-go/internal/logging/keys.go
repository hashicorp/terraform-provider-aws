// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

// Global logging keys attached to all requests.
//
// Practitioners or tooling reading logs may be depending on these keys, so be
// conscious of that when changing them.
const (
	// Attribute of the diagnostic being logged.
	KeyDiagnosticAttribute = "diagnostic_attribute"

	// Number of the error diagnostics.
	KeyDiagnosticErrorCount = "diagnostic_error_count"

	// Severity of the diagnostic being logged.
	KeyDiagnosticSeverity = "diagnostic_severity"

	// Detail of the diagnostic being logged.
	KeyDiagnosticDetail = "diagnostic_detail"

	// Summary of the diagnostic being logged.
	KeyDiagnosticSummary = "diagnostic_summary"

	// Number of the warning diagnostics.
	KeyDiagnosticWarningCount = "diagnostic_warning_count"

	// Underlying error string
	KeyError = "error"

	// Argument position of the function error.
	KeyFunctionErrorArgument = "function_error_argument"

	// Boolean indicating presence of function error
	KeyFunctionErrorExists = "function_error_exists"

	// Message of the function error.
	KeyFunctionErrorText = "function_error_text"

	// Duration in milliseconds for the RPC request
	KeyRequestDurationMs = "tf_req_duration_ms"

	// A unique ID for the RPC request
	KeyRequestID = "tf_req_id"

	// The full address of the provider, such as
	// registry.terraform.io/hashicorp/random
	KeyProviderAddress = "tf_provider_addr"

	// The RPC being run, such as "ApplyResourceChange"
	KeyRPC = "tf_rpc"

	// The type of resource being operated on, such as "random_pet"
	KeyResourceType = "tf_resource_type"

	// The type of data source being operated on, such as "archive_file"
	KeyDataSourceType = "tf_data_source_type"

	// Path to protocol data file, such as "/tmp/example.json"
	KeyProtocolDataFile = "tf_proto_data_file"

	// The protocol version being used, as a string, such as "6"
	KeyProtocolVersion = "tf_proto_version"

	// The Deferred reason for an RPC response
	KeyDeferredReason = "tf_deferred_reason"

	// Whether the GetProviderSchemaOptional server capability is enabled
	KeyServerCapabilityGetProviderSchemaOptional = "tf_server_capability_get_provider_schema_optional"

	// Whether the PlanDestroy server capability is enabled
	KeyServerCapabilityPlanDestroy = "tf_server_capability_plan_destroy"

	// Whether the DeferralAllowed client capability is enabled
	KeyClientCapabilityDeferralAllowed = "tf_client_capability_deferral_allowed"
)
