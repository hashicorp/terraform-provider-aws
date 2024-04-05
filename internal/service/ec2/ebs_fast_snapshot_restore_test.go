// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSFastSnapshotRestore_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_fast_snapshot_restore.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSFastSnapshotRestoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSFastSnapshotRestoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSFastSnapshotRestoreExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "snapshot_id", snapshotResourceName, "id"),
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

func TestAccEC2EBSFastSnapshotRestore_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_fast_snapshot_restore.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSFastSnapshotRestoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSFastSnapshotRestoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSFastSnapshotRestoreExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceEBSFastSnapshotRestore, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSFastSnapshotRestore_disappearsSnapshot(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_fast_snapshot_restore.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSFastSnapshotRestoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSFastSnapshotRestoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSFastSnapshotRestoreExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceEBSSnapshot(), snapshotResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEBSFastSnapshotRestoreDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ebs_fast_snapshot_restore" {
				continue
			}

			out, err := tfec2.FindEBSFastSnapshotRestoreByID(ctx, conn, rs.Primary.ID)
			if errors.Is(err, tfresource.ErrEmptyResult) || out.State == types.FastSnapshotRestoreStateCodeDisabled {
				return nil
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameEBSFastSnapshotRestore, rs.Primary.ID, err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameEBSFastSnapshotRestore, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckEBSFastSnapshotRestoreExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameEBSFastSnapshotRestore, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameEBSFastSnapshotRestore, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindEBSFastSnapshotRestoreByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameEBSFastSnapshotRestore, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccEBSFastSnapshotRestoreBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id
}
`, rName))
}

func testAccEBSFastSnapshotRestoreConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccEBSFastSnapshotRestoreBaseConfig(rName),
		`
resource "aws_ebs_fast_snapshot_restore" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  snapshot_id       = aws_ebs_snapshot.test.id
}
`)
}
