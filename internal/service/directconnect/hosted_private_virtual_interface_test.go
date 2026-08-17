// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectHostedPrivateVirtualInterface_basic(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_hosted_private_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_private_virtual_interface_accepter.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(t, 9))
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckHostedPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPrivateVirtualInterfaceConfig_basic(connectionID, rName, bgpAsn, vlan),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostedPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, names.AttrARN), // nosemgrep:ci.semgrep.acctest.checks.arn-resourceattrset // TODO: need DX Connection for testing
					resource.TestCheckResourceAttr(accepterResourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(accepterResourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
				),
			},
			{
				Config:            testAccHostedPrivateVirtualInterfaceConfig_basic(connectionID, rName, bgpAsn, vlan),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectHostedPrivateVirtualInterface_siteLink(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	var vifID string
	resourceName := "aws_dx_hosted_private_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_private_virtual_interface_accepter.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-hosted-private-vif-%s", acctest.RandString(t, 9))
	amazonSideASN := acctest.RandIntRange(t, 64512, 65534)
	bgpASN := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckHostedPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPrivateVirtualInterfaceConfig_siteLink(connectionID, rName, amazonSideASN, bgpASN, vlan, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostedPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					testAccCheckHostedPrivateVirtualInterfaceID(resourceName, &vifID),
					resource.TestCheckResourceAttrPair(accepterResourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(accepterResourceName, "sitelink_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccHostedPrivateVirtualInterfaceConfig_siteLink(connectionID, rName, amazonSideASN, bgpASN, vlan, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(accepterResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostedPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					testAccCheckHostedPrivateVirtualInterfaceID(resourceName, &vifID),
					resource.TestCheckResourceAttrPair(accepterResourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(accepterResourceName, "sitelink_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccHostedPrivateVirtualInterfaceConfig_siteLink(connectionID, rName, amazonSideASN, bgpASN, vlan, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(accepterResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostedPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					testAccCheckHostedPrivateVirtualInterfaceID(resourceName, &vifID),
					resource.TestCheckResourceAttrPair(accepterResourceName, "dx_gateway_id", dxGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(accepterResourceName, "sitelink_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      accepterResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDirectConnectHostedPrivateVirtualInterface_accepterTags(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")

	var vif awstypes.VirtualInterface
	resourceName := "aws_dx_hosted_private_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_private_virtual_interface_accepter.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(t, 9))
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	vlan := acctest.RandIntRange(t, 2049, 4094)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckHostedPrivateVirtualInterfaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedPrivateVirtualInterfaceConfig_accepterTags(connectionID, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, names.AttrARN), // nosemgrep:ci.semgrep.acctest.checks.arn-resourceattrset // TODO: need DX Connection for testing
					resource.TestCheckResourceAttr(accepterResourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(accepterResourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
				),
			},
			{
				Config: testAccHostedPrivateVirtualInterfaceConfig_accepterTagsUpdated(connectionID, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedPrivateVirtualInterfaceExists(ctx, t, resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "directconnect", regexache.MustCompile(fmt.Sprintf("dxvif/%s", aws.ToString(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, names.AttrARN), // nosemgrep:ci.semgrep.acctest.checks.arn-resourceattrset // TODO: need DX Connection for testing
					resource.TestCheckResourceAttr(accepterResourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(accepterResourceName, "vpn_gateway_id", vpnGatewayResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckHostedPrivateVirtualInterfaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckVirtualInterfaceDestroy(ctx, t, s, "aws_dx_hosted_private_virtual_interface")
	}
}

func testAccCheckHostedPrivateVirtualInterfaceExists(ctx context.Context, t *testing.T, name string, vif *awstypes.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckVirtualInterfaceExists(ctx, t, name, vif)
}

func testAccCheckHostedPrivateVirtualInterfaceID(name string, expectedID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		if *expectedID == "" {
			*expectedID = resource.Primary.ID
			return nil
		}
		if resource.Primary.ID != *expectedID {
			return fmt.Errorf("expected Direct Connect hosted private virtual interface ID %s, got %s", *expectedID, resource.Primary.ID)
		}

		return nil
	}
}

func testAccHostedPrivateVirtualInterfaceConfig_base(cid, rName string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_private_virtual_interface" "test" {
  address_family   = "ipv4"
  bgp_asn          = %[3]d
  connection_id    = %[1]q
  name             = %[2]q
  owner_account_id = data.aws_caller_identity.accepter.account_id
  vlan             = %[4]d

  # The aws_dx_hosted_private_virtual_interface
  # must be destroyed before the aws_vpn_gateway.
  depends_on = [aws_vpn_gateway.test]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "awsalternate"
}

resource "aws_vpn_gateway" "test" {
  provider = "awsalternate"

  tags = {
    Name = %[2]q
  }
}
`, cid, rName, bgpAsn, vlan))
}

func testAccHostedPrivateVirtualInterfaceConfig_siteLink(cid, rName string, amazonSideASN, bgpASN, vlan int, siteLinkEnabled bool) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_private_virtual_interface" "test" {
  address_family   = "ipv4"
  bgp_asn          = %[4]d
  connection_id    = %[1]q
  name             = %[2]q
  owner_account_id = data.aws_caller_identity.accepter.account_id
  vlan             = %[5]d

  # The aws_dx_hosted_private_virtual_interface
  # must be destroyed before the aws_dx_gateway.
  depends_on = [aws_dx_gateway.test]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "awsalternate"
}

resource "aws_dx_gateway" "test" {
  provider        = "awsalternate"
  amazon_side_asn = %[3]d
  name            = %[2]q
}

resource "aws_dx_hosted_private_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  dx_gateway_id        = aws_dx_gateway.test.id
  sitelink_enabled     = %[6]t
  virtual_interface_id = aws_dx_hosted_private_virtual_interface.test.id
}
`, cid, rName, amazonSideASN, bgpASN, vlan, siteLinkEnabled))
}

func testAccHostedPrivateVirtualInterfaceConfig_basic(cid, rName string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccHostedPrivateVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan), `
resource "aws_dx_hosted_private_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_private_virtual_interface.test.id
  vpn_gateway_id       = aws_vpn_gateway.test.id
}
`)
}

func testAccHostedPrivateVirtualInterfaceConfig_accepterTags(cid, rName string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccHostedPrivateVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan), fmt.Sprintf(`
resource "aws_dx_hosted_private_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_private_virtual_interface.test.id
  vpn_gateway_id       = aws_vpn_gateway.test.id

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName))
}

func testAccHostedPrivateVirtualInterfaceConfig_accepterTagsUpdated(cid, rName string, bgpAsn, vlan int) string {
	return acctest.ConfigCompose(testAccHostedPrivateVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan), fmt.Sprintf(`
resource "aws_dx_hosted_private_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_private_virtual_interface.test.id
  vpn_gateway_id       = aws_vpn_gateway.test.id

  tags = {
    Name = %[1]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, rName))
}
