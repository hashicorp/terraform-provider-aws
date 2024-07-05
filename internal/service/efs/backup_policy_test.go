// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSBackupPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.BackupPolicy
	resourceName := "aws_efs_backup_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPolicyConfig_basic(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.0.status", "ENABLED"),
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

func TestAccEFSBackupPolicy_Disappears_fs(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.BackupPolicy
	resourceName := "aws_efs_backup_policy.test"
	fsResourceName := "aws_efs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPolicyConfig_basic(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupPolicyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfefs.ResourceFileSystem(), fsResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEFSBackupPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.BackupPolicy
	resourceName := "aws_efs_backup_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackupPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackupPolicyConfig_basic(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.0.status", "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBackupPolicyConfig_basic(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.0.status", "ENABLED"),
				),
			},
			{
				Config: testAccBackupPolicyConfig_basic(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackupPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "backup_policy.0.status", "DISABLED"),
				),
			},
		},
	})
}

func testAccCheckBackupPolicyExists(ctx context.Context, n string, v *awstypes.BackupPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)

		output, err := tfefs.FindBackupPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBackupPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_efs_backup_policy" {
				continue
			}

			output, err := tfefs.FindBackupPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if output.Status == awstypes.StatusDisabled {
				continue
			}

			return fmt.Errorf("EFS Backup Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBackupPolicyConfig_basic(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_backup_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  backup_policy {
    status = %[2]q
  }
}
`, rName, status)
}
