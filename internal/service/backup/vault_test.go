// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccBackupVault_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, backup.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "backup", fmt.Sprintf("backup-vault:%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_points", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccBackupVault_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, backup.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfbackup.ResourceVault(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupVault_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, backup.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccVaultConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVaultConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccBackupVault_withKMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, backup.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccBackupVault_forceDestroyEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, backup.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_forceDestroyEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "recovery_points", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccBackupVault_forceDestroyWithRecoveryPoint(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, backup.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_forceDestroyWithDynamoDBTable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "recovery_points", "0"),
				),
			},
			{
				Config: testAccVaultConfig_forceDestroyWithDynamoDBTable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					testAccCheckRunDynamoDBTableBackupJob(ctx, rName),
				),
			},
			{
				Config: testAccVaultConfig_forceDestroyWithDynamoDBTable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "recovery_points", "1"),
				),
			},
		},
	})
}

func testAccCheckVaultDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_vault" {
				continue
			}

			_, err := tfbackup.FindVaultByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Vault %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVaultExists(ctx context.Context, name string, v *backup.DescribeBackupVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Backup Vault ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn(ctx)

		output, err := tfbackup.FindVaultByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRunDynamoDBTableBackupJob(ctx context.Context, rName string) resource.TestCheckFunc { // nosemgrep:ci.backup-in-func-name
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(*conns.AWSClient)
		conn := client.BackupConn(ctx)

		iamRoleARN := arn.ARN{
			Partition: client.Partition,
			Service:   "iam",
			AccountID: client.AccountID,
			Resource:  "role/service-role/AWSBackupDefaultServiceRole",
		}.String()
		resourceARN := arn.ARN{
			Partition: client.Partition,
			Service:   "dynamodb",
			Region:    client.Region,
			AccountID: client.AccountID,
			Resource:  fmt.Sprintf("table/%s", rName),
		}.String()
		output, err := conn.StartBackupJobWithContext(ctx, &backup.StartBackupJobInput{
			BackupVaultName: aws.String(rName),
			IamRoleArn:      aws.String(iamRoleARN),
			ResourceArn:     aws.String(resourceARN),
		})

		if err != nil {
			return fmt.Errorf("error starting Backup Job: %w", err)
		}

		jobID := aws.StringValue(output.BackupJobId)

		_, err = tfbackup.WaitJobCompleted(ctx, conn, jobID, 10*time.Minute)

		if err != nil {
			return fmt.Errorf("error waiting for Backup Job (%s) complete: %w", jobID, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn(ctx)

	input := &backup.ListBackupVaultsInput{}

	_, err := conn.ListBackupVaultsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVaultConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}
`, rName)
}

func testAccVaultConfig_kmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 10
}

resource "aws_backup_vault" "test" {
  name        = %[1]q
  kms_key_arn = aws_kms_key.test.arn
}
`, rName)
}

func testAccVaultConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVaultConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVaultConfig_forceDestroyEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q

  force_destroy = true
}
`, rName)
}

func testAccVaultConfig_forceDestroyWithDynamoDBTable(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q

  force_destroy = true
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, rName)
}
