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

func TestAccBackupRestoreTestingSelection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan awstypes.RestoreTestingSelectionForGet
	resourceName := "aws_backup_restore_testing_selection.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingSelectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingSelectionExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "restore_testing_plan_name", rName+"_plan"),
					resource.TestCheckResourceAttr(resourceName, "protected_resource_type", "EC2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        fmt.Sprintf("%s:%s", rName, rName+"_plan"),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccBackupRestoreTestingSelection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingselection awstypes.RestoreTestingSelectionForGet
	resourceName := "aws_backup_restore_testing_selection.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingSelectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingSelectionExists(ctx, resourceName, &restoretestingselection),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbackup.ResourceRestoreTestingSelection, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupRestoreTestingSelection_updates(t *testing.T) {
	ctx := acctest.Context(t)
	var restoretestingplan awstypes.RestoreTestingSelectionForGet
	resourceName := "aws_backup_restore_testing_selection.test"
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingSelectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingSelectionExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "restore_testing_plan_name", rName+"_plan"),
					resource.TestCheckResourceAttr(resourceName, "protected_resource_type", "EC2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        fmt.Sprintf("%s:%s", rName, rName+"_plan"),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{names.AttrApplyImmediately, "user"},
			},
			{
				Config: testAccRestoreTestingSelectionConfig_updates(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingSelectionExists(ctx, resourceName, &restoretestingplan),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "restore_testing_plan_name", rName+"_plan"),
					resource.TestCheckResourceAttr(resourceName, "protected_resource_type", "EC2"),
				),
			},
		},
	})
}

func testAccCheckRestoreTestingSelectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_restore_testing_selection" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

			_, err := tfbackup.FindRestoreTestingSelectionByTwoPartKey(ctx, conn, rs.Primary.Attributes["restore_testing_plan_name"], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Restore Testing Selection %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckRestoreTestingSelectionExists(ctx context.Context, n string, v *awstypes.RestoreTestingSelectionForGet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		output, err := tfbackup.FindRestoreTestingSelectionByTwoPartKey(ctx, conn, rs.Primary.Attributes["restore_testing_plan_name"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRestoreTestingSelectionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

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
  enable_key_rotation     = true
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "a" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.key_id
}

resource "aws_backup_vault" "test" {
  name        = %[1]q
  kms_key_arn = aws_kms_key.test.arn
}

resource "aws_backup_restore_testing_plan" "test" {
  name = "%[1]s_plan"

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
  name = %[1]q

  restore_testing_plan_name = aws_backup_restore_testing_plan.test.name
  protected_resource_type   = "EC2"
  iam_role_arn              = aws_iam_role.test.arn

  protected_resource_arns = ["*"]
}
`, rName))
}

func testAccRestoreTestingSelectionConfig_updates(rName string) string {
	return acctest.ConfigCompose(
		testAccRestoreTestingSelectionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_restore_testing_selection" "test" {
  name = %[1]q

  restore_testing_plan_name = aws_backup_restore_testing_plan.test.name
  protected_resource_type   = "EC2"
  iam_role_arn              = aws_iam_role.test.arn

  protected_resource_conditions {
    string_equals {
      key   = "aws:ResourceTag/backup"
      value = true
    }
  }

  validation_window_hours = 10

  restore_metadata_overrides = {
    instanceType = "t2.micro"
  }
}
`, rName))
}
