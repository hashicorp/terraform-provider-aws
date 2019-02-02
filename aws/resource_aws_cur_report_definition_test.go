package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsCurReportDefinition_basic(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"

	reportName := acctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	bucketRegion := "us-east-1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCurReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCurReportDefinitionConfig_basic(reportName, bucketName, bucketRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCurReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "s3_region", bucketRegion),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
				),
			},
		},
	})
}

func testAccCheckAwsCurReportDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).costandusagereportconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cur_report_definition" {
			continue
		}
		reportName := rs.Primary.ID
		params := &costandusagereportservice.DescribeReportDefinitionsInput{}
		var matchingReportDefinition *costandusagereportservice.ReportDefinition
		err := conn.DescribeReportDefinitionsPages(params, func(resp *costandusagereportservice.DescribeReportDefinitionsOutput, isLast bool) bool {
			for _, reportDefinition := range resp.ReportDefinitions {
				if *reportDefinition.ReportName == reportName {
					matchingReportDefinition = reportDefinition
					break
				}
			}
			return !isLast
		})
		if err != nil {
			return err
		}
		if matchingReportDefinition != nil {
			return fmt.Errorf("Report Definition still exists: %q", rs.Primary.ID)
		}
	}
	return nil

}

func testAccCheckAwsCurReportDefinitionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).costandusagereportconn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}
		reportName := rs.Primary.ID
		params := &costandusagereportservice.DescribeReportDefinitionsInput{}
		var matchingReportDefinition *costandusagereportservice.ReportDefinition
		err := conn.DescribeReportDefinitionsPages(params, func(resp *costandusagereportservice.DescribeReportDefinitionsOutput, isLast bool) bool {
			for _, reportDefinition := range resp.ReportDefinitions {
				if *reportDefinition.ReportName == reportName {
					matchingReportDefinition = reportDefinition
					break
				}
			}
			return !isLast
		})
		if err != nil {
			return err
		}
		if matchingReportDefinition == nil {
			return fmt.Errorf("Report Definition does not exist: %q", rs.Primary.ID)
		}
		return nil
	}
}

// note: cur report definitions are currently only supported in us-east-1
func testAccAwsCurReportDefinitionConfig_basic(reportName string, bucketName string, bucketRegion string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_s3_bucket" "test" {
    bucket = "%[2]s"
	acl = "private"
	force_destroy = true
    region = "%[3]s"
}

resource "aws_s3_bucket_policy" "test" {
  bucket = "${aws_s3_bucket.test.id}"
  policy = <<POLICY
{
    "Version": "2008-10-17",
    "Id": "s3policy",
    "Statement": [
        {
            "Sid": "AllowCURBillingACLPolicy",
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::386209384616:root"
            },
            "Action": [
                "s3:GetBucketAcl",
                "s3:GetBucketPolicy"
            ],
            "Resource": "arn:aws:s3:::${aws_s3_bucket.test.id}"
        },
        {
            "Sid": "AllowCURPutObject",
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::386209384616:root"
            },
            "Action": "s3:PutObject",
            "Resource": "arn:aws:s3:::${aws_s3_bucket.test.id}/*"
        }
    ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
    report_name = "%[1]s"
    time_unit = "DAILY"
    format = "textORcsv"
    compression = "GZIP"
    additional_schema_elements = ["RESOURCES"]
    s3_bucket = "${aws_s3_bucket.test.id}"
    s3_prefix = ""
    s3_region = "${aws_s3_bucket.test.region}"
	additional_artifacts = ["REDSHIFT", "QUICKSIGHT"]
}
`, reportName, bucketName, bucketRegion)
}
