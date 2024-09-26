// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cur_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccReportDefinitionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	dataSourceName := "data.aws_cur_report_definition.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionDataSourceConfig_basic(reportName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "report_name", resourceName, "report_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "time_unit", resourceName, "time_unit"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compression", resourceName, "compression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "additional_schema_elements.#", resourceName, "additional_schema_elements.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrS3Bucket, resourceName, names.AttrS3Bucket),
					resource.TestCheckResourceAttrPair(dataSourceName, "s3_prefix", resourceName, "s3_prefix"),
					resource.TestCheckResourceAttrPair(dataSourceName, "s3_region", resourceName, "s3_region"),
					resource.TestCheckResourceAttrPair(dataSourceName, "additional_artifacts.#", resourceName, "additional_artifacts.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsKey1, resourceName, acctest.CtTagsKey1),
				),
			},
		},
	})
}

func testAccReportDefinitionDataSource_additional(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	dataSourceName := "data.aws_cur_report_definition.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionDataSourceConfig_additional(reportName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "report_name", resourceName, "report_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "time_unit", resourceName, "time_unit"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compression", resourceName, "compression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "additional_schema_elements.#", resourceName, "additional_schema_elements.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrS3Bucket, resourceName, names.AttrS3Bucket),
					resource.TestCheckResourceAttrPair(dataSourceName, "s3_prefix", resourceName, "s3_prefix"),
					resource.TestCheckResourceAttrPair(dataSourceName, "s3_region", resourceName, "s3_region"),
					resource.TestCheckResourceAttrPair(dataSourceName, "additional_artifacts.#", resourceName, "additional_artifacts.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "refresh_closed_reports", resourceName, "refresh_closed_reports"),
					resource.TestCheckResourceAttrPair(dataSourceName, "report_versioning", resourceName, "report_versioning"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsKey1, resourceName, acctest.CtTagsKey1),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsKey2, resourceName, acctest.CtTagsKey2),
				),
			},
		},
	})
}

func testAccReportDefinitionDataSourceConfig_basic(reportName string, bucketName string) string {
	return fmt.Sprintf(`
data "aws_billing_service_account" "test" {}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_billing_service_account.test.arn}"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "${aws_s3_bucket.test.arn}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_billing_service_account.test.arn}"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = %[1]q
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES", "SPLIT_COST_ALLOCATION_DATA"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = ""
  s3_region                  = aws_s3_bucket.test.region
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
  tags = {
    key1 = "value1"
  }
}

data "aws_cur_report_definition" "test" {
  report_name = aws_cur_report_definition.test.report_name
}
`, reportName, bucketName)
}

func testAccReportDefinitionDataSourceConfig_additional(reportName string, bucketName string) string {
	return fmt.Sprintf(`
data "aws_billing_service_account" "test" {}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_billing_service_account.test.arn}"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "${aws_s3_bucket.test.arn}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_billing_service_account.test.arn}"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = %[1]q
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES", "SPLIT_COST_ALLOCATION_DATA"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = ""
  s3_region                  = aws_s3_bucket.test.region
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
  refresh_closed_reports     = true
  report_versioning          = "CREATE_NEW_REPORT"
  tags = {
    key1 = "value1"
    key2 = "value2"
  }
}

data "aws_cur_report_definition" "test" {
  report_name = aws_cur_report_definition.test.report_name
}
`, reportName, bucketName)
}
