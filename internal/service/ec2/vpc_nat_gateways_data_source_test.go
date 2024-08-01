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

func TestAccVPCNATGatewaysDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewaysDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_nat_gateways.by_vpc_id", "ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr("data.aws_nat_gateways.by_tags", "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr("data.aws_nat_gateways.by_filter", "ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr("data.aws_nat_gateways.empty", "ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccVPCNATGatewaysDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "172.5.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "172.5.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test1.id
  cidr_block        = "172.5.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test2.id
  cidr_block        = "172.5.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test3" {
  vpc_id            = aws_vpc.test2.id
  cidr_block        = "172.5.124.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test1" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test2" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test3" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test1" {
  vpc_id = aws_vpc.test1.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test2" {
  vpc_id = aws_vpc.test2.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test1" {
  subnet_id     = aws_subnet.test1.id
  allocation_id = aws_eip.test1.id

  tags = {
    Name     = %[1]q
    OtherTag = "some-value"
  }

  depends_on = [aws_internet_gateway.test1]
}

resource "aws_nat_gateway" "test2" {
  subnet_id     = aws_subnet.test2.id
  allocation_id = aws_eip.test2.id

  tags = {
    Name     = %[1]q
    OtherTag = "some-other-value"
  }

  depends_on = [aws_internet_gateway.test2]
}

resource "aws_nat_gateway" "test3" {
  subnet_id     = aws_subnet.test3.id
  allocation_id = aws_eip.test3.id

  tags = {
    Name     = %[1]q
    OtherTag = "some-other-value"
  }

  depends_on = [aws_internet_gateway.test2]
}

data "aws_nat_gateways" "by_vpc_id" {
  vpc_id = aws_vpc.test2.id

  depends_on = [aws_nat_gateway.test1, aws_nat_gateway.test2, aws_nat_gateway.test3]
}

data "aws_nat_gateways" "by_tags" {
  filter {
    name   = "state"
    values = ["available"]
  }

  tags = {
    OtherTag = "some-value"
  }

  depends_on = [aws_nat_gateway.test1, aws_nat_gateway.test2, aws_nat_gateway.test3]
}

data "aws_nat_gateways" "by_filter" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test1.id, aws_vpc.test2.id]
  }

  depends_on = [aws_nat_gateway.test1, aws_nat_gateway.test2, aws_nat_gateway.test3]
}

data "aws_nat_gateways" "empty" {
  vpc_id = aws_vpc.test2.id

  tags = {
    OtherTag = "some-value"
  }

  depends_on = [aws_nat_gateway.test1, aws_nat_gateway.test2, aws_nat_gateway.test3]
}
`, rName))
}
