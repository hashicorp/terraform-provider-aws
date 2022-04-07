package plugintest

// Environment variables
const (
	// Disables checkpoint.hashicorp.com calls in Terraform CLI.
	EnvCheckpointDisable = "CHECKPOINT_DISABLE"

	// Environment variable with acceptance testing temporary directory for
	// testing files and Terraform CLI installation, if installation is
	// required. By default, the operating system temporary directory is used.
	//
	// Setting TF_ACC_TERRAFORM_PATH does not override this value for Terraform
	// CLI installation, if installation is required.
	EnvTfAccTempDir = "TF_ACC_TEMP_DIR"

	// Environment variable with path to save Terraform logs during acceptance
	// testing. This value sets TF_LOG_PATH in a safe manner when executing
	// Terraform CLI commands, which would otherwise be ignored since it could
	// interfere with how the underlying execution is performed.
	EnvTfAccLogPath = "TF_ACC_LOG_PATH"

	// Environment variable with acceptance testing Terraform CLI version to
	// download from releases.hashicorp.com, checksum verify, and install. The
	// value can be any valid Terraform CLI version, such as 1.1.6, with or
	// without a prepended v character.
	//
	// Setting this value takes precedence over using an available Terraform
	// binary in the operation system PATH, or if not found, installing the
	// latest version according to checkpoint.hashicorp.com.
	//
	// By default, the binary is installed in the operating system temporary
	// directory, however that directory can be overridden with the
	// TF_ACC_TEMP_DIR environment variable.
	//
	// If TF_ACC_TERRAFORM_PATH is also set, this installation method is
	// only invoked when a binary does not exist at that path. No version
	// checks are performed against an existing TF_ACC_TERRAFORM_PATH.
	EnvTfAccTerraformVersion = "TF_ACC_TERRAFORM_VERSION"

	// Acceptance testing path to Terraform CLI binary.
	//
	// Setting this value takes precedence over using an available Terraform
	// binary in the operation system PATH, or if not found, installing the
	// latest version according to checkpoint.hashicorp.com. This value does
	// not override TF_ACC_TEMP_DIR for Terraform CLI installation, if
	// installation is required.
	//
	// If TF_ACC_TERRAFORM_VERSION is not set, the binary must exist and be
	// executable, or an error will be returned.
	//
	// If TF_ACC_TERRAFORM_VERSION is also set, that Terraform CLI version
	// will be installed if a binary is not found at the given path. No version
	// checks are performed against an existing binary.
	EnvTfAccTerraformPath = "TF_ACC_TERRAFORM_PATH"
)
