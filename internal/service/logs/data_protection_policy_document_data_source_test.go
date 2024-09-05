// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsDataProtectionPolicyDocumentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	logGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	targetName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Logs),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionPolicyDocumentDataSourceConfig_basic(logGroupName, targetName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON(
						"data.aws_cloudwatch_log_data_protection_policy_document.test", names.AttrJSON,
						testAccDataProtectionPolicyDocumentDataSourceConfig_basic_expectedJSON(targetName)),
				),
			},
		},
	})
}

func TestAccLogsDataProtectionPolicyDocumentDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	logGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Logs),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProtectionPolicyDocumentDataSourceConfig_empty(logGroupName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrEquivalentJSON(
						"data.aws_cloudwatch_log_data_protection_policy_document.test", names.AttrJSON,
						testAccDataProtectionPolicyDocumentDataSourceConfig_empty_expectedJSON),
				),
			},
		},
	})
}

func TestAccLogsDataProtectionPolicyDocumentDataSource_errorOnBadOrderOfStatements(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Logs),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDataProtectionPolicyDocumentDataSourceConfig_errorOnBadOrderOfStatements,
				ExpectError: regexache.MustCompile(`the first policy statement must contain only the audit operation`),
			},
		},
	})
}

func TestAccLogsDataProtectionPolicyDocumentDataSource_errorOnNoOperation(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Logs),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProtectionPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDataProtectionPolicyDocumentDataSourceConfig_errorOnNoOperation,
				ExpectError: regexache.MustCompile(`the second policy statement must contain only the deidentify operation`),
			},
		},
	})
}

func testAccDataProtectionPolicyDocumentDataSourceConfig_basic(logGroupName, targetName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_data_protection_policy" "test" {
  log_group_name  = aws_cloudwatch_log_group.test.name
  policy_document = data.aws_cloudwatch_log_data_protection_policy_document.test.json
}

resource "aws_cloudwatch_log_group" "audit" {
  name = %[2]q
}

resource "aws_s3_bucket" "audit" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_iam_role" "firehose" {
  name = %[2]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose" {
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.audit.arn}",
        "${aws_s3_bucket.audit.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "audit" {
  depends_on = [aws_iam_role_policy.firehose]

  name        = %[2]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.audit.arn
  }

  tags = {
    # This tag appears after create.
    LogDeliveryEnabled = "true"
  }
}

data "aws_cloudwatch_log_data_protection_policy_document" "test" {
  description = "Test Document Description"
  name        = "Test"
  version     = "2021-06-01"

  statement {
    sid = "Audit"

    data_identifiers = [
      "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress",
      "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/DriversLicense-US",
    ]

    operation {
      audit {
        findings_destination {
          cloudwatch_logs {
            log_group = aws_cloudwatch_log_group.audit.name
          }
          firehose {
            delivery_stream = aws_kinesis_firehose_delivery_stream.audit.name
          }
          s3 {
            bucket = aws_s3_bucket.audit.bucket
          }
        }
      }
    }
  }

  statement {
    sid = "Deidentify"

    data_identifiers = [
      "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress",
      "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/DriversLicense-US",
    ]

    operation {
      deidentify {
        mask_config {}
      }
    }
  }
}
`,
		logGroupName, targetName)
}

func testAccDataProtectionPolicyDocumentDataSourceConfig_basic_expectedJSON(name string) string {
	// lintignore:AWSAT005
	return fmt.Sprintf(`
{
    "Description": "Test Document Description",
    "Name": "Test",
    "Version": "2021-06-01",
    "Statement": [
        {
            "Sid": "Audit",
            "DataIdentifier": [
                "arn:aws:dataprotection::aws:data-identifier/DriversLicense-US",
                "arn:aws:dataprotection::aws:data-identifier/EmailAddress"
            ],
            "Operation": {
                "Audit": {
                    "FindingsDestination": {
                        "CloudWatchLogs": {
                            "LogGroup": %[1]q
                        },
                        "Firehose": {
                            "DeliveryStream": %[1]q
                        },
                        "S3": {
                            "Bucket": %[1]q
                        }
                    }
                }
            }
        },
        {
            "Sid": "Deidentify",
            "DataIdentifier": [
                "arn:aws:dataprotection::aws:data-identifier/DriversLicense-US",
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
`, name)
}

func testAccDataProtectionPolicyDocumentDataSourceConfig_empty(logGroupName string) string {
	// lintignore:AWSAT005
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_data_protection_policy" "test" {
  log_group_name  = aws_cloudwatch_log_group.test.name
  policy_document = data.aws_cloudwatch_log_data_protection_policy_document.test.json
}

data "aws_cloudwatch_log_data_protection_policy_document" "test" {
  name = "Test"

  statement {
    data_identifiers = [
      "arn:aws:dataprotection::aws:data-identifier/EmailAddress",
    ]

    operation {
      audit {
        findings_destination {}
      }
    }
  }

  statement {
    data_identifiers = [
      "arn:aws:dataprotection::aws:data-identifier/EmailAddress",
    ]

    operation {
      deidentify {
        mask_config {}
      }
    }
  }
}
`,
		logGroupName)
}

// lintignore:AWSAT005
const testAccDataProtectionPolicyDocumentDataSourceConfig_empty_expectedJSON = `
{
    "Name": "Test",
    "Version": "2021-06-01",
    "Statement": [
        {
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
`

// lintignore:AWSAT005
const testAccDataProtectionPolicyDocumentDataSourceConfig_errorOnBadOrderOfStatements = `
data "aws_cloudwatch_log_data_protection_policy_document" "test" {
  name = "Test"

  statement {
    data_identifiers = [
      "arn:aws:dataprotection::aws:data-identifier/EmailAddress",
    ]

    operation {
      deidentify {
        mask_config {}
      }
    }
  }

  statement {
    data_identifiers = [
      "arn:aws:dataprotection::aws:data-identifier/EmailAddress",
    ]

    operation {
      audit {
        findings_destination {}
      }
    }
  }
}
`

// lintignore:AWSAT005
const testAccDataProtectionPolicyDocumentDataSourceConfig_errorOnNoOperation = `
data "aws_cloudwatch_log_data_protection_policy_document" "test" {
  name = "Test"

  statement {
    data_identifiers = [
      "arn:aws:dataprotection::aws:data-identifier/EmailAddress",
    ]

    operation {
      audit {
        findings_destination {}
      }
    }
  }

  statement {
    data_identifiers = [
      "arn:aws:dataprotection::aws:data-identifier/EmailAddress",
    ]

    operation {}
  }
}
`
