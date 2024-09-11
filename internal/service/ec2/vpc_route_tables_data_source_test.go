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

func TestAccVPCRouteTablesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTablesDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_route_tables.by_vpc_id", "ids.#", acctest.Ct2), // Add the default route table.
					resource.TestCheckResourceAttr("data.aws_route_tables.by_tags", "ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr("data.aws_route_tables.by_filter", "ids.#", "6"), // Add the default route tables.
					resource.TestCheckResourceAttr("data.aws_route_tables.empty", "ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccVPCRouteTablesDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test1_public" {
  vpc_id = aws_vpc.test1.id

  tags = {
    Name      = %[1]q
    Tier      = "Public"
    Component = "Frontend"
  }
}

resource "aws_route_table" "test1_private1" {
  vpc_id = aws_vpc.test1.id

  tags = {
    Name      = %[1]q
    Tier      = "Private"
    Component = "Database"
  }
}

resource "aws_route_table" "test1_private2" {
  vpc_id = aws_vpc.test1.id

  tags = {
    Name      = %[1]q
    Tier      = "Private"
    Component = "AppServer"
  }
}

resource "aws_route_table" "test2_public" {
  vpc_id = aws_vpc.test2.id

  tags = {
    Name      = %[1]q
    Tier      = "Public"
    Component = "Frontend"
  }
}

data "aws_route_tables" "by_vpc_id" {
  vpc_id = aws_vpc.test2.id

  depends_on = [aws_route_table.test1_public, aws_route_table.test1_private1, aws_route_table.test1_private2, aws_route_table.test2_public]
}

data "aws_route_tables" "by_tags" {
  tags = {
    Tier = "Public"
  }

  depends_on = [aws_route_table.test1_public, aws_route_table.test1_private1, aws_route_table.test1_private2, aws_route_table.test2_public]
}

data "aws_route_tables" "by_filter" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test1.id, aws_vpc.test2.id]
  }

  depends_on = [aws_route_table.test1_public, aws_route_table.test1_private1, aws_route_table.test1_private2, aws_route_table.test2_public]
}

data "aws_route_tables" "empty" {
  vpc_id = aws_vpc.test2.id

  tags = {
    Tier = "Private"
  }

  depends_on = [aws_route_table.test1_public, aws_route_table.test1_private1, aws_route_table.test1_private2, aws_route_table.test2_public]
}
`, rName)
}
