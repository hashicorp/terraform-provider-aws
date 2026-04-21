// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupVaultPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(ctx, t, resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile("^{\"Id\":\"default\".+"))),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVaultPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(ctx, t, resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile("^{\"Id\":\"default\".+")),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile("backup:ListRecoveryPointsByBackupVault")),
				),
			},
		},
	})
}

func TestAccBackupVaultPolicy_eventualConsistency(t *testing.T) {
	ctx := acctest.Context(t)
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultPolicyConfig_eventualConsistency(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(ctx, t, resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile("^{\"Id\":\"default\".+"))),
			},
		},
	})
}

func TestAccBackupVaultPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(ctx, t, resourceName, &vault),
					acctest.CheckSDKResourceDisappears(ctx, t, tfbackup.ResourceVaultPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupVaultPolicy_Disappears_vault(t *testing.T) {
	ctx := acctest.Context(t)
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"
	vaultResourceName := "aws_backup_vault.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(ctx, t, resourceName, &vault),
					acctest.CheckSDKResourceDisappears(ctx, t, tfbackup.ResourceVault(), vaultResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupVaultPolicy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(ctx, t, resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile("\"Version\":\"2012-10-17\"")),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccVaultPolicyConfig_newOrder(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckVaultPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BackupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_vault_policy" {
				continue
			}

			_, err := tfbackup.FindVaultAccessPolicyByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Vault Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVaultPolicyExists(ctx context.Context, t *testing.T, n string, v *backup.GetBackupVaultAccessPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BackupClient(ctx)

		output, err := tfbackup.FindVaultAccessPolicyByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVaultPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "default"
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "backup:DescribeBackupVault",
        "backup:DeleteBackupVault",
        "backup:PutBackupVaultAccessPolicy",
        "backup:DeleteBackupVaultAccessPolicy",
        "backup:GetBackupVaultAccessPolicy",
        "backup:StartBackupJob",
        "backup:GetBackupVaultNotifications",
        "backup:PutBackupVaultNotifications",
      ]
      Resource = aws_backup_vault.test.arn
    }]
  })
}
`, rName)
}

func testAccVaultPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "default"
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "backup:DescribeBackupVault",
        "backup:DeleteBackupVault",
        "backup:PutBackupVaultAccessPolicy",
        "backup:DeleteBackupVaultAccessPolicy",
        "backup:GetBackupVaultAccessPolicy",
        "backup:StartBackupJob",
        "backup:GetBackupVaultNotifications",
        "backup:PutBackupVaultNotifications",
        "backup:ListRecoveryPointsByBackupVault",
      ]
      Resource = aws_backup_vault.test.arn
    }]
  })
}
`, rName)
}

func testAccVaultPolicyConfig_newOrder(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "default"
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "backup:DeleteBackupVault",
        "backup:PutBackupVaultNotifications",
        "backup:GetBackupVaultNotifications",
        "backup:GetBackupVaultAccessPolicy",
        "backup:DeleteBackupVaultAccessPolicy",
        "backup:DescribeBackupVault",
        "backup:StartBackupJob",
        "backup:PutBackupVaultAccessPolicy",
      ]
      Resource = aws_backup_vault.test.arn
    }]
  })
}
`, rName)
}

func testAccVaultPolicyConfig_eventualConsistency(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
data "aws_partition" "current" {}

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
          Service = "backup.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_partition.current.partition}:policy/service-role/AWSBackupServiceRolePolicyForBackup"
}

resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "default"
      Effect = "Allow"
      Principal = {
        AWS = aws_iam_role.test.arn
      }
      Action = [
        "backup:DescribeBackupVault",
        "backup:DeleteBackupVault",
        "backup:PutBackupVaultAccessPolicy",
        "backup:DeleteBackupVaultAccessPolicy",
        "backup:GetBackupVaultAccessPolicy",
        "backup:StartBackupJob",
        "backup:GetBackupVaultNotifications",
        "backup:PutBackupVaultNotifications",
      ]
      Resource = aws_backup_vault.test.arn
    }]
  })
}
`, rName))
}
