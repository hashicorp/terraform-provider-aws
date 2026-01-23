// Copyright IBM Corp. 2014, 2026
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

func TestAccVPCEgressOnlyInternetGatewayDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	eigwResourceName := "aws_egress_only_internet_gateway.test"
	vpcResourceName := "aws_vpc.test"
	ds1ResourceName := "data.aws_egress_only_internet_gateway.by_id"
	ds2ResourceName := "data.aws_egress_only_internet_gateway.by_filter"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEgressOnlyInternetGatewayDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrID, eigwResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(ds1ResourceName, names.AttrState, "attached"),

					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrID, eigwResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(ds2ResourceName, names.AttrState, "attached"),
				),
			},
		},
	})
}

func testAccVPCEgressOnlyInternetGatewayDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_egress_only_internet_gateway" "by_id" {
  id = aws_egress_only_internet_gateway.test.id
}

data "aws_egress_only_internet_gateway" "by_filter" {
  filter {
    name   = "tag:Name"
    values = [%[1]q]
  }

  depends_on = [aws_egress_only_internet_gateway.test]
}
`, rName)
}
