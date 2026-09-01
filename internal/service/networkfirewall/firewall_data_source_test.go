// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallFirewallDataSource_arn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	dataSourceName := "data.aws_networkfirewall_firewall.test"
	policyResourceName := "aws_networkfirewall_firewall_policy.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallDataSourceConfig_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("firewall/%s", rName)),
					resource.TestCheckResourceAttr(dataSourceName, "delete_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(dataSourceName, "enabled_analysis_types.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_configuration.0.key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_configuration.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_policy_arn", policyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.capacity_usage_summary.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.configuration_sync_state_summary", "IN_SYNC"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.status", "READY"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, names.AttrAvailabilityZone),
					resource.TestMatchTypeSetElemNestedAttrs(dataSourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexache.MustCompile(`vpce-`),
					}),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.sync_states.0.attachment.0.status", "READY"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "subnet_mapping.*.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(dataSourceName, "update_token"),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	dataSourceName := "data.aws_networkfirewall_firewall.test"
	policyResourceName := "aws_networkfirewall_firewall_policy.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallDataSourceConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("firewall/%s", rName)),
					resource.TestCheckResourceAttr(dataSourceName, "delete_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(dataSourceName, "enabled_analysis_types.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_configuration.0.key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_configuration.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_policy_arn", policyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.capacity_usage_summary.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.configuration_sync_state_summary", "IN_SYNC"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.status", "READY"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, names.AttrAvailabilityZone),
					resource.TestMatchTypeSetElemNestedAttrs(dataSourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexache.MustCompile(`vpce-`),
					}),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.sync_states.0.attachment.0.status", "READY"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "subnet_mapping.*.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(dataSourceName, "update_token"),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallDataSource_arnandname(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	dataSourceName := "data.aws_networkfirewall_firewall.test"
	policyResourceName := "aws_networkfirewall_firewall_policy.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallDataSourceConfig_arnAndName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("firewall/%s", rName)),
					resource.TestCheckResourceAttr(dataSourceName, "delete_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(dataSourceName, "enabled_analysis_types.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_configuration.0.key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_configuration.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_policy_arn", policyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.capacity_usage_summary.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.configuration_sync_state_summary", "IN_SYNC"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.status", "READY"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, names.AttrAvailabilityZone),
					resource.TestMatchTypeSetElemNestedAttrs(dataSourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexache.MustCompile(`vpce-`),
					}),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.sync_states.0.attachment.0.status", "READY"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "subnet_mapping.*.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(dataSourceName, "update_token"),
				),
			},
		},
	})
}
func TestAccNetworkFirewallFirewallDataSource_transitGatewayAttachment(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	dataSourceName := "data.aws_networkfirewall_firewall.test"
	dataSourceAvailabilityZones := "data.aws_availability_zones.available"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallDataSourceConfig_transitGatewayAttachment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.status", "READY"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_status.0.transit_gateway_attachment_sync_states.0.transit_gateway_attachment_status", "READY"),
					resource.TestCheckResourceAttr(dataSourceName, "availability_zone_change_protection", acctest.CtFalse),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "availability_zone_mapping.0.availability_zone_id", dataSourceAvailabilityZones, "zone_ids.0"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_status.0.transit_gateway_attachment_sync_states.0.attachment_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "transit_gateway_owner_account_id"),
				),
			},
		},
	})
}

func testAccFirewallDataSourceConfig_baseVPC(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}`, rName))
}

func testAccFirewallDataSourceConfig_arn(rName string) string {
	return acctest.ConfigCompose(testAccFirewallDataSourceConfig_baseVPC(rName), `
data "aws_networkfirewall_firewall" "test" {
  arn = aws_networkfirewall_firewall.test.arn
}
`)
}

func testAccFirewallDataSourceConfig_name(rName string) string {
	return acctest.ConfigCompose(testAccFirewallDataSourceConfig_baseVPC(rName), `
data "aws_networkfirewall_firewall" "test" {
  name = aws_networkfirewall_firewall.test.name
}
`)
}

func testAccFirewallDataSourceConfig_arnAndName(rName string) string {
	return acctest.ConfigCompose(testAccFirewallDataSourceConfig_baseVPC(rName), `
data "aws_networkfirewall_firewall" "test" {
  arn  = aws_networkfirewall_firewall.test.arn
  name = aws_networkfirewall_firewall.test.name
}
`)
}

func testAccFirewallDataSourceConfig_transitGatewayAttachment(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseTGW(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  transit_gateway_id  = aws_ec2_transit_gateway.test.id

  availability_zone_mapping {
    availability_zone_id = data.aws_availability_zones.available.zone_ids[0]
  }
}

data "aws_networkfirewall_firewall" "test" {
  name = aws_networkfirewall_firewall.test.name
}
`, rName))
}
