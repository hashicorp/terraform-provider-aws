package backup_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccBackupReportPlan_basic(t *testing.T) {
	var reportPlan backup.ReportPlan
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_backup_report_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccReportPlanPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
				),
			},
		},
	})
}

func TestAccBackupReportPlan_updateTags(t *testing.T) {
	var reportPlan backup.ReportPlan
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
	description := "example description"
	resourceName := "aws_backup_report_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccReportPlanPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReportPlanConfig_tags1(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReportPlanConfig_tags2(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func TestAccBackupReportPlan_updateReportDeliveryChannel(t *testing.T) {
	var reportPlan backup.ReportPlan
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
	description := "example description"
	resourceName := "aws_backup_report_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccReportPlanPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReportPlanConfig_deliveryChannel(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "report_delivery_channel.0.formats.*", "CSV"),
					resource.TestCheckTypeSetElemAttr(resourceName, "report_delivery_channel.0.formats.*", "JSON"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
				),
			},
		},
	})
}

func TestAccBackupReportPlan_disappears(t *testing.T) {
	var reportPlan backup.ReportPlan
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
	description := "disappears"
	resourceName := "aws_backup_report_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccReportPlanPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, backup.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(resourceName, &reportPlan),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourceReportPlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccReportPlanPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

	_, err := conn.ListReportPlans(&backup.ListReportPlansInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckReportPlanDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_report_plan" {
			continue
		}

		_, err := tfbackup.FindReportPlanByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Backup Report Plan %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckReportPlanExists(n string, v *backup.ReportPlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Backup Report Plan ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

		output, err := tfbackup.FindReportPlanByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccReportPlanBaseConfig(bucketName string) string {
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
`, bucketName)
}

func testAccReportPlanConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(testAccReportPlanBaseConfig(rName), fmt.Sprintf(`
resource "aws_backup_report_plan" "test" {
  name        = %[1]q
  description = %[2]q

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
    "Name" = "Test Report Plan"
  }
}
`, rName2, label))
}

func testAccReportPlanConfig_tags1(rName, rName2, label string) string {
	return acctest.ConfigCompose(testAccReportPlanBaseConfig(rName), fmt.Sprintf(`
resource "aws_backup_report_plan" "test" {
  name        = %[1]q
  description = %[2]q

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
    "Name" = "Test Report Plan"
    "Key2" = "Value2a"
  }
}
`, rName2, label))
}

func testAccReportPlanConfig_tags2(rName, rName2, label string) string {
	return acctest.ConfigCompose(testAccReportPlanBaseConfig(rName), fmt.Sprintf(`
resource "aws_backup_report_plan" "test" {
  name        = %[1]q
  description = %[2]q

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
    "Name" = "Test Report Plan"
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName2, label))
}

func testAccReportPlanConfig_deliveryChannel(rName, rName2, label string) string {
	return acctest.ConfigCompose(testAccReportPlanBaseConfig(rName), fmt.Sprintf(`
resource "aws_backup_report_plan" "test" {
  name        = %[1]q
  description = %[2]q

  report_delivery_channel {
    formats = [
      "CSV",
      "JSON"
    ]
    s3_bucket_name = aws_s3_bucket.test.id
  }

  report_setting {
    report_template = "RESTORE_JOB_REPORT"
  }

  tags = {
    "Name" = "Test Report Plan"
  }
}
`, rName2, label))
}
