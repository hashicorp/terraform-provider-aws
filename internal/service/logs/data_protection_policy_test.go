// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsDataProtectionPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy cloudwatchlogs.GetDataProtectionPolicyOutput
	resourceName := "aws_cloudwatch_log_data_protection_policy.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionPolicy_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProtectionPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLogGroupName, "aws_cloudwatch_log_group.test", names.AttrName),
					//lintignore:AWSAT005
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "policy_document", fmt.Sprintf(`
{
	"Name": "Test",
	"Version": "2021-06-01",
	"Statement": [
		{
			"Sid": "Audit",
			"DataIdentifier": [
				"arn:aws:dataprotection::aws:data-identifier/EmailAddress"
			],
			"Operation": {
				"Audit": {
					"FindingsDestination": {
                      "S3": {
                        "Bucket": %[1]q
                      }
                    }
				}
			}
		},
		{
			"Sid": "Redact",
			"DataIdentifier": [
				"arn:aws:dataprotection::aws:data-identifier/EmailAddress"
			],
			"Operation": {
				"Deidentify": {
					"MaskConfig": {}
				}
			}
		}
	]
}
`, name)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsDataProtectionPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy cloudwatchlogs.GetDataProtectionPolicyOutput
	resourceName := "aws_cloudwatch_log_data_protection_policy.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionPolicy_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProtectionPolicyExists(ctx, resourceName, &policy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceDataProtectionPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsDataProtectionPolicy_policyDocument(t *testing.T) {
	ctx := acctest.Context(t)
	var policy cloudwatchlogs.GetDataProtectionPolicyOutput
	resourceName := "aws_cloudwatch_log_data_protection_policy.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionPolicy_policyDocument1(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProtectionPolicyExists(ctx, resourceName, &policy),
					//lintignore:AWSAT005
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "policy_document", `
{
	"Name": "Test",
	"Version": "2021-06-01",
	"Statement": [
		{
			"Sid": "Audit",
			"DataIdentifier": [
				"arn:aws:dataprotection::aws:data-identifier/EmailAddress"
			],
			"Operation": {
				"Audit": {
					"FindingsDestination": {}
				}
			}
		},
		{
			"Sid": "Redact",
			"DataIdentifier": [
				"arn:aws:dataprotection::aws:data-identifier/EmailAddress"
			],
			"Operation": {
				"Deidentify": {
					"MaskConfig": {}
				}
			}
		}
	]
}
`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataProtectionPolicy_policyDocument2(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProtectionPolicyExists(ctx, resourceName, &policy),
					//lintignore:AWSAT005
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "policy_document", fmt.Sprintf(`
{
	"Name": "Test",
	"Version": "2021-06-01",
	"Statement": [
		{
			"Sid": "Audit",
			"DataIdentifier": [
				"arn:aws:dataprotection::aws:data-identifier/EmailAddress",
				"arn:aws:dataprotection::aws:data-identifier/DriversLicense-US"
			],
			"Operation": {
				"Audit": {
					"FindingsDestination": {
                      "S3": {
                        "Bucket": %[1]q
                      }
                    }
				}
			}
		},
		{
			"Sid": "Redact",
			"DataIdentifier": [
				"arn:aws:dataprotection::aws:data-identifier/EmailAddress",
				"arn:aws:dataprotection::aws:data-identifier/DriversLicense-US"
			],
			"Operation": {
				"Deidentify": {
					"MaskConfig": {}
				}
			}
		}
	]
}
`, name)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDataProtectionPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_data_protection_policy" {
				continue
			}

			_, err := tflogs.FindDataProtectionPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Data Protection Policy still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataProtectionPolicyExists(ctx context.Context, n string, v *cloudwatchlogs.GetDataProtectionPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Logs Data Protection Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		output, err := tflogs.FindDataProtectionPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDataProtectionPolicy_basic(name string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_cloudwatch_log_data_protection_policy" "test" {
  log_group_name = aws_cloudwatch_log_group.test.name
  policy_document = jsonencode({
    Name    = "Test"
    Version = "2021-06-01"

    Statement = [
      {
        Sid            = "Audit"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Audit = {
            FindingsDestination = {
              S3 = {
                Bucket = aws_s3_bucket.test.bucket
              }
            }
          }
        }
      },
      {
        Sid            = "Redact"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Deidentify = {
            MaskConfig = {}
          }
        }
      }
    ]
  })
}
`, name)
}

func testAccDataProtectionPolicy_policyDocument1(name string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_data_protection_policy" "test" {
  log_group_name = aws_cloudwatch_log_group.test.name
  policy_document = jsonencode({
    Name    = "Test"
    Version = "2021-06-01"

    Statement = [
      {
        Sid            = "Audit"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Audit = {
            FindingsDestination = {}
          }
        }
      },
      {
        Sid            = "Redact"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Deidentify = {
            MaskConfig = {}
          }
        }
      }
    ]
  })
}
`, name)
}

func testAccDataProtectionPolicy_policyDocument2(name string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_cloudwatch_log_data_protection_policy" "test" {
  log_group_name = aws_cloudwatch_log_group.test.name
  policy_document = jsonencode({
    Name    = "Test"
    Version = "2021-06-01"

    Statement = [
      {
        Sid = "Audit"
        DataIdentifier = [
          "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress",
          "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/DriversLicense-US",
        ]
        Operation = {
          Audit = {
            FindingsDestination = {
              S3 = {
                Bucket = aws_s3_bucket.test.bucket
              }
            }
          }
        }
      },
      {
        Sid = "Redact"
        DataIdentifier = [
          "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress",
          "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/DriversLicense-US",
        ]
        Operation = {
          Deidentify = {
            MaskConfig = {}
          }
        }
      }
    ]
  })
}
`, name)
}
