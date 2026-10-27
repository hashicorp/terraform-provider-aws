// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayAttachmentNATConfigurationDataSource_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_attachment_nat_configuration.test"
	resourceName := "aws_ec2_transit_gateway_attachment_nat_configuration.test"
	attachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// An attachment without a NAT configuration reports disabled support and no rules.
			// This step runs before the resource exists; reading after a destroy would happen at plan time and be stale.
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationDataSourceConfig_noNATConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "nat_gateway_rule.#", "0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrTransitGatewayAttachmentID), attachmentResourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatGatewaySupportStateDisabled)),
				},
			},
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrTransitGatewayAttachmentID), resourceName, tfjsonpath.New(names.AttrTransitGatewayAttachmentID), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("nat_system_route_support"), resourceName, tfjsonpath.New("nat_system_route_support"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatGatewaySupportStateEnabled)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("nat_gateway_rule"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"match_source_cidr": knownvalue.StringExact("10.0.0.0/16"),
							"nat_cidr":          knownvalue.StringExact("100.64.0.0/16"),
							"rule_number":       knownvalue.Int32Exact(100),
							names.AttrState:     tfknownvalue.StringExact(awstypes.TransitGatewayNatRuleStateActivated),
						}),
					})),
				},
			},
		},
	})
}

func testAccTransitGatewayAttachmentNATConfigurationDataSourceConfig_noNATConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayAttachmentNATConfigurationConfig_base(rName),
		`
data "aws_ec2_transit_gateway_attachment_nat_configuration" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}
`)
}

func testAccTransitGatewayAttachmentNATConfigurationDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayAttachmentNATConfigurationConfig_basic(rName),
		`
data "aws_ec2_transit_gateway_attachment_nat_configuration" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_attachment_nat_configuration.test.transit_gateway_attachment_id
}
`)
}
