// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cur_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccReportDefinitionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	dataSourceName := "data.aws_cur_report_definition.test"
	reportName := acctest.RandomWithPrefix(t, "tf_acc_test")
	s3BucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt(t))
	s3Prefix := "test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionDataSourceConfig_basic(reportName, s3BucketName, s3Prefix),
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
	reportName := acctest.RandomWithPrefix(t, "tf_acc_test")
	s3BucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt(t))
	s3Prefix := "test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionDataSourceConfig_additional(reportName, s3BucketName, s3Prefix),
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

func testAccReportDefinitionDataSourceConfig_basic(reportName, s3BucketName, s3Prefix string) string {
	return acctest.ConfigCompose(testAccReportDefinitionConfig_base(s3BucketName), fmt.Sprintf(`
resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = %[1]q
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES", "SPLIT_COST_ALLOCATION_DATA"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = %[2]q
  s3_region                  = aws_s3_bucket.test.region
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
  tags = {
    key1 = "value1"
  }
}

data "aws_cur_report_definition" "test" {
  report_name = aws_cur_report_definition.test.report_name
}
`, reportName, s3Prefix))
}

func testAccReportDefinitionDataSourceConfig_additional(reportName, s3BucketName, s3Prefix string) string {
	return acctest.ConfigCompose(testAccReportDefinitionConfig_base(s3BucketName), fmt.Sprintf(`
resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = %[1]q
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES", "SPLIT_COST_ALLOCATION_DATA"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = %[2]q
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
`, reportName, s3Prefix))
}
