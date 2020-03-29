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

func TestAccAWSCodeBuildReportGroups_basic(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupsBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupsExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "false"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`report-group/`)),
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

func TestAccAWSCodeBuildReportGroups_updated(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_report_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildReportGroupsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildReportGroupsFullConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupsExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.path", "/some"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`report-group/`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildReportGroupsUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildReportGroupsExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "ZIP"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.path", "/some2"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`report-group/`)),
				),
			},
		},
	})
}

func testAccCheckAWSCodeBuildReportGroupsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_report_groups" {
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

func testAccCheckAWSCodeBuildReportGroupsExists(name string, ReportGroups *codebuild.ReportGroup) resource.TestCheckFunc {
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

		if len(resp.ReportGroups) == 0 {
			return fmt.Errorf("Source Credential %s not found", rs.Primary.ID)
		}

		for _, reportGroup := range resp.ReportGroups {
			if rs.Primary.ID == aws.StringValue(reportGroup.Arn) {
				*ReportGroups = *reportGroup
				return nil
			}
		}

		return fmt.Errorf("Report Groups %s not found", rs.Primary.ID)
	}
}

func testAccAWSCodeBuildReportGroupsBasicConfigBase(rName string) string {
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

func testAccAWSCodeBuildReportGroupsBasicConfig(rName string) string {
	return testAccAWSCodeBuildReportGroupsBasicConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_groups" "test" {
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

func testAccAWSCodeBuildReportGroupsFullConfig(rName string) string {
	return testAccAWSCodeBuildReportGroupsBasicConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_groups" "test" {
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

func testAccAWSCodeBuildReportGroupsUpdatedConfig(rName string) string {
	return testAccAWSCodeBuildReportGroupsBasicConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_groups" "test" {
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
