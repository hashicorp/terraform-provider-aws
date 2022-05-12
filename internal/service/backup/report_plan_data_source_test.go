package backup_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBackupReportPlanDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_backup_report_plan.test"
	resourceName := "aws_backup_report_plan.test"
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccReportPlanPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccReportPlanDataSourceNonExistentConfig,
				ExpectError: regexp.MustCompile(`error reading Backup Report Plan`),
			},
			{
				Config: testAccReportPlanDataSourceConfig(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "creation_time", resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(datasourceName, "deployment_status"), // CREATE_IN_PROGRESS | UPDATE_IN_PROGRESS | DELETE_IN_PROGRESS | COMPLETED
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.#", resourceName, "report_delivery_channel.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.0.formats.#", resourceName, "report_delivery_channel.0.formats.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.0.formats.0", resourceName, "report_delivery_channel.0.formats.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_delivery_channel.0.s3_bucket_name", resourceName, "report_delivery_channel.0.s3_bucket_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_setting.#", resourceName, "report_setting.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_setting.0.report_template", resourceName, "report_setting.0.report_template"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key2", resourceName, "tags.Key2"),
				),
			},
		},
	})
}

const testAccReportPlanDataSourceNonExistentConfig = `
data "aws_backup_report_plan" "test" {
  name = "tf_acc_test_does_not_exist"
}
`

func testAccReportPlanDataSourceConfig(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket                  = aws_s3_bucket.test.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_backup_report_plan" "test" {
  name        = %[2]q
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
`, rName, rName2)
}
