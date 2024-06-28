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

func TestAccDirectConnectPrivateVirtualInterface_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", sdkacctest.RandString(9))
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_basic(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
				),
			},
			{
				Config: testAccPrivateVirtualInterfaceConfig_updated(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "9001"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectPrivateVirtualInterface_tags(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", sdkacctest.RandString(9))
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_tags(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
				),
			},
			{
				Config: testAccPrivateVirtualInterfaceConfig_tagsUpdated(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectPrivateVirtualInterface_dxGateway(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", sdkacctest.RandString(9))
	amzAsn := sdkacctest.RandIntRange(64512, 65534)
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_gateway(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectPrivateVirtualInterface_siteLink(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", sdkacctest.RandString(9))
	amzAsn := sdkacctest.RandIntRange(64512, 65534)
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateVirtualInterfaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateVirtualInterfaceConfig_siteLinkBasic(connectionId, rName, amzAsn, bgpAsn, vlan, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttr(resourceName, "sitelink_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccPrivateVirtualInterfaceConfig_siteLinkUpdated(connectionId, rName, amzAsn, bgpAsn, vlan, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttr(resourceName, "sitelink_enabled", acctest.CtFalse),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPrivateVirtualInterfaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckVirtualInterfaceDestroy(ctx, s, "aws_dx_private_virtual_interface")
	}
}

func testAccCheckPrivateVirtualInterfaceExists(ctx context.Context, name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckVirtualInterfaceExists(ctx, name, vif)
}

func testAccPrivateVirtualInterfaceConfig_vpnGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccPrivateVirtualInterfaceConfig_basic(cid, rName string, bgpAsn, vlan int) string {
	return testAccPrivateVirtualInterfaceConfig_vpnGateway(rName) + fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, cid, rName, bgpAsn, vlan)
}

func testAccPrivateVirtualInterfaceConfig_updated(cid, rName string, bgpAsn, vlan int) string {
	return testAccPrivateVirtualInterfaceConfig_vpnGateway(rName) + fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  mtu            = 9001
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, cid, rName, bgpAsn, vlan)
}

func testAccPrivateVirtualInterfaceConfig_tags(cid, rName string, bgpAsn, vlan int) string {
	return testAccPrivateVirtualInterfaceConfig_vpnGateway(rName) + fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id

  tags = {
    Name = %[2]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, cid, rName, bgpAsn, vlan)
}

func testAccPrivateVirtualInterfaceConfig_tagsUpdated(cid, rName string, bgpAsn, vlan int) string {
	return testAccPrivateVirtualInterfaceConfig_vpnGateway(rName) + fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id

  tags = {
    Name = %[2]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, cid, rName, bgpAsn, vlan)
}

func testAccPrivateVirtualInterfaceConfig_gateway(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = %[3]d
  name            = %[2]q
}

resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[4]d
  dx_gateway_id  = aws_dx_gateway.test.id
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[5]d
}
`, cid, rName, amzAsn, bgpAsn, vlan)
}

func testAccPrivateVirtualInterfaceConfig_siteLinkBasic(cid, rName string, amzAsn, bgpAsn, vlan int, sitelink_enabled bool) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = %[3]d
  name            = %[2]q
}

resource "aws_dx_private_virtual_interface" "test" {
  address_family   = "ipv4"
  bgp_asn          = %[4]d
  dx_gateway_id    = aws_dx_gateway.test.id
  connection_id    = %[1]q
  name             = %[2]q
  vlan             = %[5]d
  sitelink_enabled = %[6]t
}
`, cid, rName, amzAsn, bgpAsn, vlan, sitelink_enabled)
}

func testAccPrivateVirtualInterfaceConfig_siteLinkUpdated(cid, rName string, amzAsn, bgpAsn, vlan int, sitelink_enabled bool) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = %[3]d
  name            = %[2]q
}

resource "aws_dx_private_virtual_interface" "test" {
  address_family   = "ipv4"
  bgp_asn          = %[4]d
  dx_gateway_id    = aws_dx_gateway.test.id
  connection_id    = %[1]q
  name             = %[2]q
  vlan             = %[5]d
  sitelink_enabled = %[6]t
}
`, cid, rName, amzAsn, bgpAsn, vlan, sitelink_enabled)
}
