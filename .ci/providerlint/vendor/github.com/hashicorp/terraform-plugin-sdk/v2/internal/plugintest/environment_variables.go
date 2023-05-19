package plugintest

// Environment variables
const (
	// Environment variable with acceptance testing temporary directory for
	// testing files and Terraform CLI installation, if installation is
	// required. By default, the operating system temporary directory is used.
	//
	// Setting TF_ACC_TERRAFORM_PATH does not override this value for Terraform
	// CLI installation, if installation is required.
	EnvTfAccTempDir = "TF_ACC_TEMP_DIR"

	// Environment variable with level to filter Terraform logs during
	// acceptance testing. This value sets TF_LOG in a safe manner when
	// executing Terraform CLI commands, which would otherwise interfere
	// with the testing framework using TF_LOG to set the Go standard library
	// log package level.
	//
	// This value takes precedence over TF_LOG_CORE, due to precedence rules
	// in the Terraform core code, so it is not possible to set this to a level
	// and also TF_LOG_CORE=OFF. Use TF_LOG_CORE and TF_LOG_PROVIDER in that
	// case instead.
	//
	// If not set, but TF_ACC_LOG_PATH or TF_LOG_PATH_MASK is set, it defaults
	// to TRACE. If Terraform CLI is version 0.14 or earlier, it will have no
	// separate affect from the TF_ACC_LOG_PATH or TF_LOG_PATH_MASK behavior,
	// as those earlier versions of Terraform are unreliable with the logging
	// level being outside TRACE.
	EnvTfAccLog = "TF_ACC_LOG"

	// Environment variable with path to save Terraform logs during acceptance
	// testing. This value sets TF_LOG_PATH in a safe manner when executing
	// Terraform CLI commands, which would otherwise be ignored since it could
	// interfere with how the underlying execution is performed.
	//
	// If TF_LOG_PATH_MASK is set, it takes precedence over this value.
	EnvTfAccLogPath = "TF_ACC_LOG_PATH"

	// Environment variable with level to filter Terraform core logs during
	// acceptance testing. This value sets TF_LOG_CORE separate from
	// TF_LOG_PROVIDER when calling Terraform.
	//
	// This value has no affect when TF_ACC_LOG is set (which sets Terraform's
	// TF_LOG), due to precedence rules in the Terraform core code. Use
	// TF_LOG_CORE and TF_LOG_PROVIDER in that case instead.
	//
	// If not set, defaults to TF_ACC_LOG behaviors.
	EnvTfLogCore = "TF_LOG_CORE"

	// Environment variable with path containing the string %s, which is
	// replaced with the test name, to save separate Terraform logs during
	// acceptance testing. This value sets TF_LOG_PATH in a safe manner when
	// executing Terraform CLI commands, which would otherwise be ignored since
	// it could interfere with how the underlying execution is performed.
	//
	// Takes precedence over TF_ACC_LOG_PATH.
	EnvTfLogPathMask = "TF_LOG_PATH_MASK"

	// Environment variable with level to filter Terraform provider logs during
	// acceptance testing. This value sets TF_LOG_PROVIDER separate from
	// TF_LOG_CORE.
	//
	// During testing, this only affects external providers whose logging goes
	// through Terraform. The logging for the provider under test is controlled
	// by the testing framework as it is running the provider code. Provider
	// code using the Go standard library log package is controlled by TF_LOG
	// for historical compatibility.
	//
	// This value takes precedence over TF_ACC_LOG for external provider logs,
	// due to rules in the Terraform core code.
	//
	// If not set, defaults to TF_ACC_LOG behaviors.
	EnvTfLogProvider = "TF_LOG_PROVIDER"

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
