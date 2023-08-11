// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"fmt"
	"testing"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDMSReplicationSubnetGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_subnet_group.test"
	dataSourceName := "data.aws_dms_replication_subnet_group.test"

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSubnetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSubnetGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_subnet_group_id", resourceName, "replication_subnet_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_subnet_group_description", resourceName, "replication_subnet_group_description"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccReplicationSubnetGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "dms_vpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-dms-replication-subnet-group-%[1]s"
  }
}

resource "aws_subnet" "dms_subnet_1" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.dms_vpc.id

  tags = {
    Name = "tf-acc-dms-replication-subnet-group-1-%[1]s"
  }

  depends_on = [aws_vpc.dms_vpc]
}

resource "aws_subnet" "dms_subnet_2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.dms_vpc.id

  tags = {
    Name = "tf-acc-dms-replication-subnet-group-2-%[1]s"
  }

  depends_on = [aws_vpc.dms_vpc]
}

resource "aws_subnet" "dms_subnet_3" {
  cidr_block        = "10.1.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.dms_vpc.id

  tags = {
    Name = "tf-acc-dms-replication-subnet-group-3-%[1]s"
  }

  depends_on = [aws_vpc.dms_vpc]
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = [aws_subnet.dms_subnet_1.id, aws_subnet.dms_subnet_2.id]
}

data "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}
`, rName))
}
