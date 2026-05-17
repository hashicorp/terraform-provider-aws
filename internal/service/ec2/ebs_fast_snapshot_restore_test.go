// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSFastSnapshotRestore_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ebs_fast_snapshot_restore.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSFastSnapshotRestoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSFastSnapshotRestoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSFastSnapshotRestoreExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSnapshotID, snapshotResourceName, names.AttrID),
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

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ebs_fast_snapshot_restore.test"
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSFastSnapshotRestoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSFastSnapshotRestoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSFastSnapshotRestoreExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceEBSFastSnapshotRestore, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSFastSnapshotRestore_disappearsSnapshot(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ebs_fast_snapshot_restore.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSFastSnapshotRestoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSFastSnapshotRestoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSFastSnapshotRestoreExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceEBSSnapshot(), snapshotResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEBSFastSnapshotRestoreDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ebs_fast_snapshot_restore" {
				continue
			}

			_, err := tfec2.FindFastSnapshotRestoreByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAvailabilityZone], rs.Primary.Attributes[names.AttrSnapshotID])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 EBS Fast Snapshot Restore %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEBSFastSnapshotRestoreExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		_, err := tfec2.FindFastSnapshotRestoreByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAvailabilityZone], rs.Primary.Attributes[names.AttrSnapshotID])

		return err
	}
}

func testAccEBSFastSnapshotRestoreConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSFastSnapshotRestoreConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEBSFastSnapshotRestoreConfig_base(rName), `
resource "aws_ebs_fast_snapshot_restore" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  snapshot_id       = aws_ebs_snapshot.test.id
}
`)
}
