// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupReportPlan_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var reportPlan awstypes.ReportPlan
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_backup_report_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccReportPlanPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
				),
			},
		},
	})
}

func TestAccBackupReportPlan_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var reportPlan awstypes.ReportPlan
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
	description := "example description"
	resourceName := "aws_backup_report_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccReportPlanPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
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
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func TestAccBackupReportPlan_updateReportDeliveryChannel(t *testing.T) {
	ctx := acctest.Context(t)
	var reportPlan awstypes.ReportPlan
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
	description := "example description"
	resourceName := "aws_backup_report_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccReportPlanPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "report_delivery_channel.0.formats.*", "CSV"),
					resource.TestCheckTypeSetElemAttr(resourceName, "report_delivery_channel.0.formats.*", "JSON"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
				),
			},
		},
	})
}

func TestAccBackupReportPlan_updateReportSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var reportPlan awstypes.ReportPlan
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
	description := "example description"
	resourceName := "aws_backup_report_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccReportPlanPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "report_setting.0.accounts"),
					resource.TestCheckNoResourceAttr(resourceName, "report_setting.0.regions"),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReportPlanConfig_reportSettings(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_status"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_delivery_channel.0.formats.0", "CSV"),
					resource.TestCheckResourceAttrPair(resourceName, "report_delivery_channel.0.s3_bucket_name", "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.accounts.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "report_setting.0.accounts.0", "data.aws_caller_identity.current", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.regions.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "report_setting.0.regions.0", "data.aws_region.current", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "report_setting.0.report_template", "RESTORE_JOB_REPORT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Report Plan"),
				),
			},
		},
	})
}

func TestAccBackupReportPlan_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var reportPlan awstypes.ReportPlan
	rName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	rName2 := fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
	resourceName := "aws_backup_report_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccReportPlanPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportPlanConfig_basic(rName, rName2, acctest.CtDisappears),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportPlanExists(ctx, resourceName, &reportPlan),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfbackup.ResourceReportPlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccReportPlanPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

	_, err := conn.ListReportPlans(ctx, &backup.ListReportPlansInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckReportPlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_report_plan" {
				continue
			}

			_, err := tfbackup.FindReportPlanByName(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckReportPlanExists(ctx context.Context, n string, v *awstypes.ReportPlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Backup Report Plan ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		output, err := tfbackup.FindReportPlanByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccReportPlanBaseConfig(bucketName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

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

func testAccReportPlanConfig_reportSettings(rName, rName2, label string) string {
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
    accounts = [
      data.aws_caller_identity.current.id
    ]
    regions = [
      data.aws_region.current.name
    ]
    report_template = "RESTORE_JOB_REPORT"
  }

  tags = {
    "Name" = "Test Report Plan"
  }
}
`, rName2, label))
}
