// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxOpenZFSSnapshotDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_fsx_openzfs_snapshot.test"
	resourceName := "aws_fsx_openzfs_snapshot.test"
	mostRecentResourceName := "aws_fsx_openzfs_snapshot.latest"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenZFSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_time", resourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", resourceName, "volume_id"),
				),
			},
			{
				Config: testAccOpenZFSSnapshotDataSourceConfig_filterFileSystemId(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_time", resourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", resourceName, "volume_id"),
				),
			},
			{
				Config: testAccOpenZFSSnapshotDataSourceConfig_filterVolumeId(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_time", resourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", resourceName, "volume_id"),
				),
			},
			{
				Config: testAccOpenZFSSnapshotDataSourceConfig_mostRecent(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, mostRecentResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_time", mostRecentResourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, mostRecentResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, mostRecentResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", mostRecentResourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", mostRecentResourceName, "volume_id"),
				),
			},
		},
	})
}

func testAccOpenZFSSnapshotDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

data "aws_fsx_openzfs_snapshot" "test" {
  snapshot_ids = [aws_fsx_openzfs_snapshot.test.id]
}
`, rName))
}

func testAccOpenZFSSnapshotDataSourceConfig_filterFileSystemId(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

data "aws_fsx_openzfs_snapshot" "test" {
  filter {
    name   = "file-system-id"
    values = [aws_fsx_openzfs_file_system.test.id]
  }
}
`, rName))
}

func testAccOpenZFSSnapshotDataSourceConfig_filterVolumeId(rName string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

data "aws_fsx_openzfs_snapshot" "test" {
  filter {
    name   = "volume-id"
    values = [aws_fsx_openzfs_file_system.test.root_volume_id]
  }
}
`, rName))
}

func testAccOpenZFSSnapshotDataSourceConfig_mostRecent(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccOpenZFSSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

resource "aws_fsx_openzfs_snapshot" "latest" {
  # Ensure that this snapshot is created after the other.
  name      = %[2]q
  volume_id = aws_fsx_openzfs_snapshot.test.volume_id
}

data "aws_fsx_openzfs_snapshot" "test" {
  most_recent = true
  filter {
    name   = "volume-id"
    values = [aws_fsx_openzfs_file_system.test.root_volume_id]
  }
  depends_on = [aws_fsx_openzfs_snapshot.test, aws_fsx_openzfs_snapshot.latest]
}
`, rName, rName2))
}
