package conns

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-testing-interface"
)

// Standard AWS environment variables used in the Terraform AWS Provider testing.
// These are not provided as constants in the AWS Go SDK currently.
const (
	// Default static credential identifier for tests (AWS Go SDK does not provide this as constant)
	// See also AWS_SECRET_ACCESS_KEY and AWS_PROFILE
	EnvVarAccessKeyId = "AWS_ACCESS_KEY_ID"

	// Container credentials endpoint
	// See also AWS_ACCESS_KEY_ID and AWS_PROFILE
	EnvVarContainerCredentialsFullURI = "AWS_CONTAINER_CREDENTIALS_FULL_URI"

	// Default AWS region for tests (AWS Go SDK does not provide this as constant)
	EnvVarDefaultRegion = "AWS_DEFAULT_REGION"

	// Default AWS shared configuration profile for tests (AWS Go SDK does not provide this as constant)
	EnvVarProfile = "AWS_PROFILE"

	// Default static credential value for tests (AWS Go SDK does not provide this as constant)
	// See also AWS_ACCESS_KEY_ID and AWS_PROFILE
	EnvVarSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

// Custom environment variables used in the Terraform AWS Provider testing.
// Additions should also be documented in the Environment Variable Dictionary
// of the Maintainers Guide: docs/MAINTAINING.md
const (
	// For tests using an alternate AWS account, the equivalent of AWS_ACCESS_KEY_ID for that account
	EnvVarAlternateAccessKeyId = "AWS_ALTERNATE_ACCESS_KEY_ID"

	// For tests using an alternate AWS account, the equivalent of AWS_PROFILE for that account
	EnvVarAlternateProfile = "AWS_ALTERNATE_PROFILE"

	// For tests using an alternate AWS region, the equivalent of AWS_DEFAULT_REGION for that account
	EnvVarAlternateRegion = "AWS_ALTERNATE_REGION"

	// For tests using an alternate AWS account, the equivalent of AWS_SECRET_ACCESS_KEY for that account
	EnvVarAlternateSecretAccessKey = "AWS_ALTERNATE_SECRET_ACCESS_KEY"

	// For tests using a third AWS region, the equivalent of AWS_DEFAULT_REGION for that region
	EnvVarThirdRegion = "AWS_THIRD_REGION"

	// For tests requiring GitHub permissions
	EnvVarGithubToken = "GITHUB_TOKEN"

	// For tests requiring restricted IAM permissions, an existing IAM Role to assume
	// An inline assume role policy is then used to deny actions for the test
	EnvVarAccAssumeRoleARN = "TF_ACC_ASSUME_ROLE_ARN"
)

// Custom environment variables used for assuming a role with resource sweepers
const (
	// The ARN of the IAM Role to assume
	EnvVarAssumeRoleARN = "TF_AWS_ASSUME_ROLE_ARN"

	// The duration in seconds the IAM role will be assumed.
	// Defaults to 1 hour (3600) instead of the SDK default of 15 minutes.
	EnvVarAssumeRoleDuration = "TF_AWS_ASSUME_ROLE_DURATION"

	// An External ID to pass to the assumed role
	EnvVarAssumeRoleExternalID = "TF_AWS_ASSUME_ROLE_EXTERNAL_ID"

	// A session name for the assumed role
	EnvVarAssumeRoleSessionName = "TF_AWS_ASSUME_ROLE_SESSION_NAME"
)

// GetEnvVarWithDefault gets an environment variable value if non-empty or returns the default.
func GetEnvVarWithDefault(variable string, defaultValue string) string {
	value := os.Getenv(variable)

	if value == "" {
		return defaultValue
	}

	return value
}

// RequireOneOfEnvVar verifies that at least one environment variable is non-empty or returns an error.
//
// If at lease one environment variable is non-empty, returns the first name and value.
func RequireOneOfEnvVar(names []string, usageMessage string) (string, string, error) {
	for _, variable := range names {
		value := os.Getenv(variable)

		if value != "" {
			return variable, value, nil
		}
	}

	return "", "", fmt.Errorf("at least one environment variable of %v must be set. Usage: %s", names, usageMessage)
}

// RequireEnvVar verifies that an environment variable is non-empty or returns an error.
func RequireEnvVar(name string, usageMessage string) (string, error) {
	value := os.Getenv(name)

	if value == "" {
		return "", fmt.Errorf("environment variable %s must be set. Usage: %s", name, usageMessage)
	}

	return value, nil
}

// FailIfAllEnvVarEmpty verifies that at least one environment variable is non-empty or fails the test.
//
// If at lease one environment variable is non-empty, returns the first name and value.
func FailIfAllEnvVarEmpty(t testing.T, names []string, usageMessage string) (string, string) {
	t.Helper()

	name, value, err := RequireOneOfEnvVar(names, usageMessage)
	if err != nil {
		t.Fatal(err)
		return "", ""
	}

	return name, value
}

// FailIfEnvVarEmpty verifies that an environment variable is non-empty or fails the test.
//
// For acceptance tests, this function must be used outside PreCheck functions to set values for configurations.
func FailIfEnvVarEmpty(t testing.T, name string, usageMessage string) string {
	t.Helper()

	value := os.Getenv(name)

	if value == "" {
		t.Fatalf("environment variable %s must be set. Usage: %s", name, usageMessage)
	}

	return value
}

// SkipIfEnvVarEmpty verifies that an environment variable is non-empty or skips the test.
//
// For acceptance tests, this function must be used outside PreCheck functions to set values for configurations.
func SkipIfEnvVarEmpty(t testing.T, name string, usageMessage string) string {
	t.Helper()

	value := os.Getenv(name)

	if value == "" {
		t.Skipf("skipping test; environment variable %s must be set. Usage: %s", name, usageMessage)
	}

	return value
}

// SkipIfAllEnvVarEmpty verifies that at least one environment variable is non-empty or skips the test.
//
// If at lease one environment variable is non-empty, returns the first name and value.
func SkipIfAllEnvVarEmpty(t testing.T, names []string, usageMessage string) (string, string) {
	t.Helper()

	name, value, err := RequireOneOfEnvVar(names, usageMessage)
	if err != nil {
		t.Skipf("skipping test because %s.", err)
		return "", ""
	}

	return name, value
}
