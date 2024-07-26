// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectHostedTransitVirtualInterface_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccHostedTransitVirtualInterface_basic,
		"accepterTags":  testAccHostedTransitVirtualInterface_accepterTags,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccHostedTransitVirtualInterface_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_hosted_transit_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_transit_virtual_interface_accepter.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-transit-vif-%s", sdkacctest.RandString(9))
	amzAsn := sdkacctest.RandIntRange(64512, 65534)
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckHostedTransitVirtualInterfaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedTransitVirtualInterfaceConfig_basic(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedTransitVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(accepterResourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(accepterResourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
				),
			},
			// Test import.
			{
				Config:            testAccHostedTransitVirtualInterfaceConfig_basic(connectionId, rName, amzAsn, bgpAsn, vlan),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHostedTransitVirtualInterface_accepterTags(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_hosted_transit_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_transit_virtual_interface_accepter.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-transit-vif-%s", sdkacctest.RandString(9))
	amzAsn := sdkacctest.RandIntRange(64512, 65534)
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckHostedTransitVirtualInterfaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedTransitVirtualInterfaceConfig_accepterTags(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedTransitVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(accepterResourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(accepterResourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
				),
			},
			{
				Config: testAccHostedTransitVirtualInterfaceConfig_accepterTagsUpdated(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedTransitVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(accepterResourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(accepterResourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckHostedTransitVirtualInterfaceExists(ctx context.Context, name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckVirtualInterfaceExists(ctx, name, vif)
}

func testAccCheckHostedTransitVirtualInterfaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckVirtualInterfaceDestroy(ctx, s, "aws_dx_hosted_transit_virtual_interface")
	}
}

func testAccHostedTransitVirtualInterfaceConfig_base(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_transit_virtual_interface" "test" {
  address_family   = "ipv4"
  bgp_asn          = %[4]d
  connection_id    = %[1]q
  name             = %[2]q
  owner_account_id = data.aws_caller_identity.accepter.account_id
  vlan             = %[5]d

  # The aws_dx_hosted_transit_virtual_interface
  # must be destroyed before the aws_dx_gateway.
  depends_on = [aws_dx_gateway.test]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "awsalternate"
}

resource "aws_dx_gateway" "test" {
  provider = "awsalternate"

  amazon_side_asn = %[3]d
  name            = %[2]q
}
`, cid, rName, amzAsn, bgpAsn, vlan)
}

func testAccHostedTransitVirtualInterfaceConfig_basic(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccHostedTransitVirtualInterfaceConfig_base(cid, rName, amzAsn, bgpAsn, vlan) + `
resource "aws_dx_hosted_transit_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  dx_gateway_id        = aws_dx_gateway.test.id
  virtual_interface_id = aws_dx_hosted_transit_virtual_interface.test.id
}
`
}

func testAccHostedTransitVirtualInterfaceConfig_accepterTags(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccHostedTransitVirtualInterfaceConfig_base(cid, rName, amzAsn, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_transit_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  dx_gateway_id        = aws_dx_gateway.test.id
  virtual_interface_id = aws_dx_hosted_transit_virtual_interface.test.id

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName)
}

func testAccHostedTransitVirtualInterfaceConfig_accepterTagsUpdated(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccHostedTransitVirtualInterfaceConfig_base(cid, rName, amzAsn, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_transit_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  dx_gateway_id        = aws_dx_gateway.test.id
  virtual_interface_id = aws_dx_hosted_transit_virtual_interface.test.id

  tags = {
    Name = %[1]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, rName)
}
