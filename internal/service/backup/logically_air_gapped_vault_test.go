// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func TestAccBackupLAGVault_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var logicallyairgappedvault backup.DescribeBackupVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLAGVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLAGVaultConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLAGVaultExists(ctx, resourceName, &logicallyairgappedvault),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccBackupLAGVault_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var logicallyairgappedvault backup.DescribeBackupVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLAGVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLAGVaultConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLAGVaultExists(ctx, resourceName, &logicallyairgappedvault),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbackup.ResourceLogicallyAirGappedVault, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupLAGVault_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var logicallyairgappedvault backup.DescribeBackupVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLAGVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLAGVaultConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLAGVaultExists(ctx, resourceName, &logicallyairgappedvault),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLAGVaultConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLAGVaultExists(ctx, resourceName, &logicallyairgappedvault),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLAGVaultConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLAGVaultExists(ctx, resourceName, &logicallyairgappedvault),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckLAGVaultDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_logically_air_gapped_vault" {
				continue
			}

			_, err := conn.DescribeBackupVault(ctx, &backup.DescribeBackupVaultInput{
				BackupVaultName: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) || errs.MessageContains(err, "AccessDeniedException", "Insufficient privileges to perform this action") {
				return nil
			}
			if err != nil {
				return create.Error(names.Backup, create.ErrActionCheckingDestroyed, tfbackup.ResNameLogicallyAirGappedVault, rs.Primary.ID, err)
			}

			return create.Error(names.Backup, create.ErrActionCheckingDestroyed, tfbackup.ResNameLogicallyAirGappedVault, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckLAGVaultExists(ctx context.Context, name string, logicallyairgappedvault *backup.DescribeBackupVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameLogicallyAirGappedVault, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameLogicallyAirGappedVault, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupClient(ctx)
		resp, err := conn.DescribeBackupVault(ctx, &backup.DescribeBackupVaultInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Backup, create.ErrActionCheckingExistence, tfbackup.ResNameLogicallyAirGappedVault, rs.Primary.ID, err)
		}

		*logicallyairgappedvault = *resp

		return nil
	}
}

func testAccCheckLAGVaultNotRecreated(before, after *backup.DescribeBackupVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.BackupVaultName), aws.ToString(after.BackupVaultName); before != after {
			return create.Error(names.Backup, create.ErrActionCheckingNotRecreated, tfbackup.ResNameLogicallyAirGappedVault, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccLAGVaultConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  max_retention_days = 7
  min_retention_days = 7
}
`, rName)
}

func testAccLAGVaultConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  max_retention_days = 7
  min_retention_days = 7

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccLAGVaultConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  max_retention_days = 7
  min_retention_days = 7

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
