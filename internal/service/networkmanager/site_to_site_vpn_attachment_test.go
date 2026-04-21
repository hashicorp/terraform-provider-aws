// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerSiteToSiteVPNAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SiteToSiteVpnAttachment
	resourceName := "aws_networkmanager_site_to_site_vpn_attachment.test"
	vpnResourceName := "aws_vpn_connection.test"
	coreNetworkResourceName := "aws_networkmanager_core_network.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bgpASN := acctest.RandIntRange(t, 64512, 65534)
	vpnIP, err := sdkacctest.RandIpAddress("172.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteToSiteVPNAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteToSiteVPNAttachmentConfig_basic(rName, bgpASN, vpnIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSiteToSiteVPNAttachmentExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "SITE_TO_SITE_VPN"),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_arn", coreNetworkResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, vpnResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_connection_arn", vpnResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrState},
			},
		},
	})
}

func TestAccNetworkManagerSiteToSiteVPNAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SiteToSiteVpnAttachment
	resourceName := "aws_networkmanager_site_to_site_vpn_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bgpASN := acctest.RandIntRange(t, 64512, 65534)
	vpnIP, err := sdkacctest.RandIpAddress("172.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteToSiteVPNAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteToSiteVPNAttachmentConfig_basic(rName, bgpASN, vpnIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteToSiteVPNAttachmentExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfnetworkmanager.ResourceSiteToSiteVPNAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerSiteToSiteVPNAttachment_routingPolicyLabel(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SiteToSiteVpnAttachment
	resourceName := "aws_networkmanager_site_to_site_vpn_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bgpASN := acctest.RandIntRange(t, 64512, 65534)
	vpnIP, err := sdkacctest.RandIpAddress("172.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	label := "testlabel"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteToSiteVPNAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteToSiteVPNAttachmentConfig_routingPolicyLabel(rName, bgpASN, vpnIP, label),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteToSiteVPNAttachmentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrState},
			},
		},
	})
}

func TestAccNetworkManagerSiteToSiteVPNAttachment_routingPolicyLabelUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SiteToSiteVpnAttachment
	resourceName := "aws_networkmanager_site_to_site_vpn_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bgpASN := acctest.RandIntRange(t, 64512, 65534)
	vpnIP, err := sdkacctest.RandIpAddress("172.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	label1 := "testlabel1"
	label2 := "testlabel2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteToSiteVPNAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteToSiteVPNAttachmentConfig_routingPolicyLabel(rName, bgpASN, vpnIP, label1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteToSiteVPNAttachmentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label1),
				),
			},
			{
				Config: testAccSiteToSiteVPNAttachmentConfig_routingPolicyLabel(rName, bgpASN, vpnIP, label2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSiteToSiteVPNAttachmentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label2),
				),
			},
		},
	})
}

func testAccCheckSiteToSiteVPNAttachmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.SiteToSiteVpnAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		output, err := tfnetworkmanager.FindSiteToSiteVPNAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSiteToSiteVPNAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_site_to_site_vpn_attachment" {
				continue
			}

			_, err := tfnetworkmanager.FindSiteToSiteVPNAttachmentByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Site To Site VPN Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSiteToSiteVPNAttachmentConfig_base(rName string, bgpASN int, vpnIP string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn     = %[2]d
  ip_address  = %[3]q
  type        = "ipsec.1"
  device_name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

data "aws_region" "current" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.current.region
      asn      = 64512
    }
  }

  segments {
    name                          = "shared"
    description                   = "SegmentForSharedServices"
    require_attachment_acceptance = true
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }

  attachment_policies {
    rule_number     = 1
    condition_logic = "or"

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "segment"
      value    = "shared"
    }

    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}
`, rName, bgpASN, vpnIP)
}

func testAccSiteToSiteVPNAttachmentConfig_basic(rName string, bgpASN int, vpnIP string) string {
	return acctest.ConfigCompose(testAccSiteToSiteVPNAttachmentConfig_base(rName, bgpASN, vpnIP), `
resource "aws_networkmanager_site_to_site_vpn_attachment" "test" {
  core_network_id    = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpn_connection_arn = aws_vpn_connection.test.arn

  tags = {
    segment = "shared"
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_site_to_site_vpn_attachment.test.id
  attachment_type = aws_networkmanager_site_to_site_vpn_attachment.test.attachment_type
}
`)
}

func testAccSiteToSiteVPNAttachmentConfig_routingPolicyLabel(rName string, bgpASN int, vpnIP, label string) string {
	return acctest.ConfigCompose(testAccSiteToSiteVPNAttachmentConfig_baseWithRoutingPolicy(rName, bgpASN, vpnIP, label), fmt.Sprintf(`
resource "aws_networkmanager_site_to_site_vpn_attachment" "test" {
  core_network_id      = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpn_connection_arn   = aws_vpn_connection.test.arn
  routing_policy_label = %[1]q

  tags = {
    segment = "shared"
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_site_to_site_vpn_attachment.test.id
  attachment_type = aws_networkmanager_site_to_site_vpn_attachment.test.attachment_type
}
`, label))
}

func testAccSiteToSiteVPNAttachmentConfig_baseWithRoutingPolicy(rName string, bgpASN int, vpnIP, label string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn     = %[2]d
  ip_address  = %[3]q
  type        = "ipsec.1"
  device_name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

data "aws_region" "current" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["64512-64555"]

    edge_locations {
      location = data.aws_region.current.region
    }
  }

  segments {
    name                          = "shared"
    require_attachment_acceptance = false
  }

  attachment_policies {
    rule_number     = 100
    condition_logic = "or"

    conditions {
      type = "tag-exists"
      key  = "segment"
    }

    action {
      association_method = "tag"
      tag_value_of_key   = "segment"
    }
  }

  routing_policies {
    routing_policy_name      = "policy1"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1

      rule_definition {
        match_conditions {
          type  = "prefix-in-cidr"
          value = "10.0.0.0/8"
        }

        action {
          type = "allow"
        }
      }
    }
  }

  attachment_routing_policy_rules {
    rule_number = 1

    conditions {
      type  = "routing-policy-label"
      value = %[4]q
    }

    action {
      associate_routing_policies = ["policy1"]
    }
  }
}
`, rName, bgpASN, vpnIP, label)
}
