// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerDirectConnectGatewayAttachment_basic(t *testing.T) {
	t.Parallel()

	const (
		resourceName            = "aws_networkmanager_dx_gateway_attachment.test"
		coreNetworkResourceName = "aws_networkmanager_core_network.test"
		dxGatewayResourceName   = "aws_dx_gateway.test"
	)
	testCases := map[string]struct {
		acceptanceRequired bool
		expectedState      awstypes.AttachmentState
	}{
		"acceptance_required": {
			acceptanceRequired: true,
			expectedState:      awstypes.AttachmentStatePendingAttachmentAcceptance,
		},
		"acceptance_not_required": {
			acceptanceRequired: false,
			expectedState:      awstypes.AttachmentStateAvailable,
		},
	}

	for name, tc := range testCases { //nolint:paralleltest // false positive
		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
			var dxgatewayattachment awstypes.DirectConnectGatewayAttachment

			acctest.ParallelTest(ctx, t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckDirectConnectGatewayAttachmentDestroy(ctx, t),
				Steps: []resource.TestStep{
					{
						Config: testAccDirectConnectGatewayAttachmentConfig_basic(rName, tc.acceptanceRequired),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckDirectConnectGatewayAttachmentExists(ctx, t, resourceName, &dxgatewayattachment),
							acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
							resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", "1"),
							resource.TestCheckResourceAttr(resourceName, "attachment_type", "DIRECT_CONNECT_GATEWAY"),
							resource.TestCheckResourceAttrPair(resourceName, "core_network_arn", coreNetworkResourceName, names.AttrARN),
							resource.TestCheckResourceAttrPair(resourceName, "core_network_id", coreNetworkResourceName, names.AttrID),
							resource.TestCheckResourceAttr(resourceName, "edge_locations.#", "1"),
							resource.TestCheckResourceAttr(resourceName, "edge_locations.0", acctest.Region()),
							acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
							resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
							resource.TestCheckResourceAttr(resourceName, names.AttrState, string(tc.expectedState)),
							resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
						),
					},
					{
						ResourceName:      resourceName,
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccNetworkManagerDirectConnectGatewayAttachment_disappears(t *testing.T) {
	t.Parallel()

	resourceName := "aws_networkmanager_dx_gateway_attachment.test"
	testCases := map[string]struct {
		acceptanceRequired bool
	}{
		"acceptance_required": {
			acceptanceRequired: true,
		},
		"acceptance_not_required": {
			acceptanceRequired: false,
		},
	}

	for name, tc := range testCases { //nolint:paralleltest // false positive
		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var dxgatewayattachment awstypes.DirectConnectGatewayAttachment
			rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

			acctest.ParallelTest(ctx, t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckDirectConnectGatewayAttachmentDestroy(ctx, t),
				Steps: []resource.TestStep{
					{
						Config: testAccDirectConnectGatewayAttachmentConfig_basic(rName, tc.acceptanceRequired),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckDirectConnectGatewayAttachmentExists(ctx, t, resourceName, &dxgatewayattachment),
							acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkmanager.ResourceDirectConnectGatewayAttachment, resourceName),
						),
						ExpectNonEmptyPlan: true,
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PostApplyPostRefresh: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
							},
						},
					},
				},
			})
		})
	}
}

func TestAccNetworkManagerDirectConnectGatewayAttachment_update(t *testing.T) {
	// Only edge locations can be updated.
	ctx := acctest.Context(t)
	var dxgatewayattachment awstypes.DirectConnectGatewayAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkmanager_dx_gateway_attachment.test"
	coreNetworkResourceName := "aws_networkmanager_core_network.test"
	edgeLocation1 := endpoints.UsEast1RegionID
	edgeLocation2 := endpoints.UsWest2RegionID

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectConnectGatewayAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectConnectGatewayAttachmentConfig_multipleEdgeLocations(rName, edgeLocation1, edgeLocation2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectConnectGatewayAttachmentExists(ctx, t, resourceName, &dxgatewayattachment),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "DIRECT_CONNECT_GATEWAY"),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_arn", coreNetworkResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", coreNetworkResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "edge_locations.#", "1"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.AttachmentStateAvailable)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDirectConnectGatewayAttachmentConfig_multipleEdgeLocationsUpdated(rName, edgeLocation1, edgeLocation2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectConnectGatewayAttachmentExists(ctx, t, resourceName, &dxgatewayattachment),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "DIRECT_CONNECT_GATEWAY"),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_arn", coreNetworkResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", coreNetworkResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "edge_locations.#", "2"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.AttachmentStateAvailable)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerDirectConnectGatewayAttachment_accepted(t *testing.T) {
	ctx := acctest.Context(t)
	var dxgatewayattachment awstypes.DirectConnectGatewayAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkmanager_dx_gateway_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectConnectGatewayAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectConnectGatewayAttachmentConfig_Accepted_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectConnectGatewayAttachmentExists(ctx, t, resourceName, &dxgatewayattachment),
				),
			},
		},
	})
}

func TestAccNetworkManagerDirectConnectGatewayAttachment_routingPolicyLabel(t *testing.T) {
	ctx := acctest.Context(t)
	var dxgatewayattachment awstypes.DirectConnectGatewayAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkmanager_dx_gateway_attachment.test"
	label := "testlabel"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectConnectGatewayAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectConnectGatewayAttachmentConfig_routingPolicyLabel(rName, label),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectConnectGatewayAttachmentExists(ctx, t, resourceName, &dxgatewayattachment),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerDirectConnectGatewayAttachment_routingPolicyLabelUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var dxgatewayattachment awstypes.DirectConnectGatewayAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkmanager_dx_gateway_attachment.test"
	label1 := "testlabel1"
	label2 := "testlabel2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectConnectGatewayAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectConnectGatewayAttachmentConfig_routingPolicyLabel(rName, label1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectConnectGatewayAttachmentExists(ctx, t, resourceName, &dxgatewayattachment),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label1),
				),
			},
			{
				Config: testAccDirectConnectGatewayAttachmentConfig_routingPolicyLabel(rName, label2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectConnectGatewayAttachmentExists(ctx, t, resourceName, &dxgatewayattachment),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", label2),
				),
			},
		},
	})
}

func testAccCheckDirectConnectGatewayAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_dx_gateway_attachment" {
				continue
			}

			_, err := tfnetworkmanager.FindDirectConnectGatewayAttachmentByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Direct Connect Gateway Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDirectConnectGatewayAttachmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.DirectConnectGatewayAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		output, err := tfnetworkmanager.FindDirectConnectGatewayAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDirectConnectGatewayAttachmentConfig_base(rName string, requireAcceptance bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = 65000
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
    require_attachment_acceptance = %[2]t
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }

  attachment_policies {
    rule_number = 1

    conditions {
      type = "any"
    }

    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}
`, rName, requireAcceptance))
}

func testAccDirectConnectGatewayAttachmentConfig_multiRegionBase(rName string, edgeLocation1, edgeLocation2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = 65000
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

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = %[2]q
      asn      = 64512
    }
    edge_locations {
      location = %[3]q
      asn      = 64513
    }
  }

  segments {
    name                          = "shared"
    description                   = "SegmentForSharedServices"
    require_attachment_acceptance = false
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }

  attachment_policies {
    rule_number = 1

    conditions {
      type = "any"
    }

    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}
`, rName, edgeLocation1, edgeLocation2))
}

func testAccDirectConnectGatewayAttachmentConfig_basic(rName string, requireAcceptance bool) string {
	return acctest.ConfigCompose(testAccDirectConnectGatewayAttachmentConfig_base(rName, requireAcceptance), `
resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id            = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = aws_dx_gateway.test.arn
  edge_locations             = [data.aws_region.current.region]
}
`)
}

func testAccDirectConnectGatewayAttachmentConfig_Accepted_basic(rName string) string {
	return acctest.ConfigCompose(testAccDirectConnectGatewayAttachmentConfig_base(rName, true), `
resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id            = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = aws_dx_gateway.test.arn
  edge_locations             = [data.aws_region.current.region]
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_dx_gateway_attachment.test.id
  attachment_type = aws_networkmanager_dx_gateway_attachment.test.attachment_type
}
`)
}

func testAccDirectConnectGatewayAttachmentConfig_multipleEdgeLocations(rName string, edgeLocation1, edgeLocation2 string) string {
	return acctest.ConfigCompose(testAccDirectConnectGatewayAttachmentConfig_multiRegionBase(rName, edgeLocation1, edgeLocation2), fmt.Sprintf(`
resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id            = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = aws_dx_gateway.test.arn
  edge_locations             = [%[1]q]
}
`, edgeLocation1))
}

func testAccDirectConnectGatewayAttachmentConfig_multipleEdgeLocationsUpdated(rName string, edgeLocation1, edgeLocation2 string) string {
	return acctest.ConfigCompose(testAccDirectConnectGatewayAttachmentConfig_multiRegionBase(rName, edgeLocation1, edgeLocation2), fmt.Sprintf(`
resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id            = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = aws_dx_gateway.test.arn
  edge_locations             = [%[1]q, %[2]q]
}
`, edgeLocation1, edgeLocation2))
}

func testAccDirectConnectGatewayAttachmentConfig_baseWithRoutingPolicy(rName, label string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = 65000
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
    require_attachment_acceptance = false
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
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
      value = %[2]q
    }

    action {
      associate_routing_policies = ["policy1"]
    }
  }

  attachment_policies {
    rule_number = 1

    conditions {
      type = "any"
    }

    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}
`, rName, label))
}

func testAccDirectConnectGatewayAttachmentConfig_routingPolicyLabel(rName, label string) string {
	return acctest.ConfigCompose(testAccDirectConnectGatewayAttachmentConfig_baseWithRoutingPolicy(rName, label), fmt.Sprintf(`
resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id            = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = aws_dx_gateway.test.arn
  edge_locations             = [data.aws_region.current.region]
  routing_policy_label       = %[1]q
}
`, label))
}
