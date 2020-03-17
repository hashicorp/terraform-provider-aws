package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceAwsCurReportDefinition_basic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resourceName := "aws_cur_report_definition.test"
	datasourceName := "data.aws_cur_report_definition.test"

	reportName := acctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	bucketRegion := "us-east-1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCur(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_bucket(bucketName, bucketRegion),
			},
			{
				Config: testAccDataSourceAwsCurReportDefinitionConfig_basic(reportName, bucketName, bucketRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsCurReportDefinitionCheckExists(datasourceName, resourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "report_name", resourceName, "report_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "time_unit", resourceName, "time_unit"),
					resource.TestCheckResourceAttrPair(datasourceName, "compression", resourceName, "compression"),
					resource.TestCheckResourceAttrPair(datasourceName, "additional_schema_elements.#", resourceName, "additional_schema_elements.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "s3_bucket", resourceName, "s3_bucket"),
					resource.TestCheckResourceAttrPair(datasourceName, "s3_prefix", resourceName, "s3_prefix"),
					resource.TestCheckResourceAttrPair(datasourceName, "s3_region", resourceName, "s3_region"),
					resource.TestCheckResourceAttrPair(datasourceName, "additional_artifacts.#", resourceName, "additional_artifacts.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "refresh_closed_reports", resourceName, "refresh_closed_reports"),
					resource.TestCheckResourceAttrPair(datasourceName, "report_versioning", resourceName, "report_versioning"),
				),
			},
		},
	})
}

func testAccDataSourceAwsCurReportDefinitionCheckExists(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no data source called %s", datasourceName)
		}
		_, ok = s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}
		return nil
	}
}

// note: cur report definitions are currently only supported in us-east-1
func testAccDataSourceAwsCurReportDefinitionConfig_basic(reportName string, bucketName string, bucketRegion string) string {
	return fmt.Sprintf(`
%[2]s

resource "aws_cur_report_definition" "test" {
  report_name                = "%[1]s"
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES"]
  s3_bucket                  = "${aws_s3_bucket.test.id}"
  s3_prefix                  = ""
  s3_region                  = "${aws_s3_bucket.test.region}"
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
	refresh_closed_reports     = true
	report_versioning          = "CREATE_NEW_REPORT"
}

data "aws_cur_report_definition" "test" {
  report_name = "${aws_cur_report_definition.test.report_name}"
}
`, reportName, testAccAwsCurReportDefinitionConfig_bucket(bucketName, bucketRegion))
}
