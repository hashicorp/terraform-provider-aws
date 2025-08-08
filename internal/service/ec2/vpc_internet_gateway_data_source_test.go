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

func TestAccVPCInternetGatewayDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	igwResourceName := "aws_internet_gateway.test"
	vpcResourceName := "aws_vpc.test"
	ds1ResourceName := "data.aws_internet_gateway.by_id"
	ds2ResourceName := "data.aws_internet_gateway.by_filter"
	ds3ResourceName := "data.aws_internet_gateway.by_tags"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCInternetGatewayDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "internet_gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrOwnerID, igwResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "attachments.0.vpc_id", vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds1ResourceName, names.AttrARN, igwResourceName, names.AttrARN),

					resource.TestCheckResourceAttrPair(ds2ResourceName, "internet_gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds2ResourceName, names.AttrOwnerID, igwResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "attachments.0.vpc_id", vpcResourceName, names.AttrID),

					resource.TestCheckResourceAttrPair(ds3ResourceName, "internet_gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(ds3ResourceName, names.AttrOwnerID, igwResourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "attachments.0.vpc_id", vpcResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccVPCInternetGatewayDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_internet_gateway" "by_id" {
  internet_gateway_id = aws_internet_gateway.test.id
}

data "aws_internet_gateway" "by_tags" {
  tags = {
    Name = aws_internet_gateway.test.tags["Name"]
  }
}

data "aws_internet_gateway" "by_filter" {
  filter {
    name   = "internet-gateway-id"
    values = [aws_internet_gateway.test.id]
  }
}
`, rName)
}
