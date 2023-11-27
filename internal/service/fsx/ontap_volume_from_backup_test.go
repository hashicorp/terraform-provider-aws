// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"

	"github.com/YakDriver/regexache"

	"github.com/aws/aws-sdk-go-v2/service/fsx"
	"github.com/aws/aws-sdk-go-v2/service/fsx/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
)

func TestAccFSxOntapVolumeFromBackup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ontapvolumefrombackup awstypes.Volume
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fsx_ontap_volume_from_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeFromBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOntapVolumeFromBackupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeFromBackupExists(ctx, resourceName, &ontapvolumefrombackup),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexache.MustCompile(`ontapvolumefrombackup:+.`)),
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

func TestAccFSxOntapVolumeFromBackup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ontapvolumefrombackup awstypes.Volume
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fsx_ontap_volume_from_backup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, fsx.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, fsx.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOntapVolumeFromBackupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOntapVolumeFromBackupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOntapVolumeFromBackupExists(ctx, resourceName, &ontapvolumefrombackup),

					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tffsx.ResourceOntapVolumeFromBackup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOntapVolumeFromBackupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_ontap_volume_from_backup" {
				continue
			}

			_, err := tffsx.FindOntapVolumeByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *types.VolumeNotFound
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.FSx, create.ErrActionCheckingDestroyed, tffsx.ResNameOntapVolumeFromBackup, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}

}

func testAccCheckOntapVolumeFromBackupExists(ctx context.Context, name string, ontapvolumefrombackup *awstypes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FSx, create.ErrActionCheckingExistence, tffsx.ResNameOntapVolumeFromBackup, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FSx, create.ErrActionCheckingExistence, tffsx.ResNameOntapVolumeFromBackup, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)
		resp, err := tffsx.FindOntapVolumeByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.FSx, create.ErrActionCheckingExistence, tffsx.ResNameOntapVolumeFromBackup, rs.Primary.ID, err)
		}

		ontapvolumefrombackup = resp

		return nil
	}
}

/*func testAccCheckOntapVolumeFromBackupNotRecreated(before, after *fsx.DescribeOntapVolumeFromBackupResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.OntapVolumeFromBackupId), aws.ToString(after.OntapVolumeFromBackupId); before != after {
			return create.Error(names.FSx, create.ErrActionCheckingNotRecreated, tffsx.ResNameOntapVolumeFromBackup, aws.ToString(before.OntapVolumeFromBackupId), errors.New("recreated"))
		}

		return nil
	}
}*/

func testAccOntapVolumeFromBackupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_fsx_ontap_volume_from_backup" "test" {
  	name             = %[1]q
	backup_id = "backup-0b655441cd2dc54fa"
}
`, rName)
}
