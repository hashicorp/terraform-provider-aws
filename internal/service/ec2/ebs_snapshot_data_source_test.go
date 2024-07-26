// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSSnapshotDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ebs_snapshot.test"
	resourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEncrypted, resourceName, names.AttrEncrypted),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_alias", resourceName, "owner_alias"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_tier", resourceName, "storage_tier"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_id", resourceName, "volume_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVolumeSize, resourceName, names.AttrVolumeSize),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ebs_snapshot.test"
	resourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotDataSource_mostRecent(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ebs_snapshot.test"
	resourceName := "aws_ebs_snapshot.b"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotDataSourceConfig_mostRecent(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccEBSSnapshotDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
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

data "aws_ebs_snapshot" "test" {
  snapshot_ids = [aws_ebs_snapshot.test.id]
}
`, rName))
}

func testAccEBSSnapshotDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
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

data "aws_ebs_snapshot" "test" {
  filter {
    name   = "snapshot-id"
    values = [aws_ebs_snapshot.test.id]
  }
}
`, rName))
}

func testAccEBSSnapshotDataSourceConfig_mostRecent(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "a" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "b" {
  # Ensure that this snapshot is created after the other.
  volume_id = aws_ebs_snapshot.a.volume_id

  tags = {
    Name = %[1]q
  }
}

data "aws_ebs_snapshot" "test" {
  most_recent = true

  filter {
    name   = "tag:Name"
    values = [%[1]q]
  }

  depends_on = [aws_ebs_snapshot.a, aws_ebs_snapshot.b]
}
`, rName))
}
