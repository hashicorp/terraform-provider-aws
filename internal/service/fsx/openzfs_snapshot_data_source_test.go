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
)

func TestAccFSxOpenzfsSnapshotDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_fsx_openzfs_snapshot.test"
	resourceName := "aws_fsx_openzfs_snapshot.test"
	mostRecentResourceName := "aws_fsx_openzfs_snapshot.latest"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, fsx.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenzfsSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenzfsSnapshotDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_time", resourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", resourceName, "volume_id"),
				),
			},
			{
				Config: testAccOpenzfsSnapshotDataSourceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccOpenzfsSnapshotDataSourceConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccOpenzfsSnapshotDataSourceConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccOpenzfsSnapshotDataSourceConfig_filterFileSystemId(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_time", resourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", resourceName, "volume_id"),
				),
			},
			{
				Config: testAccOpenzfsSnapshotDataSourceConfig_filterVolumeId(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_time", resourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", resourceName, "volume_id"),
				),
			},
			{
				Config: testAccOpenzfsSnapshotDataSourceConfig_mostRecent(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", mostRecentResourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_time", mostRecentResourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", mostRecentResourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", mostRecentResourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", mostRecentResourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", mostRecentResourceName, "volume_id"),
				),
			},
		},
	})
}

func testAccOpenzfsSnapshotDataSourceBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOpenzfsSnapshotDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotDataSourceBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

data "aws_fsx_openzfs_snapshot" "test" {
  snapshot_ids = [aws_fsx_openzfs_snapshot.test.id]
}
`, rName))
}

func testAccOpenzfsSnapshotDataSourceConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotDataSourceBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id

  tags = {
    %[2]q = %[3]q
  }
}

data "aws_fsx_openzfs_snapshot" "test" {
  snapshot_ids = [aws_fsx_openzfs_snapshot.test.id]
}
`, rName, tagKey1, tagValue1))
}

func testAccOpenzfsSnapshotDataSourceConfig_tags2(rName string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotDataSourceBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

data "aws_fsx_openzfs_snapshot" "test" {
  snapshot_ids = [aws_fsx_openzfs_snapshot.test.id]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccOpenzfsSnapshotDataSourceConfig_filterFileSystemId(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotDataSourceBaseConfig(rName), fmt.Sprintf(`
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

func testAccOpenzfsSnapshotDataSourceConfig_filterVolumeId(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotDataSourceBaseConfig(rName), fmt.Sprintf(`
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

func testAccOpenzfsSnapshotDataSourceConfig_mostRecent(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotDataSourceBaseConfig(rName), fmt.Sprintf(`
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
