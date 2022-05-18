package codebuild_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codebuild"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodebuild "github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
)

func TestAccCodeBuildReportGroup_basic(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckReportGroup(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "NO_EXPORT"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codebuild", fmt.Sprintf("report-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_reports"},
			},
		},
	})
}

func TestAccCodeBuildReportGroup_Export_s3(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckReportGroup(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportGroupS3ExportConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.path", "/some"),
					resource.TestCheckResourceAttrPair(resourceName, "export_config.0.s3_destination.0.encryption_key", "aws_kms_key.test", "arn"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codebuild", fmt.Sprintf("report-group/%s", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_reports"},
			},
			{
				Config: testAccReportGroupS3ExportUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "export_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.packaging", "ZIP"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.encryption_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "export_config.0.s3_destination.0.path", "/some2"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codebuild", fmt.Sprintf("report-group/%s", rName)),
				),
			},
		},
	})
}

func TestAccCodeBuildReportGroup_tags(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckReportGroup(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportGroupTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_reports"},
			},
			{
				Config: testAccReportGroupTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccReportGroupTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCodeBuildReportGroup_deleteReports(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckReportGroup(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportGroupDeleteReportsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportGroupExists(resourceName, &reportGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_reports"},
			},
		},
	})
}

func TestAccCodeBuildReportGroup_disappears(t *testing.T) {
	var reportGroup codebuild.ReportGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_report_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckReportGroup(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codebuild.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportGroupExists(resourceName, &reportGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodebuild.ResourceReportGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckReportGroup(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

	input := &codebuild.ListReportGroupsInput{}

	_, err := conn.ListReportGroups(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckReportGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_report_group" {
			continue
		}

		resp, err := tfcodebuild.FindReportGroupByARN(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("Found Report Group %s", rs.Primary.ID)
		}

	}
	return nil
}

func testAccCheckReportGroupExists(name string, reportGroup *codebuild.ReportGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildConn

		resp, err := tfcodebuild.FindReportGroupByARN(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("Report Group %s not found", rs.Primary.ID)
		}

		*reportGroup = *resp

		return nil
	}
}

func testAccReportGroupBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }
}
`, rName)
}

func testAccReportGroupBasicS3ExportBaseConfig(rName string) string {
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

func testAccReportGroupS3ExportConfig(rName string) string {
	return testAccReportGroupBasicS3ExportBaseConfig(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "S3"

    s3_destination {
      bucket              = aws_s3_bucket.test.id
      encryption_disabled = false
      encryption_key      = aws_kms_key.test.arn
      packaging           = "NONE"
      path                = "/some"
    }
  }
}
`, rName)
}

func testAccReportGroupS3ExportUpdatedConfig(rName string) string {
	return testAccReportGroupBasicS3ExportBaseConfig(rName) +
		fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "S3"

    s3_destination {
      bucket              = aws_s3_bucket.test.id
      encryption_key      = aws_kms_key.test.arn
      encryption_disabled = false
      packaging           = "ZIP"
      path                = "/some2"
    }
  }
}
`, rName)
}

func testAccReportGroupTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccReportGroupTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name = %[1]q
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccReportGroupDeleteReportsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_report_group" "test" {
  name           = %[1]q
  type           = "TEST"
  delete_reports = true

  export_config {
    type = "NO_EXPORT"
  }
}
`, rName)
}
