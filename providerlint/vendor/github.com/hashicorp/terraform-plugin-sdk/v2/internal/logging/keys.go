package logging

// Structured logging keys.
//
// Practitioners or tooling reading logs may be depending on these keys, so be
// conscious of that when changing them.
//
// Refer to the terraform-plugin-go logging keys as well, which should be
// equivalent to these when possible.
const (
	// The type of data source being operated on, such as "archive_file"
	KeyDataSourceType = "tf_data_source_type"

	// Underlying Go error string when logging an error.
	KeyError = "error"

	// The full address of the provider, such as
	// registry.terraform.io/hashicorp/random
	KeyProviderAddress = "tf_provider_addr"

	// The type of resource being operated on, such as "random_pet"
	KeyResourceType = "tf_resource_type"

	// The name of the test being executed.
	KeyTestName = "test_name"

	// The TestStep number of the test being executed. Starts at 1.
	KeyTestStepNumber = "test_step_number"

	// The path to the Terraform CLI used for an acceptance test.
	KeyTestTerraformPath = "test_terraform_path"

	// The working directory of the acceptance test.
	KeyTestWorkingDirectory = "test_working_directory"
)
