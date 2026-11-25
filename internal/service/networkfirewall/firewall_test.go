// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-cty/cty/msgpack"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallFirewall_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	policyResourceName := "aws_networkfirewall_firewall_policy.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("firewall/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_policy_arn", policyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, names.AttrAvailabilityZone),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexache.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_mapping.*", map[string]string{
						names.AttrIPAddressType: string(awstypes.IPAddressTypeIpv4),
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
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

func TestAccNetworkFirewallFirewall_dualstackSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	policyResourceName := "aws_networkfirewall_firewall_policy.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_dualstackSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("firewall/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_policy_arn", policyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, names.AttrAvailabilityZone),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexache.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_mapping.*", map[string]string{
						names.AttrIPAddressType: string(awstypes.IPAddressTypeDualstack),
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
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

func TestAccNetworkFirewallFirewall_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				Config: testAccFirewallConfig_description(rName, "updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated"),
				),
			},
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
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

func TestAccNetworkFirewallFirewall_deleteProtection(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{

			{
				Config: testAccFirewallConfig_deleteProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", acctest.CtFalse),
				),
			},
			{
				Config: testAccFirewallConfig_deleteProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", acctest.CtTrue),
				),
			},
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", acctest.CtFalse),
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

func TestAccNetworkFirewallFirewall_enabledAnalysisTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_enabledAnalysisTypes(rName, []string{"HTTP_HOST"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled_analysis_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_analysis_types.0", "HTTP_HOST"),
				),
			},
			{
				Config: testAccFirewallConfig_enabledAnalysisTypes(rName, []string{"TLS_SNI"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled_analysis_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_analysis_types.0", "TLS_SNI"),
				),
			},
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled_analysis_types.#", "0"),
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

func TestAccNetworkFirewallFirewall_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_encryptionConfiguration(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "0"),
				),
			},
			{
				Config: testAccFirewallConfig_encryptionConfiguration(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
				),
			},
			{
				Config: testAccFirewallConfig_encryptionConfiguration(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewall_SubnetMappings_updateSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	subnetResourceName := "aws_subnet.test.0"
	updateSubnetResourceName := "aws_subnet.example"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{

			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, names.AttrID),
				),
			},
			{
				Config: testAccFirewallConfig_updateSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", updateSubnetResourceName, names.AttrAvailabilityZone),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexache.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", updateSubnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", updateSubnetResourceName, names.AttrID),
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

func TestAccNetworkFirewallFirewall_SubnetMappings_updateMultipleSubnets(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	subnetResourceName := "aws_subnet.test.0"
	updateSubnetResourceName := "aws_subnet.example"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, names.AttrID),
				),
			},
			{
				Config: testAccFirewallConfig_updateMultipleSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, names.AttrAvailabilityZone),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", updateSubnetResourceName, names.AttrAvailabilityZone),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", updateSubnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", updateSubnetResourceName, names.AttrID),
				),
			},
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, names.AttrAvailabilityZone),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexache.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, names.AttrID),
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

func TestAccNetworkFirewallFirewall_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFirewallConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewall_transitGatewayAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_transitGatewayAttachment(rName, false, 0, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
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

func TestAccNetworkFirewallFirewall_transitGatewayAttachment_updateProtection(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_transitGatewayAttachment(rName, true, 0, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_change_protection", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallConfig_transitGatewayAttachment(rName, false, 0, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_change_protection", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewall_transitGatewayAttachment_updateAvailabilityZone(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_transitGatewayAttachment(rName, false, 0, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_change_protection", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallConfig_transitGatewayAttachment(rName, false, 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_change_protection", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewall_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfnetworkfirewall.ResourceFirewall(), resourceName),
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

func TestAccNetworkFirewallFirewall_proxy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	importStep := resource.TestStep{
		ResourceName:      resourceName,
		ImportState:       true,
		ImportStateVerify: true,
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_proxy(rName, 1, 3128, 8443, 10),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: testAccCheckFirewallProxy(ctx, t, resourceName, 1, 3128, 8443, 10),
			},
			importStep,
			{
				// The only AZ can't be removed and added again, so changing it
				// replaces the firewall.
				Config: testAccFirewallConfig_proxy(rName, 1, 3128, 8443, 11),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: testAccCheckFirewallProxy(ctx, t, resourceName, 1, 3128, 8443, 11),
			},
			importStep,
			{
				// Add an AZ and change the listeners in place.
				Config: testAccFirewallConfig_proxy(rName, 2, 0, 9443, 11),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: testAccCheckFirewallProxy(ctx, t, resourceName, 2, 0, 9443, 11),
			},
			importStep,
			{
				// Change the first AZ's NAT gateway in place, reusing its Elastic
				// IP. The released NAT gateway must be deleted first to free it.
				Config: testAccFirewallConfig_proxy(rName, 2, 0, 9443, 12),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: testAccCheckFirewallProxy(ctx, t, resourceName, 2, 0, 9443, 12),
			},
			importStep,
			{
				// Remove the second AZ in place. Its released NAT gateway is deleted.
				Config: testAccFirewallConfig_proxy(rName, 1, 0, 9443, 12),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: testAccCheckFirewallProxy(ctx, t, resourceName, 1, 0, 9443, 12),
			},
			importStep,
		},
	})
}

func TestAccNetworkFirewallFirewall_proxyAvailabilityZoneValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Create the subnets first, so that the plan can check how the
				// endpoint subnets and NAT gateway mappings pair up by AZ.
				Config: testAccFirewallConfig_baseProxyAvailabilityZoneValidation(rName),
			},
			{
				Config:      testAccFirewallConfig_proxyAvailabilityZoneValidation(rName, []string{"aws_subnet.endpoint[0]", "aws_subnet.endpoint[1]"}, []string{"aws_subnet.nat[0]"}),
				ExpectError: regexache.MustCompile(`"subnet_mapping" in .+ has no "nat_gateway_create_mapping" in the same Availability Zone`),
			},
			{
				Config:      testAccFirewallConfig_proxyAvailabilityZoneValidation(rName, []string{"aws_subnet.endpoint[0]"}, []string{"aws_subnet.nat[0]", "aws_subnet.nat[1]"}),
				ExpectError: regexache.MustCompile(`"nat_gateway_create_mapping" in .+ has no "subnet_mapping" in the same Availability Zone`),
			},
			{
				Config:      testAccFirewallConfig_proxyAvailabilityZoneValidation(rName, []string{"aws_subnet.endpoint[0]"}, []string{"aws_subnet.nat[0]", "aws_subnet.extra"}),
				ExpectError: regexache.MustCompile(`can have only one "nat_gateway_create_mapping" per Availability Zone`),
			},
			{
				Config:      testAccFirewallConfig_proxyAvailabilityZoneValidation(rName, []string{"aws_subnet.endpoint[0]", "aws_subnet.extra"}, []string{"aws_subnet.nat[0]"}),
				ExpectError: regexache.MustCompile(`can have only one "subnet_mapping" per Availability Zone`),
			},
		},
	})
}

// testAccCheckFirewallProxy checks every attribute of a proxy firewall created
// by testAccFirewallConfig_proxy: the subnets and NAT gateway mapping of each
// AZ, the listeners, and the per-AZ status. A port of 0 means no listener.
func testAccCheckFirewallProxy(ctx context.Context, t *testing.T, resourceName string, azCount, httpPort, httpsPort, firstPrivateIPHost int) resource.TestCheckFunc {
	listenerCount := 0
	checks := []resource.TestCheckFunc{
		testAccCheckFirewallExists(ctx, t, resourceName),
		acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "network-firewall", "firewall/{name}"),
		resource.TestCheckResourceAttr(resourceName, "delete_associated_nat_gateways", acctest.CtTrue),
		resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.#", strconv.Itoa(azCount)),
		resource.TestCheckResourceAttr(resourceName, "proxy_settings.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.certificate_arn", ""),
		resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.domain_name", ""),
		resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", strconv.Itoa(azCount)),
		resource.TestCheckResourceAttr(resourceName, "transparent_filtering_nat_scope", string(awstypes.TransparentFilteringNatScopeNatAll)),
		resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
		resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", strconv.Itoa(azCount)),
		resource.TestCheckResourceAttrSet(resourceName, "update_token"),
	}
	for typ, port := range map[awstypes.ListenerPropertyType]int{awstypes.ListenerPropertyTypeHttp: httpPort, awstypes.ListenerPropertyTypeHttps: httpsPort} {
		if port == 0 {
			continue
		}
		listenerCount++
		checks = append(checks, resource.TestCheckTypeSetElemNestedAttrs(resourceName, "proxy_settings.0.listener.*", map[string]string{
			names.AttrPort: strconv.Itoa(port),
			names.AttrType: string(typ),
		}))
	}
	checks = append(checks, resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.listener.#", strconv.Itoa(listenerCount)))

	for i := range azCount {
		host := 10
		if i == 0 {
			host = firstPrivateIPHost
		}
		checks = append(checks,
			resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", fmt.Sprintf("aws_subnet.endpoint.%d", i), names.AttrID),
			testAccCheckFirewallProxyAvailabilityZone(resourceName, i, fmt.Sprintf("10.0.%d.%d", 10+i, host)),
		)
	}

	return resource.ComposeAggregateTestCheckFunc(checks...)
}

// testAccCheckFirewallProxyAvailabilityZone checks the NAT gateway mapping and
// the status of the AZ of aws_subnet.endpoint[index] and aws_subnet.nat[index].
func testAccCheckFirewallProxyAvailabilityZone(resourceName string, index int, privateIP string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attrs := func(name string) (map[string]string, error) {
			rs, ok := s.RootModule().Resources[name]
			if !ok {
				return nil, fmt.Errorf("not found: %s", name)
			}
			return rs.Primary.Attributes, nil
		}

		firewall, err := attrs(resourceName)
		if err != nil {
			return err
		}
		endpointSubnet, err := attrs(fmt.Sprintf("aws_subnet.endpoint.%d", index))
		if err != nil {
			return err
		}
		natSubnet, err := attrs(fmt.Sprintf("aws_subnet.nat.%d", index))
		if err != nil {
			return err
		}
		eip, err := attrs(fmt.Sprintf("aws_eip.test.%d", index))
		if err != nil {
			return err
		}

		mappingFound := false
		for i := range atoi(firewall["nat_gateway_create_mapping.#"]) {
			key := fmt.Sprintf("nat_gateway_create_mapping.%d.", i)
			if firewall[key+names.AttrSubnetID] != natSubnet[names.AttrID] {
				continue
			}
			mappingFound = true
			if got, want := firewall[key+"allocation_ids.#"], "1"; got != want {
				return fmt.Errorf("%s: %sallocation_ids.# = %q, want %q", resourceName, key, got, want)
			}
			if got, want := firewall[key+"allocation_ids.0"], eip["allocation_id"]; got != want {
				return fmt.Errorf("%s: %sallocation_ids.0 = %q, want %q", resourceName, key, got, want)
			}
			if got, want := firewall[key+"private_ips.#"], "1"; got != want {
				return fmt.Errorf("%s: %sprivate_ips.# = %q, want %q", resourceName, key, got, want)
			}
			if got, want := firewall[key+"private_ips.0"], privateIP; got != want {
				return fmt.Errorf("%s: %sprivate_ips.0 = %q, want %q", resourceName, key, got, want)
			}
		}
		if !mappingFound {
			return fmt.Errorf("%s: no nat_gateway_create_mapping for subnet %s", resourceName, natSubnet[names.AttrID])
		}

		for i := range atoi(firewall["firewall_status.0.sync_states.#"]) {
			key := fmt.Sprintf("firewall_status.0.sync_states.%d.", i)
			if firewall[key+"attachment.0.subnet_id"] != endpointSubnet[names.AttrID] {
				continue
			}
			want := map[string]string{
				key + names.AttrAvailabilityZone:        endpointSubnet[names.AttrAvailabilityZone],
				key + "attachment.0.status":             string(awstypes.AttachmentStatusReady),
				key + "nat_gateway_attachment.#":        "1",
				key + "nat_gateway_attachment.0.status": string(awstypes.NatGatewayAttachmentStatusReady),
			}
			for k, v := range want {
				if firewall[k] != v {
					return fmt.Errorf("%s: %s = %q, want %q", resourceName, k, firewall[k], v)
				}
			}
			if firewall[key+"attachment.0.dns_name"] == "" {
				return fmt.Errorf("%s: %sattachment.0.dns_name is empty", resourceName, key)
			}
			if !strings.HasPrefix(firewall[key+"attachment.0.endpoint_id"], "vpce-") {
				return fmt.Errorf("%s: %sattachment.0.endpoint_id = %q, want a VPC endpoint ID", resourceName, key, firewall[key+"attachment.0.endpoint_id"])
			}
			if !strings.HasPrefix(firewall[key+"nat_gateway_attachment.0.nat_gateway_id"], "nat-") {
				return fmt.Errorf("%s: %snat_gateway_attachment.0.nat_gateway_id = %q, want a NAT gateway ID", resourceName, key, firewall[key+"nat_gateway_attachment.0.nat_gateway_id"])
			}
			return nil
		}

		return fmt.Errorf("%s: no firewall_status.0.sync_states entry for subnet %s", resourceName, endpointSubnet[names.AttrID])
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func TestAccNetworkFirewallFirewall_proxyCustomDomain(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	domainName := acctest.RandomDomainName(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	// ACM lists an imported certificate's common name among its subject
	// alternative names, which Network Firewall checks against domain_name.
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_proxyCustomDomain(rName, domainName, certificate, key, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "proxy_settings.0.certificate_arn", "aws_acm_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.listener.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transparent_filtering_nat_scope", string(awstypes.TransparentFilteringNatScopeNatNone)),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]string{
						"attachment.0.dns_name": domainName,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Adding a listener must resend the unchanged certificate and domain name.
				Config: testAccFirewallConfig_proxyCustomDomain(rName, domainName, certificate, key, 3128),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.listener.#", "2"),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewall_proxyValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccFirewallConfig_proxySettingsWithoutNATGateway(rName),
				ExpectError: regexache.MustCompile(`"proxy_settings" requires "nat_gateway_create_mapping"`),
			},
			{
				Config:      testAccFirewallConfig_proxy(rName, 1, 3128, 3128, 10),
				ExpectError: regexache.MustCompile(`"listener" ports must be different`),
			},
		},
	})
}

func testAccCheckFirewallDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_firewall" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

			_, err := tfnetworkfirewall.FindFirewallByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("NetworkFirewall Firewall %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFirewallExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		_, err := tfnetworkfirewall.FindFirewallByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

	input := &networkfirewall.ListFirewallsInput{}

	_, err := conn.ListFirewalls(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFirewallConfig_baseVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName))
}

func testAccFirewallConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName))
}

func testAccFirewallConfig_deleteProtection(rName string, deleteProtection bool) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  delete_protection   = %[1]t
  name                = %[2]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, deleteProtection, rName))
}

func testAccFirewallConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccFirewallConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccFirewallConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  description         = %[2]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName, description))
}

func testAccFirewallConfig_enabledAnalysisTypes(rName string, enabledAnalysisTypes []string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                   = %[1]q
  enabled_analysis_types = ["%[2]s"]
  firewall_policy_arn    = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                 = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName, strings.Join(enabledAnalysisTypes, "\", \"")))
}

func testAccFirewallConfig_updateSubnet(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }

  timeouts {
    update = "1h"
  }
}
`, rName))
}

func testAccFirewallConfig_updateMultipleSubnets(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }

  timeouts {
    update = "1h"
  }
}
`, rName))
}

func testAccFirewallConfig_encryptionConfiguration(rName, description string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id
  description         = %[2]q

  encryption_configuration {
    key_id = aws_kms_key.test.arn
    type   = "CUSTOMER_KMS"
  }

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName, description))
}

func testAccFirewallConfig_dualstackSubnet(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id       = aws_subnet.test[0].id
    ip_address_type = "DUALSTACK"
  }
}
`, rName))
}

func testAccFirewallConfig_baseTGW(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName))
}

func testAccFirewallConfig_transitGatewayAttachment(rName string, changeProtection bool, availabilityZoneStartIndex, availabilityZoneEndIndex int) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseTGW(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                                = %[1]q
  firewall_policy_arn                 = aws_networkfirewall_firewall_policy.test.arn
  transit_gateway_id                  = aws_ec2_transit_gateway.test.id
  availability_zone_change_protection = %[2]t

  dynamic "availability_zone_mapping" {
    for_each = slice(data.aws_availability_zones.available.zone_ids, %[3]d, %[4]d)
    content {
      availability_zone_id = availability_zone_mapping.value
    }
  }
}
`, rName, changeProtection, availabilityZoneStartIndex, availabilityZoneEndIndex))
}

// testAccFirewallConfig_baseProxy creates, per Availability Zone, an endpoint
// subnet, a NAT subnet and an Elastic IP. Two AZs are always created so that
// tests can grow the firewall without destroying resources still in use.
func testAccFirewallConfig_baseProxy(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "endpoint" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "nat" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index + 10)

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  count = 2

  domain = "vpc"

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:forward_to_sfe"]
  }
}
`, rName))
}

// testAccFirewallConfig_proxy creates a proxy firewall in azCount Availability
// Zones. A port of 0 omits that listener. firstPrivateIPHost sets the first
// AZ's NAT gateway private IP, to change that AZ in place.
func testAccFirewallConfig_proxy(rName string, azCount, httpPort, httpsPort, firstPrivateIPHost int) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseProxy(rName), fmt.Sprintf(`
locals {
  http_port  = %[3]d
  https_port = %[4]d
  listeners  = [for l in [{ type = "HTTP", port = local.http_port }, { type = "HTTPS", port = local.https_port }] : l if l.port != 0]
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  dynamic "subnet_mapping" {
    for_each = slice(aws_subnet.endpoint, 0, %[2]d)
    content {
      subnet_id = subnet_mapping.value.id
    }
  }

  dynamic "nat_gateway_create_mapping" {
    for_each = slice(aws_subnet.nat, 0, %[2]d)
    content {
      subnet_id      = nat_gateway_create_mapping.value.id
      allocation_ids = [aws_eip.test[nat_gateway_create_mapping.key].allocation_id]
      private_ips    = [cidrhost(nat_gateway_create_mapping.value.cidr_block, nat_gateway_create_mapping.key == 0 ? %[5]d : 10)]
    }
  }

  proxy_settings {
    dynamic "listener" {
      for_each = local.listeners
      content {
        port = listener.value.port
        type = listener.value.type
      }
    }
  }
}
`, rName, azCount, httpPort, httpsPort, firstPrivateIPHost))
}

// testAccFirewallConfig_baseProxyAvailabilityZoneValidation adds a third subnet,
// in the first AZ, to testAccFirewallConfig_baseProxy.
func testAccFirewallConfig_baseProxyAvailabilityZoneValidation(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseProxy(rName), fmt.Sprintf(`
resource "aws_subnet" "extra" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 20)

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

// testAccFirewallConfig_proxyAvailabilityZoneValidation creates a proxy firewall
// with the given endpoint subnets and NAT gateway subnets, by resource address.
func testAccFirewallConfig_proxyAvailabilityZoneValidation(rName string, endpointSubnets, natSubnets []string) string {
	var blocks strings.Builder
	for _, subnet := range endpointSubnets {
		fmt.Fprintf(&blocks, `
  subnet_mapping {
    subnet_id = %[1]s.id
  }
`, subnet)
	}
	for i, subnet := range natSubnets {
		fmt.Fprintf(&blocks, `
  nat_gateway_create_mapping {
    subnet_id      = %[1]s.id
    allocation_ids = [aws_eip.test[%[2]d].allocation_id]
    private_ips    = [cidrhost(%[1]s.cidr_block, 10)]
  }
`, subnet, i)
	}

	return acctest.ConfigCompose(testAccFirewallConfig_baseProxyAvailabilityZoneValidation(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id
%[2]s
  proxy_settings {
    listener {
      port = 3128
      type = "HTTP"
    }
  }
}
`, rName, blocks.String()))
}

func testAccFirewallConfig_proxySettingsWithoutNATGateway(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  proxy_settings {
    listener {
      port = 3128
      type = "HTTP"
    }
  }
}
`, rName))
}

// testAccFirewallConfig_proxyCustomDomain creates a single-AZ proxy firewall
// with a custom domain and an HTTPS listener. An httpPort of 0 omits the HTTP listener.
func testAccFirewallConfig_proxyCustomDomain(rName, domainName, certificate, key string, httpPort int) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseProxy(rName), fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[3]s"
  private_key      = "%[4]s"
}

locals {
  http_port = %[5]d
}

resource "aws_networkfirewall_firewall" "test" {
  name                            = %[1]q
  firewall_policy_arn             = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                          = aws_vpc.test.id
  transparent_filtering_nat_scope = "NAT_NONE"

  subnet_mapping {
    subnet_id = aws_subnet.endpoint[0].id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat[0].id
    allocation_ids = [aws_eip.test[0].allocation_id]
    private_ips    = [cidrhost(aws_subnet.nat[0].cidr_block, 10)]
  }

  proxy_settings {
    certificate_arn = aws_acm_certificate.test.arn
    domain_name     = %[2]q

    listener {
      port = 8443
      type = "HTTPS"
    }

    dynamic "listener" {
      for_each = local.http_port == 0 ? [] : [local.http_port]
      content {
        port = listener.value
        type = "HTTP"
      }
    }
  }
}
`, rName, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), httpPort))
}

// TestFirewallPlanNATGatewayCreateMappingUnknown checks that planning a new
// proxy firewall whose NAT mapping has values unknown at plan time doesn't add
// an empty element to the set, which would end up in state after create.
func TestFirewallPlanNATGatewayCreateMappingUnknown(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)
	r := tfnetworkfirewall.ResourceFirewall()
	server := schema.NewGRPCProviderServer(&schema.Provider{
		ResourcesMap: map[string]*schema.Resource{"aws_networkfirewall_firewall": r},
	})
	ty := r.CoreConfigSchema().ImpliedType()

	object := func(ty cty.Type, attrs map[string]cty.Value) cty.Value {
		vals := make(map[string]cty.Value)
		for name, at := range ty.AttributeTypes() {
			if v, ok := attrs[name]; ok {
				vals[name] = v
			} else {
				vals[name] = cty.NullVal(at)
			}
		}
		return cty.ObjectVal(vals)
	}
	config := object(ty, map[string]cty.Value{
		names.AttrName:        cty.StringVal("test"),
		"firewall_policy_arn": cty.StringVal("arn:aws:network-firewall:us-west-2:123456789012:firewall-policy/test"),
		names.AttrVPCID:       cty.StringVal("vpc-12345678"),
		"subnet_mapping": cty.SetVal([]cty.Value{object(ty.AttributeType("subnet_mapping").ElementType(), map[string]cty.Value{
			names.AttrSubnetID: cty.StringVal("subnet-12345678"),
		})}),
		"nat_gateway_create_mapping": cty.SetVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
			names.AttrSubnetID: cty.UnknownVal(cty.String),
			"allocation_ids":   cty.ListVal([]cty.Value{cty.UnknownVal(cty.String)}),
			"private_ips":      cty.ListVal([]cty.Value{cty.StringVal("10.0.10.10")}),
		})}),
		"proxy_settings": cty.ListVal([]cty.Value{object(ty.AttributeType("proxy_settings").ElementType(), map[string]cty.Value{
			"listener": cty.SetVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
				names.AttrPort: cty.NumberIntVal(3128),
				names.AttrType: cty.StringVal("HTTP"),
			})}),
		})}),
	})

	encode := func(v cty.Value) *tfprotov5.DynamicValue {
		b, err := msgpack.Marshal(v, ty)
		if err != nil {
			t.Fatal(err)
		}
		return &tfprotov5.DynamicValue{MsgPack: b}
	}

	resp, err := server.PlanResourceChange(ctx, &tfprotov5.PlanResourceChangeRequest{
		TypeName:         "aws_networkfirewall_firewall",
		PriorState:       encode(cty.NullVal(ty)),
		ProposedNewState: encode(config),
		Config:           encode(config),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range resp.Diagnostics {
		t.Errorf("plan diagnostic: %s: %s", d.Summary, d.Detail)
	}

	planned, err := msgpack.Unmarshal(resp.PlannedState.MsgPack, ty)
	if err != nil {
		t.Fatal(err)
	}
	if got := planned.GetAttr("nat_gateway_create_mapping").LengthInt(); got != 1 {
		t.Errorf("planned nat_gateway_create_mapping has %d elements, want 1", got)
	}
}
