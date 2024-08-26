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

func TestAccVPCPeeringConnectionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_peering_connections.test_by_filters", "ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionsDataSource_NoMatches(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionsDataSourceConfig_noMatches(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_peering_connections.test", "ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccVPCPeeringConnectionsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test3" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test1" {
  vpc_id      = aws_vpc.test1.id
  peer_vpc_id = aws_vpc.test2.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test2" {
  vpc_id      = aws_vpc.test1.id
  peer_vpc_id = aws_vpc.test3.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_peering_connections" "test_by_filters" {
  filter {
    name   = "vpc-peering-connection-id"
    values = [aws_vpc_peering_connection.test1.id, aws_vpc_peering_connection.test2.id]
  }
}
`, rName)
}

func testAccVPCPeeringConnectionsDataSourceConfig_noMatches(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_peering_connections" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
