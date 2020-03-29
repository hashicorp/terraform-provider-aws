package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSCodeBuildReportGroup_basic(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "false"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`report-group/.+`)),
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

func TestAccAWSCodeBuildReportGroup_updated(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupFullConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.path", "/some"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`report-group/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildReportGroupUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "ZIP"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.path", "/some2"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`report-group/.+`)),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildReportGroup_disappears(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupExists(resourceName, &reportGroup),
					testAccCheckAWSCodeBuildReportGroupDisappears(&reportGroup),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCodeBuildReportGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_report_group" {
			continue
		}

		resp, err := conn.BatchGetReportGroups(&codebuild.BatchGetReportGroupsInput{
			ReportGroupArns: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}

		if len(resp.ReportGroups) == 0 {
			return nil
		}

		for _, reportGroup := range resp.ReportGroups {
			if rs.Primary.ID == aws.StringValue(reportGroup.Arn) {
				return fmt.Errorf("Found Report Groups %s", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAWSCodeBuildReportGroupExists(name string, reportGroup *codebuild.ReportGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).codebuildconn

		resp, err := conn.BatchGetReportGroups(&codebuild.BatchGetReportGroupsInput{
			ReportGroupArns: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}

		if len(resp.ReportGroups) != 1 ||
			aws.StringValue(resp.ReportGroups[0].Arn) != rs.Primary.ID {
			return fmt.Errorf("Report Group %s not found", rs.Primary.ID)
		}

		*reportGroup = *resp.ReportGroups[0]

		return nil
	}
}

func testAccCheckAWSCodeBuildReportGroupDisappears(reportGroup *codebuild.ReportGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).codebuildconn

		_, err := conn.DeleteReportGroup(&codebuild.DeleteReportGroupInput{
			Arn: reportGroup.Arn,
		})

		return err
	}
}

func testAccAWSCodeBuildReportGroupBasicConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccAWSCodeBuildReportGroupBasicConfig(rName string) string {
	return testAccAWSCodeBuildReportGroupBasicConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "S3"

    s3_destination {
      bucket              = "${aws_s3_bucket.test.id}"
      encryption_key      = "${aws_kms_key.test.arn}"
    }
  }
}
`, rName)
}

func testAccAWSCodeBuildReportGroupFullConfig(rName string) string {
	return testAccAWSCodeBuildReportGroupBasicConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "S3"

    s3_destination {
      bucket              = "${aws_s3_bucket.test.id}"
      encryption_disabled = false
      encryption_key      = "${aws_kms_key.test.arn}"
      packaging           = "NONE"
      path                = "/some"
    }
  }
}
`, rName)
}

func testAccAWSCodeBuildReportGroupUpdatedConfig(rName string) string {
	return testAccAWSCodeBuildReportGroupBasicConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "S3"

    s3_destination {
      bucket              = "${aws_s3_bucket.test.id}"
      encryption_disabled = true
      encryption_key      = "${aws_kms_key.test.arn}"
      packaging           = "ZIP"
      path                = "/some2"
    }
  }
}
`, rName)
}
