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

func TestIAMStatementsDistributeFunction_basic(t *testing.T) {
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
				Config: testIAMStatementsDistributeFunctionConfig(validPolicy, "customer-managed", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: Add specific output checks once implementation is complete
				),
			},
		},
	})
}

func TestIAMStatementsDistributeFunction_invalidJSON(t *testing.T) {
	t.Parallel()

	invalidJSON := `{"Version": "2012-10-17", "Statement": [`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config:      testIAMStatementsDistributeFunctionConfig(invalidJSON, "customer-managed", ""),
				ExpectError: regexache.MustCompile("JSON syntax error|unexpected end of JSON input"),
			},
		},
	})
}

func testIAMStatementsDistributeFunctionConfig(policyJSON, policyType, maxSize string) string {
	if maxSize != "" {
		return fmt.Sprintf(`
output "test" {
  value = provider::aws::iam_statements_distribute(%[1]q, %[2]q, %[3]s)
}
`, policyJSON, policyType, maxSize)
	}

	if policyType != "" {
		return fmt.Sprintf(`
output "test" {
  value = provider::aws::iam_statements_distribute(%[1]q, %[2]q)
}
`, policyJSON, policyType)
	}

	return fmt.Sprintf(`
output "test" {
  value = provider::aws::iam_statements_distribute(%[1]q)
}
`, policyJSON)
}
func TestIAMStatementsDistributeFunction_missingVersion(t *testing.T) {
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
				Config:      testIAMStatementsDistributeFunctionConfig(policyMissingVersion, "customer-managed", ""),
				ExpectError: regexache.MustCompile("policy document missing required field: Version"),
			},
		},
	})
}

func TestIAMStatementsDistributeFunction_missingStatement(t *testing.T) {
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
				Config:      testIAMStatementsDistributeFunctionConfig(policyMissingStatement, "customer-managed", ""),
				ExpectError: regexache.MustCompile("policy document missing required field: Statement"),
			},
		},
	})
}

func TestIAMStatementsDistributeFunction_missingPolicyType(t *testing.T) {
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
				Config:      testIAMStatementsDistributeFunctionConfig(validPolicy, "", ""),
				ExpectError: regexache.MustCompile("Not enough function arguments|Missing value for \"policy_type\""),
			},
		},
	})
}

func TestIAMStatementsDistributeFunction_invalidPolicyType(t *testing.T) {
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
				Config:      testIAMStatementsDistributeFunctionConfig(validPolicy, "invalid", ""),
				ExpectError: regexache.MustCompile("policy_type 'invalid'"),
			},
		},
	})
}
func TestIAMStatementsDistributeFunction_policyLimits(t *testing.T) {
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
				Config: testIAMStatementsDistributeFunctionConfig(validPolicy, "customer-managed", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: Add specific output checks once implementation is complete
				),
			},
			{
				Config: testIAMStatementsDistributeFunctionConfig(validPolicy, "service-control-policy", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: Add specific output checks once implementation is complete
				),
			},
		},
	})
}
func TestIAMStatementsDistributeFunction_allPolicyTypes(t *testing.T) {
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
				Config: testIAMStatementsDistributeFunctionConfig(validPolicy, "customer-managed", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
			{
				Config: testIAMStatementsDistributeFunctionConfig(validPolicy, "service-control-policy", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
		},
	})
}

func TestIAMStatementsDistributeFunction_policyWithId(t *testing.T) {
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
				Config: testIAMStatementsDistributeFunctionConfig(policyWithId, "customer-managed", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
		},
	})
}

func TestIAMStatementsDistributeFunction_complexStatements(t *testing.T) {
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
				Config: testIAMStatementsDistributeFunctionConfig(complexPolicy, "customer-managed", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
		},
	})
}

func TestIAMStatementsDistributeFunction_edgeCases(t *testing.T) {
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
				Config: testIAMStatementsDistributeFunctionConfig(minimalPolicy, "customer-managed", ""),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// Function should execute without error
				),
			},
		},
	})
}

func TestIAMStatementsDistributeFunction_malformedStatements(t *testing.T) {
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
				Config:      testIAMStatementsDistributeFunctionConfig(policyMissingEffect, "customer-managed", ""),
				ExpectError: regexache.MustCompile("statement missing required field: Effect"),
			},
		},
	})
}

func TestIAMStatementsDistributeFunction_unsupportedVersion(t *testing.T) {
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
				Config:      testIAMStatementsDistributeFunctionConfig(policyUnsupportedVersion, "customer-managed", ""),
				ExpectError: regexache.MustCompile("unsupported policy version"),
			},
		},
	})
}
