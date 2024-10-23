// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
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

func TestAccBackupSelection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.BackupSelection
	resourceName := "aws_backup_selection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSelectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelectionExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBackupSelection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.BackupSelection
	resourceName := "aws_backup_selection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSelectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelectionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfbackup.ResourceSelection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupSelection_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.BackupSelection
	resourceName := "aws_backup_selection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSelectionConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "selection_tag.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBackupSelection_conditionsWithTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.BackupSelection
	resourceName := "aws_backup_selection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSelectionConfig_conditionsTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "condition.0.string_equals.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "condition.0.string_like.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "condition.0.string_not_equals.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "condition.0.string_not_like.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBackupSelection_withResources(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.BackupSelection
	resourceName := "aws_backup_selection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSelectionConfig_resources(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "resources.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBackupSelection_withNotResources(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.BackupSelection
	resourceName := "aws_backup_selection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSelectionConfig_notResources(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "not_resources.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBackupSelection_updateTag(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.BackupSelection
	resourceName := "aws_backup_selection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSelectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSelectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelectionExists(ctx, resourceName, &v1),
				),
			},
			{
				Config: testAccSelectionConfig_updateTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSelectionExists(ctx, resourceName, &v2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSelectionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckSelectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_selection" {
				continue
			}

			_, err := tfbackup.FindSelectionByTwoPartKey(ctx, conn, rs.Primary.Attributes["plan_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Selection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSelectionExists(ctx context.Context, n string, v *awstypes.BackupSelection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		output, err := tfbackup.FindSelectionByTwoPartKey(ctx, conn, rs.Primary.Attributes["plan_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSelectionImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		id := fmt.Sprintf("%s|%s",
			rs.Primary.Attributes["plan_id"],
			rs.Primary.ID)

		return id, nil
	}
}

func testAccSelectionConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_plan" "test" {
  name = %[1]q

  rule {
    rule_name         = %[1]q
    target_vault_name = aws_backup_vault.test.name
    schedule          = "cron(0 12 * * ? *)"
  }
}
`, rName)
}

func testAccSelectionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSelectionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name         = %[1]q
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  resources = [
    "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/*"
  ]
}
`, rName))
}

func testAccSelectionConfig_tags(rName string) string {
	return acctest.ConfigCompose(
		testAccSelectionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name         = %[1]q
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "boo"
    value = "far"
  }

  resources = [
    "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/*"
  ]
}
`, rName))
}

func testAccSelectionConfig_conditionsTags(rName string) string {
	return acctest.ConfigCompose(
		testAccSelectionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name = %[1]q

  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  condition {
    string_equals {
      key   = "aws:ResourceTag/Component"
      value = "rds"
    }
    string_equals {
      key   = "aws:ResourceTag/Team"
      value = "dev"
    }
    string_like {
      key   = "aws:ResourceTag/Application"
      value = "app*"
    }
    string_not_equals {
      key   = "aws:ResourceTag/Backup"
      value = "false"
    }
    string_not_equals {
      key   = "aws:ResourceTag/Team"
      value = "infra"
    }
    string_not_like {
      key   = "aws:ResourceTag/Environment"
      value = "test*"
    }
  }

  resources = [
    "arn:${data.aws_partition.current.partition}:rds:*:*:cluster:*",
    "arn:${data.aws_partition.current.partition}:rds:*:*:db:*"
  ]
}
`, rName))
}

func testAccSelectionConfig_resources(rName string) string {
	return acctest.ConfigCompose(
		testAccSelectionConfig_base(rName),
		fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name         = %[1]q
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  resources = aws_ebs_volume.test[*].arn
}
`, rName))
}

func testAccSelectionConfig_notResources(rName string) string {
	return acctest.ConfigCompose(
		testAccSelectionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name         = %[1]q
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo"
    value = "bar"
  }

  not_resources = ["arn:${data.aws_partition.current.partition}:fsx:*"]
  resources     = ["*"]
}
`, rName))
}

func testAccSelectionConfig_updateTag(rName string) string {
	return acctest.ConfigCompose(
		testAccSelectionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_backup_selection" "test" {
  plan_id = aws_backup_plan.test.id

  name         = %[1]q
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

  selection_tag {
    type  = "STRINGEQUALS"
    key   = "foo2"
    value = "bar2"
  }

  resources = [
    "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:volume/*"
  ]
}
`, rName))
}
