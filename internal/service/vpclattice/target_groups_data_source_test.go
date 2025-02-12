// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeTargetGroupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupsDataSourceConfig_dataSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue("data.aws_vpclattice_target_groups.all", "ids.#", 3),
					resource.TestCheckResourceAttr("data.aws_vpclattice_target_groups.name_prefix", "ids.#", "3"),
					resource.TestCheckResourceAttr("data.aws_vpclattice_target_groups.tags", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_vpclattice_target_groups.type", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_vpclattice_target_groups.vpc_id", "ids.#", "2"),
				),
			},
		},
	})
}

func testAccTargetGroupsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test_1" {
  cidr_block = "172.16.0.0/16"
}

resource "aws_vpc" "test_2" {
  cidr_block = "172.16.0.0/16"
}

resource "aws_vpclattice_target_group" "test_1" {
  name = format("%%s-1", %[1]q)
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test_1.id
  }

  tags = {
    Name = format("%%s-1", %[1]q)
  }
}

resource "aws_vpclattice_target_group" "test_2" {
  name = format("%%s-2", %[1]q)
  type = "IP"
  
  config {
    port           = 443
    protocol       = "HTTPS"
    vpc_identifier = aws_vpc.test_1.id
  }

  tags = {
    Name = format("%%s-2", %[1]q),
  }
}

resource "aws_vpclattice_target_group" "test_3" {
  name = format("%%s-3", %[1]q)
  type = "IP"
  
  config {
    port           = 443
    protocol       = "HTTPS"
    vpc_identifier = aws_vpc.test_2.id
  }

  tags = {
    Name = format("%%s-3", %[1]q)
  }
}
`, rName)
}

func testAccTargetGroupsDataSourceConfig_dataSource(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupsDataSourceConfig_basic(rName), fmt.Sprintf(`
data "aws_vpclattice_target_groups" "all" {
  depends_on = [
    aws_vpclattice_target_group.test_1,
    aws_vpclattice_target_group.test_2,
    aws_vpclattice_target_group.test_3,
  ]
}

data "aws_vpclattice_target_groups" "name_prefix" {
  name_prefix = %[1]q

  depends_on = [
    aws_vpclattice_target_group.test_1,
    aws_vpclattice_target_group.test_2,
    aws_vpclattice_target_group.test_3,
  ]
}

data "aws_vpclattice_target_groups" "type" {
  name_prefix = %[1]q
  type        = "IP"

  depends_on = [
    aws_vpclattice_target_group.test_1,
    aws_vpclattice_target_group.test_2,
    aws_vpclattice_target_group.test_3,
  ]
}

data "aws_vpclattice_target_groups" "vpc_id" {
  vpc_id = aws_vpc.test_1.id

  depends_on = [
    aws_vpclattice_target_group.test_1,
    aws_vpclattice_target_group.test_2,
    aws_vpclattice_target_group.test_3,
  ]
}

data "aws_vpclattice_target_groups" "tags" {
  tags = {
    Name = format("%%s-3", %[1]q)
  }

  depends_on = [
    aws_vpclattice_target_group.test_1,
    aws_vpclattice_target_group.test_2,
    aws_vpclattice_target_group.test_3,
  ]
}
`, rName))
}
