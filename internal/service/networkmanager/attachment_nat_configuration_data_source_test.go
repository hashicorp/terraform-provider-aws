// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerAttachmentNATConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_networkmanager_attachment_nat_configuration.test"
	resourceName := "aws_networkmanager_attachment_nat_configuration.test"
	attachmentResourceName := "aws_networkmanager_vpc_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// An attachment without a NAT configuration reports disabled support and no rules.
			// This step runs before the resource exists; reading after a destroy would happen at plan time and be stale.
			{
				Config: testAccAttachmentNATConfigurationDataSourceConfig_noNATConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "nat_gateway_rule.#", "0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("attachment_id"), attachmentResourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.NatGatewaySupportStateDisabled)),
				},
			},
			{
				Config: testAccAttachmentNATConfigurationDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("attachment_id"), resourceName, tfjsonpath.New("attachment_id"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("nat_system_route_support"), resourceName, tfjsonpath.New("nat_system_route_support"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.NatGatewaySupportStateEnabled)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("nat_gateway_rule"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"match_source_cidr": knownvalue.StringExact("10.0.0.0/16"),
							"nat_cidr":          knownvalue.StringExact("100.64.0.0/16"),
							"rule_number":       knownvalue.Int32Exact(100),
							names.AttrState:     tfknownvalue.StringExact(awstypes.NatRuleStateActivated),
						}),
					})),
				},
			},
		},
	})
}

func testAccAttachmentNATConfigurationDataSourceConfig_noNATConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_basic(rName, false),
		`
data "aws_networkmanager_attachment_nat_configuration" "test" {
  attachment_id = aws_networkmanager_vpc_attachment.test.id
}
`)
}

func testAccAttachmentNATConfigurationDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAttachmentNATConfigurationConfig_basic(rName),
		`
data "aws_networkmanager_attachment_nat_configuration" "test" {
  attachment_id = aws_networkmanager_attachment_nat_configuration.test.attachment_id
}
`)
}
