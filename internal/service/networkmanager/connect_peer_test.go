// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerConnectPeer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectPeer
	resourceName := "aws_networkmanager_connect_peer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	insideCidrBlocksv4 := "169.254.10.0/29"
	peerAddress := "1.1.1.1"
	protocol := awstypes.TunnelProtocolGre
	asn := "65501"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectPeerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectPeerConfig_basic(rName, insideCidrBlocksv4, peerAddress, asn, protocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectPeerExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNRegexp("networkmanager", regexache.MustCompile(`connect-peer/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"bgp_configurations":   knownvalue.ListSizeExact(2),
							"core_network_address": knownvalue.NotNull(),
							"inside_cidr_blocks":   knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact(insideCidrBlocksv4)}),
							"peer_address":         knownvalue.StringExact(peerAddress),
							names.AttrProtocol:     tfknownvalue.StringExact(protocol),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("connect_peer_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("core_network_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreatedAt), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("edge_location"), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerConnectPeer_noDependsOn(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectPeer
	resourceName := "aws_networkmanager_connect_peer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	insideCidrBlocksv4 := "169.254.10.0/29"
	peerAddress := "1.1.1.1"
	protocol := awstypes.TunnelProtocolGre
	asn := "65501"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectPeerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectPeerConfig_noDependsOn(rName, insideCidrBlocksv4, peerAddress, asn, protocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectPeerExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerConnectPeer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectPeer
	resourceName := "aws_networkmanager_connect_peer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	insideCidrBlocksv4 := "169.254.10.0/29"
	peerAddress := "1.1.1.1"
	protocol := awstypes.TunnelProtocolGre
	asn := "65501"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectPeerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectPeerConfig_basic(rName, insideCidrBlocksv4, peerAddress, asn, protocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectPeerExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfnetworkmanager.ResourceConnectPeer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccNetworkManagerConnectPeer_subnetARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectPeer
	resourceName := "aws_networkmanager_connect_peer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	peerAddress := "10.0.2.100" // Must be an address inside the subnet CIDR range.
	protocol := awstypes.TunnelProtocolNoEncap

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectPeerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectPeerConfig_subnetARN(rName, peerAddress, protocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectPeerExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("bgp_options"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"peer_asn": knownvalue.NotNull(),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerConnectPeer_4BytePeerASN(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectPeer
	resourceName := "aws_networkmanager_connect_peer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	insideCidrBlocksv4 := "169.254.10.0/29"
	peerAddress := "1.1.1.1"
	protocol := awstypes.TunnelProtocolGre
	asn := "4294967290"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectPeerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectPeerConfig_basic(rName, insideCidrBlocksv4, peerAddress, asn, protocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectPeerExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerConnectPeer_upgradeFromV6_26_0(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectPeer
	resourceName := "aws_networkmanager_connect_peer.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	insideCidrBlocksv4 := "169.254.10.0/29"
	peerAddress := "1.1.1.1"
	protocol := awstypes.TunnelProtocolGre
	asn := "65501"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		CheckDestroy: testAccCheckConnectPeerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.26.0",
					},
				},
				Config: testAccConnectPeerConfig_basic(rName, insideCidrBlocksv4, peerAddress, asn, protocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectPeerExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccConnectPeerConfig_basic(rName, insideCidrBlocksv4, peerAddress, asn, protocol),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectPeerExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckConnectPeerExists(ctx context.Context, t *testing.T, n string, v *awstypes.ConnectPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		output, err := tfnetworkmanager.FindConnectPeerByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectPeerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_connect_peer" {
				continue
			}

			_, err := tfnetworkmanager.FindConnectPeerByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Connect Peer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConnectPeerConfig_base(rName string, protocol awstypes.TunnelProtocol) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 2), fmt.Sprintf(`
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
    vpn_ecmp_support   = false
    asn_ranges         = ["64512-64555"]
    inside_cidr_blocks = ["172.16.0.0/16"]
    edge_locations {
      location           = data.aws_region.current.region
      asn                = 64512
      inside_cidr_blocks = ["172.16.0.0/18"]
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

resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn
  tags = {
    segment = "shared"
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}

resource "aws_networkmanager_connect_attachment" "test" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location
  options {
    protocol = %[2]q
  }
  tags = {
    segment = "shared"
  }
  depends_on = [
    "aws_networkmanager_attachment_accepter.test"
  ]
}

resource "aws_networkmanager_attachment_accepter" "test2" {
  attachment_id   = aws_networkmanager_connect_attachment.test.id
  attachment_type = aws_networkmanager_connect_attachment.test.attachment_type
}
`, rName, protocol))
}

func testAccConnectPeerConfig_basic(rName, insideCidrBlocks, peerAddress, asn string, protocol awstypes.TunnelProtocol) string {
	return acctest.ConfigCompose(testAccConnectPeerConfig_base(rName, protocol), fmt.Sprintf(`
resource "aws_networkmanager_connect_peer" "test" {
  connect_attachment_id = aws_networkmanager_connect_attachment.test.id
  peer_address          = %[2]q

  bgp_options {
    peer_asn = %[3]q
  }

  inside_cidr_blocks = [
	%[1]q
  ]

  depends_on = [
    "aws_networkmanager_attachment_accepter.test"
  ]
}
`, insideCidrBlocks, peerAddress, asn))
}

func testAccConnectPeerConfig_noDependsOn(rName, insideCidrBlocks, peerAddress, asn string, protocol awstypes.TunnelProtocol) string {
	return acctest.ConfigCompose(testAccConnectPeerConfig_base(rName, protocol), fmt.Sprintf(`
resource "aws_networkmanager_connect_peer" "test" {
  connect_attachment_id = aws_networkmanager_connect_attachment.test.id
  peer_address          = %[2]q

  bgp_options {
    peer_asn = %[3]q
  }

  inside_cidr_blocks = [
	%[1]q
  ]
}
`, insideCidrBlocks, peerAddress, asn))
}

func testAccConnectPeerConfig_subnetARN(rName, peerAddress string, protocol awstypes.TunnelProtocol) string {
	return acctest.ConfigCompose(testAccConnectPeerConfig_base(rName, protocol), fmt.Sprintf(`
resource "aws_networkmanager_connect_peer" "test" {
  connect_attachment_id = aws_networkmanager_connect_attachment.test.id
  peer_address          = %[2]q
  subnet_arn            = aws_subnet.test2.arn

  depends_on = [
    "aws_networkmanager_attachment_accepter.test"
  ]
}

resource "aws_subnet" "test2" {
  vpc_id     = aws_vpc.test.id
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 2)

  tags = {
    Name = %[1]q
  }
}
`, rName, peerAddress))
}
