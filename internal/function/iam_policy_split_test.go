// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestIAMPolicySplitFunction_basic(t *testing.T) {
	t.Parallel()

	validPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::example-bucket/*"
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testIAMPolicySplitFunctionConfig(validPolicy, "inline", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: Add specific output checks once implementation is complete
				),
			},
		},
	})
}

func TestIAMPolicySplitFunction_invalidJSON(t *testing.T) {
	t.Parallel()

	invalidJSON := `{"Version": "2012-10-17", "Statement": [`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config:      testIAMPolicySplitFunctionConfig(invalidJSON, "inline", ""),
				ExpectError: regexache.MustCompile("JSON syntax error"),
			},
		},
	})
}

func testIAMPolicySplitFunctionConfig(policyJSON, serviceType, maxSize string) string {
	if maxSize != "" {
		return fmt.Sprintf(`
output "test" {
  value = provider::aws::iam_policy_split(%[1]q, %[2]q, %[3]s)
}
`, policyJSON, serviceType, maxSize)
	}

	if serviceType != "" {
		return fmt.Sprintf(`
output "test" {
  value = provider::aws::iam_policy_split(%[1]q, %[2]q)
}
`, policyJSON, serviceType)
	}

	return fmt.Sprintf(`
output "test" {
  value = provider::aws::iam_policy_split(%[1]q)
}
`, policyJSON)
}
func TestIAMPolicySplitFunction_missingVersion(t *testing.T) {
	t.Parallel()

	policyMissingVersion := `{
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::example-bucket/*"
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config:      testIAMPolicySplitFunctionConfig(policyMissingVersion, "inline", ""),
				ExpectError: regexache.MustCompile("policy document missing required field: Version"),
			},
		},
	})
}

func TestIAMPolicySplitFunction_missingStatement(t *testing.T) {
	t.Parallel()

	policyMissingStatement := `{
		"Version": "2012-10-17",
		"Statement": []
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config:      testIAMPolicySplitFunctionConfig(policyMissingStatement, "inline", ""),
				ExpectError: regexache.MustCompile("policy document missing required field: Statement"),
			},
		},
	})
}

func TestIAMPolicySplitFunction_invalidServiceType(t *testing.T) {
	t.Parallel()

	validPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::example-bucket/*"
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config:      testIAMPolicySplitFunctionConfig(validPolicy, "invalid", ""),
				ExpectError: regexache.MustCompile("service_type 'invalid'"),
			},
		},
	})
}
func TestIAMPolicySplitFunction_serviceLimits(t *testing.T) {
	t.Parallel()

	validPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::example-bucket/*"
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testIAMPolicySplitFunctionConfig(validPolicy, "inline", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: Add specific output checks once implementation is complete
				),
			},
			{
				Config: testIAMPolicySplitFunctionConfig(validPolicy, "managed", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: Add specific output checks once implementation is complete
				),
			},
			{
				Config: testIAMPolicySplitFunctionConfig(validPolicy, "resource-based", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: Add specific output checks once implementation is complete
				),
			},
		},
	})
}
func TestIAMPolicySplitFunction_allServiceTypes(t *testing.T) {
	t.Parallel()

	validPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": ["s3:GetObject", "s3:PutObject"],
				"Resource": "arn:aws:s3:::example-bucket/*"
			},
			{
				"Effect": "Deny",
				"Action": "s3:DeleteObject",
				"Resource": "arn:aws:s3:::example-bucket/*"
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testIAMPolicySplitFunctionConfig(validPolicy, "inline", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
			{
				Config: testIAMPolicySplitFunctionConfig(validPolicy, "managed", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
			{
				Config: testIAMPolicySplitFunctionConfig(validPolicy, "resource-based", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
		},
	})
}

func TestIAMPolicySplitFunction_policyWithId(t *testing.T) {
	t.Parallel()

	policyWithId := `{
		"Version": "2012-10-17",
		"Id": "MyTestPolicy",
		"Statement": [
			{
				"Sid": "AllowS3Access",
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::example-bucket/*"
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testIAMPolicySplitFunctionConfig(policyWithId, "inline", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
		},
	})
}

func TestIAMPolicySplitFunction_complexStatements(t *testing.T) {
	t.Parallel()

	complexPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"s3:GetObject",
					"s3:PutObject",
					"s3:DeleteObject"
				],
				"Resource": [
					"arn:aws:s3:::bucket1/*",
					"arn:aws:s3:::bucket2/*"
				],
				"Condition": {
					"StringEquals": {
						"s3:x-amz-server-side-encryption": "AES256"
					}
				}
			},
			{
				"Effect": "Deny",
				"NotAction": "s3:GetObject",
				"NotResource": "arn:aws:s3:::secure-bucket/*",
				"Principal": {
					"AWS": "arn:aws:iam::123456789012:user/testuser"
				}
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testIAMPolicySplitFunctionConfig(complexPolicy, "managed", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
		},
	})
}

func TestIAMPolicySplitFunction_edgeCases(t *testing.T) {
	t.Parallel()

	// Test with minimal policy
	minimalPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "*",
				"Resource": "*"
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testIAMPolicySplitFunctionConfig(minimalPolicy, "inline", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
		},
	})
}

func TestIAMPolicySplitFunction_malformedStatements(t *testing.T) {
	t.Parallel()

	// Policy with missing Effect
	policyMissingEffect := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::bucket/*"
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config:      testIAMPolicySplitFunctionConfig(policyMissingEffect, "inline", ""),
				ExpectError: regexache.MustCompile("statement missing required field: Effect"),
			},
		},
	})
}

func TestIAMPolicySplitFunction_unsupportedVersion(t *testing.T) {
	t.Parallel()

	policyUnsupportedVersion := `{
		"Version": "2020-01-01",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::bucket/*"
			}
		]
	}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config:      testIAMPolicySplitFunctionConfig(policyUnsupportedVersion, "inline", ""),
				ExpectError: regexache.MustCompile("unsupported policy version"),
			},
		},
	})
}
