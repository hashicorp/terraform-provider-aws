// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

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

func TestAccBackupRestoreTestingPlan_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan awstypes.RestoreTestingPlanForGet
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "cron(0 12 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0), // no tags
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan awstypes.RestoreTestingPlanForGet
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbackup.ResourceRestoreTestingPlan, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan awstypes.RestoreTestingPlanForGet
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccRestoreTestingPlanConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRestoreTestingPlanConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_includeVaults(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan awstypes.RestoreTestingPlanForGet
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_includeVaults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "recovery_point_selection.0.include_vaults.0", "backup", fmt.Sprintf("backup-vault:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "cron(0 12 ? * * *)"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_excludeVaults(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan awstypes.RestoreTestingPlanForGet
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_excludeVaults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "recovery_point_selection.0.exclude_vaults.0", "backup", fmt.Sprintf("backup-vault:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "cron(0 12 ? * * *)"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_additionals(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan awstypes.RestoreTestingPlanForGet
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_additionals("365", "cron(0 12 ? * * *)", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.selection_window_days", "365"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "cron(0 12 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, "start_window_hours", "168"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_additionalsWithUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan awstypes.RestoreTestingPlanForGet
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_additionals("365", "cron(0 1 ? * * *)", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.selection_window_days", "365"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "cron(0 1 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, "start_window_hours", "168"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccRestoreTestingPlanConfig_additionals(acctest.Ct1, "cron(0 12 ? * * *)", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.selection_window_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "cron(0 12 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, "start_window_hours", "168"),
				),
			},
		},
	})
}

func testAccCheckRestoreTestingPlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_restore_testing_plan" {
				continue
			}

			_, err := tfbackup.FindRestoreTestingPlanByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Restore Testing Plan %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckRestoreTestingPlanExists(ctx context.Context, n string, v *awstypes.RestoreTestingPlanForGet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		output, err := tfbackup.FindRestoreTestingPlanByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRestoreTestingPlanConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = ["*"]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00
}
`, rName)
}

func testAccRestoreTestingPlanConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = ["*"]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRestoreTestingPlanConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = ["*"]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRestoreTestingPlanConfig_additionals(selectionWindowDays, scheduleExpression, rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[3]q

  recovery_point_selection {
    algorithm             = "LATEST_WITHIN_WINDOW"
    include_vaults        = ["*"]
    recovery_point_types  = ["CONTINUOUS", "SNAPSHOT"]
    selection_window_days = %[1]s
  }

  schedule_expression = %[2]q
  start_window_hours  = 168
}
`, selectionWindowDays, scheduleExpression, rName)
}

func testAccRestoreTestingPlanConfig_baseVaults(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  enable_key_rotation     = true
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_backup_vault" "test" {
  name        = %[1]q
  kms_key_arn = aws_kms_key.test.arn
}
`, rName)
}

func testAccRestoreTestingPlanConfig_includeVaults(rName string) string {
	return acctest.ConfigCompose(
		testAccRestoreTestingPlanConfig_baseVaults(rName),
		fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = [resource.aws_backup_vault.test.arn]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00
}
`, rName))
}

func testAccRestoreTestingPlanConfig_excludeVaults(rName string) string {
	return acctest.ConfigCompose(
		testAccRestoreTestingPlanConfig_baseVaults(rName),
		fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = ["*"]
    exclude_vaults       = [resource.aws_backup_vault.test.arn]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00
}
`, rName))
}
