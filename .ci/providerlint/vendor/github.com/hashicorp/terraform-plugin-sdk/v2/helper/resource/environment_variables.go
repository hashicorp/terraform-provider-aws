// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

// Environment variables for acceptance testing. Additional environment
// variable constants can be found in the internal/plugintest package.
const (
	// Environment variable to enable acceptance tests using this package's
	// ParallelTest and Test functions whose TestCase does not enable the
	// IsUnitTest field. Defaults to disabled, in which each test will call
	// (*testing.T).Skip(). Can be set to any value to enable acceptance tests,
	// however "1" is conventional.
	EnvTfAcc = "TF_ACC"

	// Environment variable with hostname for the provider under acceptance
	// test. The hostname is the first portion of the full provider source
	// address, such as "example.com" in example.com/myorg/myprovider. Defaults
	// to "registry.terraform.io".
	//
	// Only required if any Terraform configuration set via the TestStep
	// type Config field includes a provider source, such as the terraform
	// configuration block required_providers attribute.
	EnvTfAccProviderHost = "TF_ACC_PROVIDER_HOST"

	// Environment variable with namespace for the provider under acceptance
	// test. The namespace is the second portion of the full provider source
	// address, such as "myorg" in registry.terraform.io/myorg/myprovider.
	// Defaults to "-" for Terraform 0.12-0.13 compatibility and "hashicorp".
	//
	// Only required if any Terraform configuration set via the TestStep
	// type Config field includes a provider source, such as the terraform
	// configuration block required_providers attribute.
	EnvTfAccProviderNamespace = "TF_ACC_PROVIDER_NAMESPACE"
)
