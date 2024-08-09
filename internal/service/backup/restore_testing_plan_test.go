// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRestoreTestingPlan_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var restoreTestingPlan backup.GetRestoreTestingPlanOutput
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "backup", regexache.MustCompile(`restore-testing-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 12 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "RANDOM_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recovery_point_selection.0.include_vaults.*", "*"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recovery_point_selection.0.recovery_point_types.*", "CONTINUOUS"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recovery_point_selection.0.recovery_point_types.*", "SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccRestoreTestingPlan_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var restoreTestingPlan backup.GetRestoreTestingPlanOutput
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfbackup.ResourceRestoreTestingPlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRestoreTestingPlan_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var restoreTestingPlan backup.GetRestoreTestingPlanOutput
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRestoreTestingPlanConfig_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
			{
				Config: testAccRestoreTestingPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccRestoreTestingPlan_updateRecoveryPointSelection(t *testing.T) {
	ctx := acctest.Context(t)
	var restoreTestingPlan backup.GetRestoreTestingPlanOutput
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "backup", regexache.MustCompile(`restore-testing-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 12 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "RANDOM_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recovery_point_selection.0.include_vaults.*", "*"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recovery_point_selection.0.recovery_point_types.*", "CONTINUOUS"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recovery_point_selection.0.recovery_point_types.*", "SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRestoreTestingPlanConfig_recoveryPointSelectionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "backup", regexache.MustCompile(`restore-testing-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 12 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "LATEST_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recovery_point_selection.0.recovery_point_types.*", "SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccRestoreTestingPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 12 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "RANDOM_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccRestoreTestingPlan_withOptionalArguments(t *testing.T) {
	ctx := acctest.Context(t)
	var restoreTestingPlan backup.GetRestoreTestingPlanOutput
	resourceName := "aws_backup_restore_testing_plan.test"
	rName := fmt.Sprintf("tf_testacc_backup_%s", sdkacctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestoreTestingPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestoreTestingPlanConfig_optionalArguments(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "backup", regexache.MustCompile(`restore-testing-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 12 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", "EST5EDT"),
					resource.TestCheckResourceAttr(resourceName, "start_window", "168"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "RANDOM_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recovery_point_selection.0.recovery_point_types.*", "CONTINUOUS"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recovery_point_selection.0.recovery_point_types.*", "SNAPSHOT"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.selection_window", "365"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRestoreTestingPlanConfig_optionalArgumentsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "backup", regexache.MustCompile(`restore-testing-plan:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 12 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", "PST8PDT"),
					resource.TestCheckResourceAttr(resourceName, "start_window", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "RANDOM_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.exclude_vaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.selection_window", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccRestoreTestingPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestoreTestingPlanExists(ctx, resourceName, &restoreTestingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 12 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.algorithm", "RANDOM_WITHIN_WINDOW"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.include_vaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recovery_point_selection.0.recovery_point_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckRestoreTestingPlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_plan" {
				continue
			}

			_, err := tfbackup.FindRestoreTestingPlanByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Restore Testing Plan %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRestoreTestingPlanExists(ctx context.Context, n string, v *backup.GetRestoreTestingPlanOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		output, err := tfbackup.FindRestoreTestingPlanByID(ctx, conn, rs.Primary.ID)

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
  schedule = "cron(0 12 * * ? *)"
  recovery_point_selection {
	algorithm = "RANDOM_WITHIN_WINDOW"
	include_vaults = ["*"]
	recovery_point_types = ["CONTINUOUS", "SNAPSHOT"]
  }
}
`, rName)
}

func testAccRestoreTestingPlanConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q
  schedule = "cron(0 12 * * ? *)"
  recovery_point_selection {
	algorithm = "RANDOM_WITHIN_WINDOW"
	include_vaults = ["*"]
	recovery_point_types = ["CONTINUOUS", "SNAPSHOT"]
  }

  tags = {
    Name = %[1]q
	Key1 = "Value1"
	Key2 = "Value2a"
  }
}
`, rName)
}

func testAccRestoreTestingPlanConfig_tagsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q
  schedule = "cron(0 12 * * ? *)"
  recovery_point_selection {
	algorithm = "RANDOM_WITHIN_WINDOW"
	include_vaults = ["*"]
	recovery_point_types = ["CONTINUOUS", "SNAPSHOT"]
  }

  tags = {
    Name = %[1]q
	Key2 = "Value2b"
	Key3 = "Value3"
  }
}
`, rName)
}

func testAccRestoreTestingPlanConfig_recoveryPointSelectionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q
  schedule = "cron(0 12 * * ? *)"
  recovery_point_selection {
	algorithm = "LATEST_WITHIN_WINDOW"
	include_vaults = ["*"]
	recovery_point_types = ["SNAPSHOT"]
  }
}
`, rName)
}

func testAccRestoreTestingConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test_one" {
  name = "%[1]s_1"
}

resource "aws_backup_vault" "test_two" {
  name = "%[1]s_2"
}
`, rName)
}

func testAccRestoreTestingPlanConfig_optionalArguments(rName string) string {
	return acctest.ConfigCompose(
		testAccRestoreTestingConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q
  schedule = "cron(0 12 * * ? *)"
  schedule_timezone = "EST5EDT"
  start_window = 168
  recovery_point_selection {
	algorithm = "RANDOM_WITHIN_WINDOW"
	include_vaults = [aws_backup_vault.test_one.arn]
	exclude_vaults = [aws_backup_vault.test_two.arn]
	recovery_point_types = ["CONTINUOUS", "SNAPSHOT"]
	selection_window = 365
  }
}
`, rName))
}

func testAccRestoreTestingPlanConfig_optionalArgumentsUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccRestoreTestingConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_restore_testing_plan" "test" {
  name = %[1]q
  schedule = "cron(0 12 * * ? *)"
  schedule_timezone = "PST8PDT"
  start_window = 1
  recovery_point_selection {
	algorithm = "RANDOM_WITHIN_WINDOW"
	include_vaults = [aws_backup_vault.test_one.arn, aws_backup_vault.test_two.arn]
	exclude_vaults = []
	recovery_point_types = ["CONTINUOUS", "SNAPSHOT"]
	selection_window = 1
  }
}
`, rName))
}
