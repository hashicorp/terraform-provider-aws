// Copyright IBM Corp. 2014, 2025
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

func TestAccSiteVPNConnectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(65501, 65534)
	dataSourceName := "data.aws_vpn_connection.test"
	resourceName := "aws_vpn_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPNConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionDataSourceConfig_byId(rName, rBgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "vpn_connection_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "customer_gateway_id", resourceName, "customer_gateway_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpn_gateway_id", resourceName, "vpn_gateway_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(dataSourceName, "customer_gateway_configuration"),
					resource.TestCheckResourceAttrSet(dataSourceName, "category"),
					resource.TestCheckResourceAttr(dataSourceName, "routes.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "vgw_telemetries.#", "2"),
				),
			},
		},
	})
}

func TestAccSiteVPNConnectionDataSource_byFilter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(65501, 65534)
	dataSourceName := "data.aws_vpn_connection.test"
	resourceName := "aws_vpn_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPNConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionDataSourceConfig_byFilter(rName, rBgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "vpn_connection_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "customer_gateway_id", resourceName, "customer_gateway_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpn_gateway_id", resourceName, "vpn_gateway_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(dataSourceName, "customer_gateway_configuration"),
				),
			},
		},
	})
}

func TestAccSiteVPNConnectionDataSource_nonExistentId(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPNConnectionDataSourceConfig_nonExistentId(),
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func TestAccSiteVPNConnectionDataSource_noInput(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPNConnectionDataSourceConfig_noInput(),
				ExpectError: regexache.MustCompile(`Missing Attribute Configuration`),
			},
		},
	})
}

func testAccVPNConnectionDataSourceConfigBase(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccVPNConnectionDataSourceConfig_byId(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccVPNConnectionDataSourceConfigBase(rName, rBgpAsn),
		`
data "aws_vpn_connection" "test" {
  vpn_connection_id = aws_vpn_connection.test.id
}
`)
}

func testAccVPNConnectionDataSourceConfig_byFilter(rName string, rBgpAsn int) string {
	return acctest.ConfigCompose(
		testAccVPNConnectionDataSourceConfigBase(rName, rBgpAsn),
		`
data "aws_vpn_connection" "test" {
  filter {
    name   = "customer-gateway-id"
    values = [aws_customer_gateway.test.id]
  }

  filter {
    name   = "vpn-gateway-id"
    values = [aws_vpn_gateway.test.id]
  }

  depends_on = [aws_vpn_connection.test]
}
`)
}

func testAccVPNConnectionDataSourceConfig_nonExistentId() string {
	return `
data "aws_vpn_connection" "test" {
  vpn_connection_id = "vpn-12345678901234567"
}
`
}

func testAccVPNConnectionDataSourceConfig_noInput() string {
	return `
data "aws_vpn_connection" "test" {
  # No vpn_connection_id or filter specified
}
`
}
