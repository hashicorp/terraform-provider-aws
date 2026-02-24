// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBackupLogicallyAirGappedVault_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogicallyAirGappedVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogicallyAirGappedVaultConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("max_retention_days"), knownvalue.Int64Exact(10)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("min_retention_days"), knownvalue.Int64Exact(7)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
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

func TestAccBackupLogicallyAirGappedVault_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogicallyAirGappedVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogicallyAirGappedVaultConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbackup.ResourceLogicallyAirGappedVault, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupLogicallyAirGappedVault_encryptionKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogicallyAirGappedVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogicallyAirGappedVaultConfig_encryptionKeyARN(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("encryption_key_arn"), kmsKeyResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
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

func TestAccBackupLogicallyAirGappedVault_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogicallyAirGappedVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogicallyAirGappedVaultConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLogicallyAirGappedVaultConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
				),
			},
			{
				Config: testAccLogicallyAirGappedVaultConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
				),
			},
		},
	})
}

func testAccCheckLogicallyAirGappedVaultDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BackupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_logically_air_gapped_vault" {
				continue
			}

			_, err := tfbackup.FindLogicallyAirGappedBackupVaultByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Logically Air Gapped Vault %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLogicallyAirGappedVaultExists(ctx context.Context, t *testing.T, n string, v *backup.DescribeBackupVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BackupClient(ctx)

		output, err := tfbackup.FindLogicallyAirGappedBackupVaultByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLogicallyAirGappedVaultConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  max_retention_days = 10
  min_retention_days = 7
}
`, rName)
}

func testAccLogicallyAirGappedVaultConfig_encryptionKeyARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  max_retention_days = 10
  min_retention_days = 7
  encryption_key_arn = aws_kms_key.test.arn
}
`, rName)
}

func testAccLogicallyAirGappedVaultConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

func testAccLogicallyAirGappedVaultConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
