// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFlattenTransitGatewayPeeringAttachmentOptions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		apiObject *awstypes.TransitGatewayPeeringAttachmentOptions
		want      []any
	}{
		"nil": {
			apiObject: nil,
			want:      nil,
		},
		"empty": {
			// The API can return a present-but-empty Options object (for example when a field is in the wire response but not yet modeled in the SDK).
			// This must flatten to no block, otherwise the provider would detect drift against a configuration with no options block and force a replacement.
			apiObject: &awstypes.TransitGatewayPeeringAttachmentOptions{},
			want:      nil,
		},
		"empty dynamic_routing": {
			apiObject: &awstypes.TransitGatewayPeeringAttachmentOptions{
				DynamicRouting: "",
			},
			want: nil,
		},
		"dynamic_routing enable": {
			apiObject: &awstypes.TransitGatewayPeeringAttachmentOptions{
				DynamicRouting: awstypes.DynamicRoutingValueEnable,
			},
			want: []any{map[string]any{
				"dynamic_routing": awstypes.DynamicRoutingValueEnable,
			}},
		},
		"multipath resilience disallowed": {
			// The API returns the multipath resilience settings for every attachment.
			apiObject: &awstypes.TransitGatewayPeeringAttachmentOptions{
				AccepterAllowMultipathResilience:  aws.Bool(false),
				MultipathResilienceStatus:         awstypes.TransitGatewayPeeringAttachmentMultipathResilienceStatusDisabled,
				RequesterAllowMultipathResilience: aws.Bool(false),
			},
			want: []any{map[string]any{
				"allow_multipath_resilience": false,
				"dynamic_routing":            awstypes.DynamicRoutingValue(""),
			}},
		},
		"multipath resilience allowed": {
			apiObject: &awstypes.TransitGatewayPeeringAttachmentOptions{
				AccepterAllowMultipathResilience:  aws.Bool(true),
				AccepterMultipathResilienceMode:   awstypes.TransitGatewayPeeringAttachmentMultipathResilienceModeAlwaysOn,
				MultipathResilienceStatus:         awstypes.TransitGatewayPeeringAttachmentMultipathResilienceStatusEnabled,
				RequesterAllowMultipathResilience: aws.Bool(true),
				RequesterMultipathResilienceMode:  awstypes.TransitGatewayPeeringAttachmentMultipathResilienceModeAlwaysOn,
			},
			want: []any{map[string]any{
				"allow_multipath_resilience": true,
				"dynamic_routing":            awstypes.DynamicRoutingValue(""),
				"multipath_resilience_mode":  awstypes.TransitGatewayPeeringAttachmentMultipathResilienceModeAlwaysOn,
			}},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfec2.FlattenTransitGatewayPeeringAttachmentOptions(testCase.apiObject)

			if diff := cmp.Diff(got, testCase.want); diff != "" {
				t.Errorf("unexpected diff (+want, -got):\n%s", diff)
			}
		})
	}
}

func testAccTransitGatewayPeeringAttachment_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ec2", "transit-gateway-attachment/{id}"),
					resource.TestCheckResourceAttr(resourceName, "multipath_resilience_status", string(awstypes.TransitGatewayPeeringAttachmentMultipathResilienceStatusDisabled)),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.allow_multipath_resilience", acctest.CtFalse),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "peer_account_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_options(t *testing.T, semaphore tfsync.Semaphore) { //nolint:unparam
	acctest.Skip(t, "IncorrectState: You cannot create a dynamic peering attachment")

	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_options_sameAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.dynamic_routing", "enable"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentConfig_options_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_multipathResilience(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	accepterResourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.peer"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               testAccErrorCheckSkipTransitGatewayPeeringAttachmentMultipathResilience(t),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_multipathResilience(rName, acctest.CtTrue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "options.0.allow_multipath_resilience", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "options.0.multipath_resilience_mode", string(awstypes.TransitGatewayPeeringAttachmentMultipathResilienceModeAlwaysOn)),
					// Accepting the attachment also allows multipath resilience for the accepter.
					resource.TestCheckResourceAttr(accepterResourceName, "options.0.allow_multipath_resilience", acctest.CtTrue),
					resource.TestCheckResourceAttr(accepterResourceName, "multipath_resilience_status", string(awstypes.TransitGatewayPeeringAttachmentMultipathResilienceStatusEnabled)),
				),
			},
			{
				// Refresh the requester's status now that the attachment has been accepted.
				Config: testAccTransitGatewayPeeringAttachmentConfig_multipathResilience(rName, acctest.CtTrue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "multipath_resilience_status", string(awstypes.TransitGatewayPeeringAttachmentMultipathResilienceStatusEnabled)),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentConfig_multipathResilience(rName, acctest.CtTrue),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_multipathResilience(rName, acctest.CtFalse),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "options.0.allow_multipath_resilience", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "multipath_resilience_status", string(awstypes.TransitGatewayPeeringAttachmentMultipathResilienceStatusPending)),
				),
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_multipathResilience(rName, acctest.CtTrue),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "options.0.allow_multipath_resilience", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "multipath_resilience_status", string(awstypes.TransitGatewayPeeringAttachmentMultipathResilienceStatusEnabled)),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceTransitGatewayPeeringAttachment(), resourceName),
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

func testAccTransitGatewayPeeringAttachment_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_differentAccount(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, t, resourceName, &transitGatewayPeeringAttachment),
					// Test that the peer account ID != the primary (request) account ID
					func(s *terraform.State) error {
						if acctest.CheckResourceAttrAccountID(ctx, resourceName, "peer_account_id") == nil {
							return fmt.Errorf("peer_account_id attribute incorrectly to the requester's account ID")
						}
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentConfig_differentAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTransitGatewayPeeringAttachmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.TransitGatewayPeeringAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Peering Attachment ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayPeeringAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_peering_attachment" {
				continue
			}

			_, err := tfec2.FindTransitGatewayPeeringAttachmentByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Peering Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayPeeringAttachmentConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "peer" {
  provider = "awsalternate"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentConfig_base(rName),
	)
}

func testAccTransitGatewayPeeringAttachmentConfig_differentAccount_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentConfig_base(rName),
	)
}

func testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
}
`, acctest.AlternateRegion()))
}

func testAccTransitGatewayPeeringAttachmentConfig_options_sameAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  options {
    dynamic_routing = "enable"
  }
}
`, acctest.AlternateRegion()))
}

func testAccTransitGatewayPeeringAttachmentConfig_differentAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_differentAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_account_id         = aws_ec2_transit_gateway.peer.owner_id
  peer_region             = %[2]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.AlternateRegion()))
}

func testAccTransitGatewayPeeringAttachmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, acctest.AlternateRegion(), tagKey1, tagValue1))
}

func testAccTransitGatewayPeeringAttachmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, acctest.AlternateRegion(), tagKey1, tagValue1, tagKey2, tagValue2))
}

// testAccErrorCheckSkipTransitGatewayPeeringAttachmentMultipathResilience skips the test when the Region pair doesn't support multipath resilience.
// A Region without multipath resilience support doesn't recognize the parameter at all; a supported Region rejects an unsupported peer Region.
func testAccErrorCheckSkipTransitGatewayPeeringAttachmentMultipathResilience(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSequence(
		acctest.ErrorCheck(t, names.EC2ServiceID),
		acctest.ErrorCheckSkipMessagesContaining(t,
			"MultipathResilience is not supported between",
			"The parameter Options.AllowMultipathResilience is not recognized",
		),
	)
}

func testAccTransitGatewayPeeringAttachmentConfig_multipathResilience(rName, allow string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[2]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  options {
    allow_multipath_resilience = %[3]s
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_peering_attachment_accepter" "peer" {
  provider = "awsalternate"

  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.AlternateRegion(), allow))
}
