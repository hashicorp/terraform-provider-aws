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

func TestAccDirectConnectPublicVirtualInterface_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_public_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", sdkacctest.RandString(10))
	// DirectConnectClientException: Amazon Address is not allowed to contain a private IP
	// DirectConnectClientException: Amazon Address and Customer Address must be in the same CIDR
	// DirectConnectClientException: Amazon Address is address 0 on its subnet.
	// DirectConnectClientException: Amazon Address is the broadcast address on its subnet.
	amazonAddress := "175.45.176.1/28"
	customerAddress := "175.45.176.2/28"
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicVirtualInterfaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicVirtualInterfaceConfig_basic(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
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

func TestAccDirectConnectPublicVirtualInterface_tags(t *testing.T) {
	ctx := acctest.Context(t)
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_public_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", sdkacctest.RandString(10))
	amazonAddress := "175.45.176.3/28"
	customerAddress := "175.45.176.4/28"
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)
	vlan := sdkacctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicVirtualInterfaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicVirtualInterfaceConfig_tags(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
				),
			},
			{
				Config: testAccPublicVirtualInterfaceConfig_tagsUpdated(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicVirtualInterfaceExists(ctx, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionId),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
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

func testAccCheckPublicVirtualInterfaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckVirtualInterfaceDestroy(ctx, s, "aws_dx_public_virtual_interface")
	}
}

func testAccCheckPublicVirtualInterfaceExists(ctx context.Context, name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckVirtualInterfaceExists(ctx, name, vif)
}

func testAccPublicVirtualInterfaceConfig_basic(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccPublicVirtualInterfaceConfig_tags(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]

  tags = {
    Name = %[2]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccPublicVirtualInterfaceConfig_tagsUpdated(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]

  tags = {
    Name = %[2]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}
