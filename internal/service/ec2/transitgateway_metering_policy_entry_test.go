// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayMeteringPolicyEntry_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayMeteringPolicyEntry
	resourceName := "aws_ec2_transit_gateway_metering_policy_entry.test"
	policyResourceName := "aws_ec2_transit_gateway_metering_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayMeteringPolicyEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMeteringPolicyEntryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayMeteringPolicyEntryExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_metering_policy_id", policyResourceName, "transit_gateway_metering_policy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule_number"), knownvalue.Int64Exact(100)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metered_account"), knownvalue.StringExact(string(awstypes.TransitGatewayMeteringPayerTypeSourceAttachmentOwner))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), knownvalue.StringExact("available")),
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

func testAccTransitGatewayMeteringPolicyEntry_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayMeteringPolicyEntry
	resourceName := "aws_ec2_transit_gateway_metering_policy_entry.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayMeteringPolicyEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMeteringPolicyEntryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayMeteringPolicyEntryExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceTransitGatewayMeteringPolicyEntry, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayMeteringPolicyEntry_fullRule(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayMeteringPolicyEntry
	resourceName := "aws_ec2_transit_gateway_metering_policy_entry.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayMeteringPolicyEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMeteringPolicyEntryConfig_fullRule(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayMeteringPolicyEntryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule_number"), knownvalue.Int64Exact(200)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("metered_account"), knownvalue.StringExact(string(awstypes.TransitGatewayMeteringPayerTypeDestinationAttachmentOwner))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("source_cidr_block"), knownvalue.StringExact("10.0.0.0/8")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("destination_cidr_block"), knownvalue.StringExact("172.16.0.0/12")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("protocol"), knownvalue.StringExact("6")),
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

func testAccCheckTransitGatewayMeteringPolicyEntryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_metering_policy_entry" {
				continue
			}

			policyID, ruleNumber, err := tfec2.TransitGatewayMeteringPolicyEntryParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfec2.FindTransitGatewayMeteringPolicyEntryByTwoPartKey(ctx, conn, policyID, ruleNumber)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Metering Policy Entry %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTransitGatewayMeteringPolicyEntryExists(ctx context.Context, t *testing.T, n string, v *awstypes.TransitGatewayMeteringPolicyEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		policyID, ruleNumber, err := tfec2.TransitGatewayMeteringPolicyEntryParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfec2.FindTransitGatewayMeteringPolicyEntryByTwoPartKey(ctx, conn, policyID, ruleNumber)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTransitGatewayMeteringPolicyEntryConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_metering_policy" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayMeteringPolicyEntryConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayMeteringPolicyEntryConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_metering_policy_entry" "test" {
  transit_gateway_metering_policy_id = aws_ec2_transit_gateway_metering_policy.test.transit_gateway_metering_policy_id
  policy_rule_number                 = 100
  metered_account                    = "source-attachment-owner"
}
`,
	)
}

func testAccTransitGatewayMeteringPolicyEntryConfig_fullRule(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayMeteringPolicyEntryConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_metering_policy_entry" "test" {
  transit_gateway_metering_policy_id = aws_ec2_transit_gateway_metering_policy.test.transit_gateway_metering_policy_id
  policy_rule_number                 = 200
  metered_account                    = "destination-attachment-owner"
  source_cidr_block                  = "10.0.0.0/8"
  destination_cidr_block             = "172.16.0.0/12"
  protocol                           = "6"
}
`,
	)
}
