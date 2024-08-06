// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

// Structured logging keys.
//
// Practitioners or tooling reading logs may be depending on these keys, so be
// conscious of that when changing them.
//
// Refer to the terraform-plugin-go logging keys as well, which should be
// equivalent to these when possible.
const (
	// Attribute path representation, which is typically in flatmap form such
	// as parent.0.child in this project.
	KeyAttributePath = "tf_attribute_path"

	// The type of data source being operated on, such as "archive_file"
	KeyDataSourceType = "tf_data_source_type"

	// Underlying Go error string when logging an error.
	KeyError = "error"

	// The full address of the provider, such as
	// registry.terraform.io/hashicorp/random
	KeyProviderAddress = "tf_provider_addr"

	// The type of resource being operated on, such as "random_pet"
	KeyResourceType = "tf_resource_type"

	// The Deferred reason for an RPC response
	KeyDeferredReason = "tf_deferred_reason"

	// The name of the test being executed.
	KeyTestName = "test_name"

	// The TestStep number of the test being executed. Starts at 1.
	KeyTestStepNumber = "test_step_number"

	// Terraform configuration used during acceptance testing Terraform operations.
	KeyTestTerraformConfiguration = "test_terraform_configuration"

	// The Terraform CLI logging level (TF_LOG) used for an acceptance test.
	KeyTestTerraformLogLevel = "test_terraform_log_level"

	// The Terraform CLI logging level (TF_LOG_CORE) used for an acceptance test.
	KeyTestTerraformLogCoreLevel = "test_terraform_log_core_level"

	// The Terraform CLI logging level (TF_LOG_PROVIDER) used for an acceptance test.
	KeyTestTerraformLogProviderLevel = "test_terraform_log_provider_level"

	// The path to the Terraform CLI logging file used for an acceptance test.
	//
	// This should match where the rest of the acceptance test logs are going
	// already, but is provided for troubleshooting in case it does not.
	KeyTestTerraformLogPath = "test_terraform_log_path"

	// The path to the Terraform CLI used for an acceptance test.
	KeyTestTerraformPath = "test_terraform_path"

	// Terraform plan output generated during a TestStep.
	KeyTestTerraformPlan = "test_terraform_plan"

	// The working directory of the acceptance test.
	KeyTestWorkingDirectory = "test_working_directory"
)
