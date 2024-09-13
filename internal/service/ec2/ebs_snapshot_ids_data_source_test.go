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

func TestAccEC2EBSSnapshotIDsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ebs_snapshot_ids.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotIdsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", 0),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ids.*", "aws_ebs_snapshot.test", names.AttrID),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotIDsDataSource_sorted(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ebs_snapshot_ids.test"
	resource1Name := "aws_ebs_snapshot.a"
	resource2Name := "aws_ebs_snapshot.b"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotIdsDataSourceConfig_sorted(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", resource2Name, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.1", resource1Name, names.AttrID),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotIDsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotIdsDataSourceConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ebs_snapshot_ids.empty", "ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccEBSSnapshotIdsDataSourceConfig_basic(rName string) string {
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

data "aws_ebs_snapshot_ids" "test" {
  owners = ["self"]

  depends_on = [aws_ebs_snapshot.test]
}
`, rName))
}

func testAccEBSSnapshotIdsDataSourceConfig_sorted(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  count = 2

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "a" {
  volume_id   = aws_ebs_volume.test[0].id
  description = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "b" {
  volume_id   = aws_ebs_volume.test[1].id
  description = %[1]q

  # We want to ensure that 'aws_ebs_snapshot.a.creation_date' is less than
  # 'aws_ebs_snapshot.b.creation_date'/ so that we can ensure that the
  # snapshots are being sorted correctly.
  depends_on = [aws_ebs_snapshot.a]

  tags = {
    Name = %[1]q
  }
}

data "aws_ebs_snapshot_ids" "test" {
  owners = ["self"]

  filter {
    name   = "description"
    values = [%[1]q]
  }

  depends_on = [aws_ebs_snapshot.a, aws_ebs_snapshot.b]
}
`, rName))
}

const testAccEBSSnapshotIdsDataSourceConfig_empty = `
data "aws_ebs_snapshot_ids" "empty" {
  owners = ["000000000000"]
}
`
