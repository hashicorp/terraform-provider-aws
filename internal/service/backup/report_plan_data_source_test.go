// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupReportPlanDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_backup_report_plan.test"
	resourceName := "aws_backup_report_plan.test"
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccReportPlanPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccReportPlanDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`reading Backup Report Plan`),
			},
			{
				Config: testAccReportPlanDataSourceConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCreationTime, resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(datasourceName, "deployment_status"), // CREATE_IN_PROGRESS | UPDATE_IN_PROGRESS | DELETE_IN_PROGRESS | COMPLETED
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.#", resourceName, "report_delivery_channel.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.0.formats.#", resourceName, "report_delivery_channel.0.formats.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.0.formats.0", resourceName, "report_delivery_channel.0.formats.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.0.s3_bucket_name", resourceName, "report_delivery_channel.0.s3_bucket_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_setting.#", resourceName, "report_setting.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_setting.0.report_template", resourceName, "report_setting.0.report_template"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key2", resourceName, "tags.Key2"),
				),
			},
		},
	})
}

func TestAccBackupReportPlanDataSource_reportSettings(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_backup_report_plan.test"
	resourceName := "aws_backup_report_plan.test"
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccReportPlanPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanDataSourceConfig_reportSettings(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCreationTime, resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(datasourceName, "deployment_status"), // CREATE_IN_PROGRESS | UPDATE_IN_PROGRESS | DELETE_IN_PROGRESS | COMPLETED
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.#", resourceName, "report_delivery_channel.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.0.formats.#", resourceName, "report_delivery_channel.0.formats.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.0.formats.0", resourceName, "report_delivery_channel.0.formats.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.0.s3_bucket_name", resourceName, "report_delivery_channel.0.s3_bucket_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_setting.#", resourceName, "report_setting.#"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.accounts.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "report_setting.0.accounts.0", "data.aws_caller_identity.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.regions.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "report_setting.0.regions.0", "data.aws_region.current", names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "report_setting.0.report_template", resourceName, "report_setting.0.report_template"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key2", resourceName, "tags.Key2"),
				),
			},
		},
	})
}

const testAccReportPlanDataSourceConfig_nonExistent = `
data "aws_backup_report_plan" "test" {
  name = "tf_acc_test_does_not_exist"
}
`

func testAccReportPlanDataSourceConfig_basic(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccReportPlanBaseConfig(rName), fmt.Sprintf(`
resource "aws_backup_report_plan" "test" {
  name        = %[1]q
  description = "Test report plan data source"

  report_delivery_channel {
    formats = [
      "CSV"
    ]
    s3_bucket_name = aws_s3_bucket.test.id
  }

  report_setting {
    report_template = "RESTORE_JOB_REPORT"
  }

  tags = {
    "Name" = "Test Report Plan Data Source"
    "Key2" = "Value2a"
  }
}

data "aws_backup_report_plan" "test" {
  name = aws_backup_report_plan.test.name
}
`, rName2))
}

func testAccReportPlanDataSourceConfig_reportSettings(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccReportPlanConfig_reportSettings(rName, rName2, "Test report plan data source"),
		`
data "aws_backup_report_plan" "test" {
  name = aws_backup_report_plan.test.name
}
`)
}
