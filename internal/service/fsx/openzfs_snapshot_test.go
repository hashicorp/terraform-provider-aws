// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxOpenZFSSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`snapshot/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "volume_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccFSxOpenZFSSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceOpenZFSSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxOpenZFSSnapshot_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot),
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
				Config: testAccOpenZFSSnapshotConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccOpenZFSSnapshotConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSSnapshot_name(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot1, snapshot2 awstypes.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSSnapshotConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot2),
					testAccCheckOpenZFSSnapshotNotRecreated(&snapshot1, &snapshot2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccFSxOpenZFSSnapshot_childVolume(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_childVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`snapshot/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccFSxOpenZFSSnapshot_volumeID(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot1, snapshot2 awstypes.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_volumeID1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSSnapshotConfig_volumeID2(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenZFSSnapshotExists(ctx, resourceName, &snapshot2),
					testAccCheckOpenZFSSnapshotRecreated(&snapshot1, &snapshot2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func testAccCheckOpenZFSSnapshotExists(ctx context.Context, n string, v *awstypes.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		output, err := tffsx.FindSnapshotByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckOpenZFSSnapshotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_openzfs_snapshot" {
				continue
			}

			_, err := tffsx.FindSnapshotByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx OpenZFS Snapshot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOpenZFSSnapshotNotRecreated(i, j *awstypes.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.SnapshotId) != aws.ToString(j.SnapshotId) {
			return fmt.Errorf("FSx OpenZFS Snapshot (%s) recreated", aws.ToString(i.SnapshotId))
		}

		return nil
	}
}

func testAccCheckOpenZFSSnapshotRecreated(i, j *awstypes.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.SnapshotId) == aws.ToString(j.SnapshotId) {
			return fmt.Errorf("FSx OpenZFS Snapshot (%s) not recreated", aws.ToString(i.SnapshotId))
		}

		return nil
	}
}

func testAccOpenZFSSnapshotConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test[0].id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOpenZFSSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}
`, rName))
}

func testAccOpenZFSSnapshotConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccOpenZFSSnapshotConfig_tags2(rName string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id


  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccOpenZFSSnapshotConfig_childVolume(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_volume.test.id
}
`, rName))
}

func testAccOpenZFSSnapshotConfig_volumeID1(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test1" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_volume.test1.id
}
`, rName))
}

func testAccOpenZFSSnapshotConfig_volumeID2(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test2" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_volume.test2.id
}
`, rName))
}
