// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCRouteServerPeer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RouteServerPeer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpc_route_server_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerPeerConfig_basic(rName, rAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerPeerExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`route-server-peer/rsp-[a-z0-9]+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("bgp_options"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"peer_asn":                knownvalue.Int64Exact(65000),
							"peer_liveness_detection": tfknownvalue.StringExact(awstypes.RouteServerPeerLivenessModeBgpKeepalive),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("endpoint_eni_address"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("endpoint_eni_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("route_server_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("route_server_peer_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrSubnetID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVPCID), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "route_server_peer_id"),
				ImportStateVerifyIdentifierAttribute: "route_server_peer_id",
			},
		},
	})
}

func TestAccVPCRouteServerPeer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RouteServerPeer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpc_route_server_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerPeerConfig_basic(rName, rAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerPeerExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCRouteServerPeer, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCRouteServerPeer_bgpOptionsBfd(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RouteServerPeer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpc_route_server_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerPeerConfig_bgpOptionsBfd(rName, rAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerPeerExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("bgp_options"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"peer_asn":                knownvalue.Int64Exact(65000),
							"peer_liveness_detection": tfknownvalue.StringExact(awstypes.RouteServerPeerLivenessModeBfd),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "route_server_peer_id"),
				ImportStateVerifyIdentifierAttribute: "route_server_peer_id",
			},
		},
	})
}
func TestAccVPCRouteServerPeer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RouteServerPeer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpc_route_server_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerPeerConfig_tags1(rName, rAsn, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerPeerExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "route_server_peer_id"),
				ImportStateVerifyIdentifierAttribute: "route_server_peer_id",
			},
			{
				Config: testAccVPCRouteServerPeerConfig_tags2(rName, rAsn, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerPeerExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccVPCRouteServerPeerConfig_tags1(rName, rAsn, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerPeerExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckVPCRouteServerPeerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_route_server_peer" {
				continue
			}

			_, err := tfec2.FindRouteServerPeerByID(ctx, conn, rs.Primary.Attributes["route_server_peer_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Route Server Peer %s still exists", rs.Primary.Attributes["route_server_peer_id"])
		}

		return nil
	}
}

func testAccCheckVPCRouteServerPeerExists(ctx context.Context, n string, v *awstypes.RouteServerPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindRouteServerPeerByID(ctx, conn, rs.Primary.Attributes["route_server_peer_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCRouteServerPeerConfig_base(rName string, rAsn int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_vpc_route_server" "test" {
  amazon_side_asn = %[2]d

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_route_server_vpc_association" "test" {
  route_server_id = aws_vpc_route_server.test.route_server_id
  vpc_id          = aws_vpc.test.id
}

resource "aws_vpc_route_server_endpoint" "test" {
  route_server_id = aws_vpc_route_server_vpc_association.test.route_server_id
  subnet_id       = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, rAsn))
}

func testAccVPCRouteServerPeerConfig_basic(rName string, rAsn int) string {
	return acctest.ConfigCompose(testAccVPCRouteServerPeerConfig_base(rName, rAsn), ` 
resource "aws_vpc_route_server_peer" "test" {
  route_server_endpoint_id = aws_vpc_route_server_endpoint.test.route_server_endpoint_id
  peer_address             = "10.0.1.250"

  bgp_options {
    peer_asn = 65000
  }
}
`)
}

func testAccVPCRouteServerPeerConfig_bgpOptionsBfd(rName string, rAsn int) string {
	return acctest.ConfigCompose(testAccVPCRouteServerPeerConfig_base(rName, rAsn), ` 
resource "aws_vpc_route_server_peer" "test" {
  route_server_endpoint_id = aws_vpc_route_server_endpoint.test.route_server_endpoint_id
  peer_address             = "10.0.1.250"

  bgp_options {
    peer_asn                = 65000
    peer_liveness_detection = "bfd"
  }
}
`)
}

func testAccVPCRouteServerPeerConfig_tags1(rName string, rAsn int, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccVPCRouteServerPeerConfig_base(rName, rAsn), fmt.Sprintf(` 
resource "aws_vpc_route_server_peer" "test" {
  route_server_endpoint_id = aws_vpc_route_server_endpoint.test.route_server_endpoint_id
  peer_address             = "10.0.1.250"

  bgp_options {
    peer_asn = 65000
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value))
}

func testAccVPCRouteServerPeerConfig_tags2(rName string, rAsn int, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccVPCRouteServerPeerConfig_base(rName, rAsn), fmt.Sprintf(` 
resource "aws_vpc_route_server_peer" "test" {
  route_server_endpoint_id = aws_vpc_route_server_endpoint.test.route_server_endpoint_id
  peer_address             = "10.0.1.250"

  bgp_options {
    peer_asn = 65000
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tag2Key, tag2Value))
}
