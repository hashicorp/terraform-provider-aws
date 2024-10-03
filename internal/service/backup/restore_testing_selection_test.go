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
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupRestoreTestingSelection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan backup.GetRestoreTestingSelectionOutput
	resourceName := "aws_backup_restore_testing_selection.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanSelection(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingSelectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingSelectionExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "restore_testing_plan_name", rName),
					resource.TestCheckResourceAttr(resourceName, "protected_resource_type", "EC2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        fmt.Sprintf("%s:%s", rName, rName),
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBackupRestoreTestingSelection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingselection backup.GetRestoreTestingSelectionOutput
	resourceName := "aws_backup_restore_testing_selection.test"
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
				Config: testAccRestoreTestingSelectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingSelectionExists(ctx, resourceName, &restoretestingselection),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbackup.RestoreTestingSelectionResource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupRestoreTestingSelection_updates(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan backup.GetRestoreTestingSelectionOutput
	resourceName := "aws_backup_restore_testing_selection.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanSelection(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingSelectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingSelectionExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "restore_testing_plan_name", rName),
					resource.TestCheckResourceAttr(resourceName, "protected_resource_type", "EC2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        fmt.Sprintf("%s:%s", rName, rName),
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
			{
				Config: testAccRestoreTestingSelectionConfig_updates(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingSelectionExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "restore_testing_plan_name", rName),
					resource.TestCheckResourceAttr(resourceName, "protected_resource_type", "EC2"),
				),
			},
		},
	})
}

func testAccCheckRestoreTestingPlanSelection(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_restore_testing_selection" {
				continue
			}

			if rs.Primary.Attributes["name"] == "" {
				return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameRestoreTestingSelection, "unknown", errors.New("not set"))
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)
			_, err := tfbackup.FindRestoreTestingSelectionByName(ctx, conn, rs.Primary.Attributes["name"], rs.Primary.Attributes["restore_testing_plan_name"])
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Backup, create.ErrActionCheckingDestroyed, tfbackup.ResNameRestoreTestingSelection, rs.Primary.Attributes["name"], err)
			}

			return create.Error(names.Backup, create.ErrActionCheckingDestroyed, tfbackup.ResNameRestoreTestingSelection, rs.Primary.Attributes["name"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRestoreTestingSelectionExists(ctx context.Context, name string, restoretestingplan *backup.GetRestoreTestingSelectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameRestoreTestingSelection, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["name"] == "" {
			return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameRestoreTestingSelection, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)
		resp, err := tfbackup.FindRestoreTestingSelectionByName(ctx, conn, rs.Primary.Attributes["name"], rs.Primary.Attributes["restore_testing_plan_name"])

		if err != nil {
			return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameRestoreTestingSelection, name, err)
		}

		*restoretestingplan = *resp

		return nil
	}
}

func testAccRestoreTestingSelectionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%[1]s"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_kms_key" "test" {
  enable_key_rotation = true
}

resource "aws_kms_alias" "a" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.key_id
}

resource "aws_backup_vault" "test" {
  name        = "%[1]s"
  kms_key_arn = aws_kms_key.test.arn
}

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

func testAccRestoreTestingSelectionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccRestoreTestingSelectionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_restore_testing_selection" "test" {
  name = "%[1]s"

  restore_testing_plan_name = aws_backup_restore_testing_plan.test.name
  protected_resource_type   = "EC2"
  iam_role_arn              = aws_iam_role.test.arn

  protected_resource_conditions {
  }
}
`, rName),
	)
}

func testAccRestoreTestingSelectionConfig_updates(rName string) string {
	return acctest.ConfigCompose(
		testAccRestoreTestingSelectionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_restore_testing_selection" "test" {
  name = "%[1]s"

  restore_testing_plan_name = aws_backup_restore_testing_plan.test.name
  protected_resource_type   = "EC2"
  iam_role_arn              = aws_iam_role.test.arn

  protected_resource_conditions {
    string_equals = [
      {
        key   = "aws:ResourceTag/backup"
        value = true
      }
    ]
  }

  validation_window_hours = 10

  restore_metadata_overrides = {
    instanceType = "t2.micro"
  }
}
`, rName),
	)
}
