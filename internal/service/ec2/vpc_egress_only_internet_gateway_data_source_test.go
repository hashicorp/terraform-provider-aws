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

func TestVPCEgressOnlyInternetGatewayDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEgressOnlyInternetGatewayDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_egress_only_internet_gateway.by_id", "egress_only_internet_gateway_id", "aws_egress_only_internet_gateway.test", names.AttrID),
					resource.TestCheckResourceAttrPair("data.aws_egress_only_internet_gateway.by_id", names.AttrOwnerID, "aws_egress_only_internet_gateway.test", names.AttrOwnerID),
					resource.TestCheckResourceAttrPair("data.aws_egress_only_internet_gateway.by_id", "attachments.0.vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttrPair("data.aws_egress_only_internet_gateway.by_id", names.AttrARN, "aws_egress_only_internet_gateway.test", names.AttrARN),
				),
			},
		},
	})
}

func TestVPCEgressOnlyInternetGatewayDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEgressOnlyInternetGatewayDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_egress_only_internet_gateway.by_tags", "egress_only_internet_gateway_id", "aws_egress_only_internet_gateway.test", names.AttrID),
					resource.TestCheckResourceAttrPair("data.aws_egress_only_internet_gateway.by_tags", names.AttrOwnerID, "aws_egress_only_internet_gateway.test", names.AttrOwnerID),
					resource.TestCheckResourceAttrPair("data.aws_egress_only_internet_gateway.by_tags", "attachments.0.vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttrPair("data.aws_egress_only_internet_gateway.by_tags", names.AttrARN, "aws_egress_only_internet_gateway.test", names.AttrARN),
				),
			},
		},
	})
}

func testAccVPCEgressOnlyInternetGatewayDataSourceConfig_basic() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

data "aws_egress_only_internet_gateway" "by_id" {
  egress_only_internet_gateway_id = aws_egress_only_internet_gateway.test.id
}
`
}

func testAccVPCEgressOnlyInternetGatewayDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_egress_only_internet_gateway" "by_tags" {
  tags = {
    Name = %[1]q
  }

  depends_on = [
    aws_egress_only_internet_gateway.test,
  ]
}
`, rName)
}
