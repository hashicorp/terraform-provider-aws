// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
)

func TestAccBackupRestoreTestingPlan_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan backup.GetRestoreTestingPlanOutput
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
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 12 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"), // no tags
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan backup.GetRestoreTestingPlanOutput
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
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbackup.RestoreTestingPlanResource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan backup.GetRestoreTestingPlanOutput
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
				Config: testAccRestoreTestingPlanConfig_tags("Name", "RestoreTestingPlan", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 12 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "RestoreTestingPlan"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
			{
				Config: testAccRestoreTestingPlanConfig_tags("Name", "Testing1", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 12 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Testing1"),
				),
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_includeVaults(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan backup.GetRestoreTestingPlanOutput
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
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "recovery_point_selection.0.include_vaults.0", "backup", fmt.Sprintf("backup-vault:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 12 ? * * *)"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_excludeVaults(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan backup.GetRestoreTestingPlanOutput
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
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", "1"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "recovery_point_selection.0.exclude_vaults.0", "backup", fmt.Sprintf("backup-vault:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 12 ? * * *)"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_additionals(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan backup.GetRestoreTestingPlanOutput
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
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.selection_window_days", "365"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 12 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, "start_window_hours", "168"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBackupRestoreTestingPlan_additionalwithupdates(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan backup.GetRestoreTestingPlanOutput
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
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.selection_window_days", "365"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 1 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, "start_window_hours", "168"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        rName,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
			{
				Config: testAccRestoreTestingPlanConfig_additionals("1", "cron(0 12 ? * * *)", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.selection_window_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 12 ? * * *)"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression_timezone", "Europe/London"),
					resource.TestCheckResourceAttr(resourceName, "start_window_hours", "12"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "RestoreTestingPlan"),
				),
			},
		},
	})
}

func testAccCheckRestoreTestingPlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_restore_testing_plan" {
				continue
			}

			if rs.Primary.Attributes["name"] == "" {
				return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameRestoreTestingPlan, "unknown", errors.New("not set"))
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)
			_, err := tfbackup.FindRestoreTestingPlanByName(ctx, conn, rs.Primary.Attributes["name"])
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Backup, create.ErrActionCheckingDestroyed, tfbackup.ResNameRestoreTestingPlan, rs.Primary.Attributes["name"], err)
			}

			return create.Error(names.Backup, create.ErrActionCheckingDestroyed, tfbackup.ResNameRestoreTestingPlan, rs.Primary.Attributes["name"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRestoreTestingPlanExists(ctx context.Context, name string, restoretestingplan *backup.GetRestoreTestingPlanOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameRestoreTestingPlan, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["name"] == "" {
			return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameRestoreTestingPlan, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)
		resp, err := tfbackup.FindRestoreTestingPlanByName(ctx, conn, rs.Primary.Attributes["name"])

		if err != nil {
			return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameRestoreTestingPlan, name, err)
		}

		*restoretestingplan = *resp

		return nil
	}
}

func testAccRestoreTestingPlanConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  enable_key_rotation = true
}

resource "aws_backup_vault" "test" {
  name        = "%[1]s"
  kms_key_arn = aws_kms_key.test.arn
}
`, rName)
}

func testAccRestoreTestingPlanConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = "%[1]s"

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = ["*"]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00
}
`, rName)
}

func testAccRestoreTestingPlanConfig_tags(tagName, tagValue, rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = "%[3]s"

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = ["*"]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00

  tags = {
    "%[1]s" = "%[2]s"
  }
}
`, tagName, tagValue, rName)
}

func testAccRestoreTestingPlanConfig_additionals(selectionWindowDays, scheduleExpression, rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = "%[3]s"

  recovery_point_selection {
    algorithm             = "LATEST_WITHIN_WINDOW"
    include_vaults        = ["*"]
    recovery_point_types  = ["CONTINUOUS", "SNAPSHOT"]
    selection_window_days = %[1]s
  }

  schedule_expression = "%[2]s"
  start_window_hours  = 168
}
`, selectionWindowDays, scheduleExpression, rName)
}

func testAccRestoreTestingPlanConfig_includeVaults(rName string) string {
	return acctest.ConfigCompose(
		testAccRestoreTestingPlanConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = "%[1]s"

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = [resource.aws_backup_vault.test.arn]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00
}
`, rName),
	)
}

func testAccRestoreTestingPlanConfig_excludeVaults(rName string) string {
	return acctest.ConfigCompose(
		testAccRestoreTestingPlanConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = "%[1]s"

  recovery_point_selection {
    algorithm            = "LATEST_WITHIN_WINDOW"
    include_vaults       = ["*"]
    exclude_vaults       = [resource.aws_backup_vault.test.arn]
    recovery_point_types = ["CONTINUOUS"]
  }

  schedule_expression = "cron(0 12 ? * * *)" # Daily at 12:00
}
`, rName),
	)
}
