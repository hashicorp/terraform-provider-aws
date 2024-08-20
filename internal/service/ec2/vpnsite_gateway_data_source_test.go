// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSiteVPNGatewayDataSource_unattached(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceNameById := "data.aws_vpn_gateway.test_by_id"
	dataSourceNameByTags := "data.aws_vpn_gateway.test_by_tags"
	dataSourceNameByAsn := "data.aws_vpn_gateway.test_by_amazon_side_asn"
	resourceName := "aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNGatewayDataSourceConfig_unattached(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceNameById, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceNameById, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceNameByAsn, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceNameById, names.AttrState),
					resource.TestCheckResourceAttr(dataSourceNameByTags, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckNoResourceAttr(dataSourceNameById, "attached_vpc_id"),
					resource.TestCheckResourceAttr(dataSourceNameByAsn, "amazon_side_asn", "4294967293"),
				),
			},
		},
	})
}

func TestAccSiteVPNGatewayDataSource_attached(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNGatewayDataSourceConfig_attached(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, "aws_vpn_gateway.test", names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "attached_vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrState, regexache.MustCompile("(?i)available")),
				),
			},
		},
	})
}

func testAccSiteVPNGatewayDataSourceConfig_unattached(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
    ABC  = "abc"
    XYZ  = "xyz"
  }

  amazon_side_asn = 4294967293
}

data "aws_vpn_gateway" "test_by_id" {
  id = aws_vpn_gateway.test.id
}

data "aws_vpn_gateway" "test_by_tags" {
  tags = aws_vpn_gateway.test.tags
}

data "aws_vpn_gateway" "test_by_amazon_side_asn" {
  amazon_side_asn = aws_vpn_gateway.test.amazon_side_asn
  state           = "available"
}
`, rName)
}

func testAccSiteVPNGatewayDataSourceConfig_attached(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = aws_vpn_gateway.test.id
}

data "aws_vpn_gateway" "test" {
  attached_vpc_id = aws_vpn_gateway_attachment.test.vpc_id
}
`, rName)
}
