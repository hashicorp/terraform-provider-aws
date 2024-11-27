// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
)

func TestAccNetworkManagerDXGatewayAttachment_basic(t *testing.T) {
	const (
		resourceName            = "aws_networkmanager_dx_gateway_attachment.test"
		coreNetworkResourceName = "aws_networkmanager_core_network.test"
		dxGatewayResourceName   = "aws_dx_gateway.test"
	)

	t.Parallel()

	testcases := map[string]struct {
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

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckDXGatewayAttachmentDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccDXGWAttachmentConfig_basic(rName, testcase.acceptanceRequired),
						Check: resource.ComposeAggregateTestCheckFunc(
							// testAccCheckVPCAttachmentExists(ctx, resourceName, &v),
							// acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
							resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", "1"),
							resource.TestCheckResourceAttr(resourceName, "attachment_type", "DIRECT_CONNECT_GATEWAY"),
							resource.TestCheckResourceAttrPair(resourceName, "core_network_arn", coreNetworkResourceName, names.AttrARN),
							resource.TestCheckResourceAttrPair(resourceName, "core_network_id", coreNetworkResourceName, names.AttrID),
							// resource.TestCheckResourceAttr(resourceName, "edge_locations.#", acctest.Region()),
							acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
							resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
							resource.TestCheckResourceAttr(resourceName, names.AttrState, string(testcase.expectedState)),
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

func TestAccNetworkManagerDXGatewayAttachment_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	resourceName := "aws_networkmanager_dx_gateway_attachment.test"

	t.Parallel()

	testcases := map[string]struct {
		acceptanceRequired bool
	}{
		"acceptance_required": {
			acceptanceRequired: true,
		},

		"acceptance_not_required": {
			acceptanceRequired: false,
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			// var dxgatewayattachment awstypes.DirectConnectGatewayAttachment
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckDXGatewayAttachmentDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccDXGWAttachmentConfig_basic(rName, testcase.acceptanceRequired),
						Check: resource.ComposeTestCheckFunc(
							// testAccCheckDXGatewayAttachmentExists(ctx, resourceName, &dxgatewayattachment),
							// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
							// but expects a new resource factory function as the third argument. To expose this
							// private function to the testing package, you may need to add a line like the following
							// to exports_test.go:
							//
							//   var ResourceDXGatewayAttachment = newResourceDXGatewayAttachment
							acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceDXGatewayAttachment, resourceName),
						),
						ExpectNonEmptyPlan: true,
					},
				},
			})
		})
	}
}

func testAccCheckDXGatewayAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_dx_gateway_attachment" {
				continue
			}

			input := &networkmanager.GetDirectConnectGatewayAttachmentInput{
				AttachmentId: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetDirectConnectGatewayAttachment(ctx, input)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.NetworkManager, create.ErrActionCheckingDestroyed, tfnetworkmanager.ResNameDXGatewayAttachment, rs.Primary.ID, err)
			}

			return create.Error(names.NetworkManager, create.ErrActionCheckingDestroyed, tfnetworkmanager.ResNameDXGatewayAttachment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDXGatewayAttachmentExists(ctx context.Context, name string, dxgatewayattachment *networkmanager.GetDirectConnectGatewayAttachmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.NetworkManager, create.ErrActionCheckingExistence, tfnetworkmanager.ResNameDXGatewayAttachment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.NetworkManager, create.ErrActionCheckingExistence, tfnetworkmanager.ResNameDXGatewayAttachment, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerClient(ctx)
		resp, err := conn.GetDirectConnectGatewayAttachment(ctx, &networkmanager.GetDirectConnectGatewayAttachmentInput{
			AttachmentId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.NetworkManager, create.ErrActionCheckingExistence, tfnetworkmanager.ResNameDXGatewayAttachment, rs.Primary.ID, err)
		}

		*dxgatewayattachment = *resp

		return nil
	}
}

func testAccDXGWAttachmentConfig_basic(rName string, requireAcceptance bool) string {
	return acctest.ConfigCompose(
		testAccDXGWAttachmentConfig_base(rName, requireAcceptance), `
resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = "arn:aws:directconnect::${data.aws_caller_identity.current.account_id}:dx-gateway/${aws_dx_gateway.test.id}"
  edge_locations = [data.aws_region.current.name]
}
`)
}

func testAccDXGWAttachmentConfig_Accepted_basic(rName string, requireAcceptance bool) string {
	return acctest.ConfigCompose(
		testAccDXGWAttachmentConfig_base(rName, requireAcceptance), `
resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = "arn:aws:directconnect::${data.aws_caller_identity.current.account_id}:dx-gateway/${aws_dx_gateway.test.id}"
  edge_locations = [data.aws_region.current.name]
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id = aws_networkmanager_dx_gateway_attachment.test.id
  attachment_type = aws_networkmanager_dx_gateway_attachment.test.attachment_type
}
`)
}

func testAccDXGWAttachmentConfig_base(rName string, requireAcceptance bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

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
      location = data.aws_region.current.name
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
