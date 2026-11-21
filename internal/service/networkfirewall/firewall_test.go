// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
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

// NAT/Proxy firewall tests allocate Elastic IP addresses, whose default limit is 5 per Region.
// Each group runs its tests serially, so the two groups together hold at most 3 addresses,
// leaving room for TestAccNetworkFirewallFirewallDataSource_natProxy.
func TestAccNetworkFirewallFirewall_NATProxy1_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:                  testAccFirewall_NATProxy_basic,
		"listenerProperties":             testAccFirewall_NATProxy_listenerProperties,
		"capabilityForcesReplacement":    testAccFirewall_NATProxy_capabilityForcesReplacement,
		"allocationIDsWithoutPrivateIPs": testAccFirewall_NATProxy_allocationIDsWithoutPrivateIPs,
		"tlsInspection":                  testAccFirewall_NATProxy_tlsInspection,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func TestAccNetworkFirewallFirewall_NATProxy2_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"updateAvailabilityZones": testAccFirewall_NATProxy_updateAvailabilityZones,
		"privateNATGateway":       testAccFirewall_NATProxy_privateNATGateway,
		"customDomain":            testAccFirewall_NATProxy_customDomain,
		"keepNATGateways":         testAccFirewall_NATProxy_keepNATGateways,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccFirewall_NATProxy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_natProxy(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_associated_nat_gateways", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.dns_name":                   regexache.MustCompile(`\.nfw\.`),
						"attachment.0.endpoint_id":                regexache.MustCompile(`vpce-`),
						"nat_gateway_attachment.0.nat_gateway_id": regexache.MustCompile(`nat-`),
					}),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "nat_gateway_create_mapping.*.subnet_id", "aws_subnet.nat.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "nat_gateway_create_mapping.*.allocation_ids.0", "aws_eip.test.0", "allocation_id"),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.domain_name", ""),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.listener_property.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "proxy_settings.0.listener_property.*", map[string]string{
						names.AttrPort: "3128",
						names.AttrType: string(awstypes.ListenerPropertyTypeHttp),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "proxy_settings.0.listener_property.*", map[string]string{
						names.AttrPort: "8443",
						names.AttrType: string(awstypes.ListenerPropertyTypeHttps),
					}),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", "aws_subnet.firewall.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "transparent_filtering_nat_scope", string(awstypes.TransparentFilteringNatScopeNatAll)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_associated_nat_gateways"},
			},
			// An imported firewall reads its NAT gateway mappings from the NAT gateways, so it isn't replaced.
			// delete_associated_nat_gateways isn't imported, which leaves an in-place update.
			{
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateKind:    resource.ImportBlockWithID,
				ExpectNonEmptyPlan: true,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccFirewallConfig_natProxyTransparentFilteringNATScope(rName, string(awstypes.TransparentFilteringNatScopeNatNone)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "transparent_filtering_nat_scope", string(awstypes.TransparentFilteringNatScopeNatNone)),
				),
			},
		},
	})
}

func testAccFirewall_NATProxy_privateNATGateway(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_natProxyPrivateNATGateway(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.0.allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.0.private_ips.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"nat_gateway_attachment.0.nat_gateway_id": regexache.MustCompile(`nat-`),
					}),
					resource.TestCheckResourceAttr(resourceName, "transparent_filtering_nat_scope", string(awstypes.TransparentFilteringNatScopeNatNone)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_associated_nat_gateways"},
			},
			{
				Config: testAccFirewallConfig_natProxyPrivateNATGateway(rName, 11),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "nat_gateway_create_mapping.*", map[string]*regexp.Regexp{
						"private_ips.0": regexache.MustCompile(`\.11$`),
					}),
				),
			},
		},
	})
}

func testAccFirewall_NATProxy_listenerProperties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_natProxyListenerProperties1(rName, 8080, string(awstypes.ListenerPropertyTypeHttp)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.listener_property.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "proxy_settings.0.listener_property.*", map[string]string{
						names.AttrPort: "8080",
						names.AttrType: string(awstypes.ListenerPropertyTypeHttp),
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_associated_nat_gateways"},
			},
			{
				Config: testAccFirewallConfig_natProxyListenerProperties2(rName, 8081, string(awstypes.ListenerPropertyTypeHttp), 8444, string(awstypes.ListenerPropertyTypeHttps)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.listener_property.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "proxy_settings.0.listener_property.*", map[string]string{
						names.AttrPort: "8081",
						names.AttrType: string(awstypes.ListenerPropertyTypeHttp),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "proxy_settings.0.listener_property.*", map[string]string{
						names.AttrPort: "8444",
						names.AttrType: string(awstypes.ListenerPropertyTypeHttps),
					}),
				),
			},
		},
	})
}

func testAccFirewall_NATProxy_customDomain(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	commonName := acctest.RandomDomain(t)
	domainName := commonName.RandomSubdomain(t).String()
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_natProxyCustomDomain(rName, commonName.String(), domainName, 3128, 8443),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "proxy_settings.0.certificate_arn", "aws_acm_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.domain_name", domainName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]string{
						"attachment.0.dns_name": domainName,
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_associated_nat_gateways"},
			},
			// UpdateProxySettings must repeat the custom hostname and certificate.
			{
				Config: testAccFirewallConfig_natProxyCustomDomain(rName, commonName.String(), domainName, 3129, 8444),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "proxy_settings.0.certificate_arn", "aws_acm_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.0.domain_name", domainName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "proxy_settings.0.listener_property.*", map[string]string{
						names.AttrPort: "3129",
						names.AttrType: string(awstypes.ListenerPropertyTypeHttp),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "proxy_settings.0.listener_property.*", map[string]string{
						names.AttrPort: "8444",
						names.AttrType: string(awstypes.ListenerPropertyTypeHttps),
					}),
				),
			},
		},
	})
}

func testAccFirewall_NATProxy_tlsInspection(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	resourceName := "aws_networkfirewall_firewall.test"
	tlsInspectionConfigurationResourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {
				Source:            "hashicorp/tls",
				VersionConstraint: "4.0.5",
			},
		},
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckFirewallDestroy(ctx, t),
			testAccCheckTLSInspectionConfigurationDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_natProxyTLSInspection(rName, commonName.String(), certificateDomainName, "DECRYPT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_policy_arn", "aws_networkfirewall_firewall_policy.tls", names.AttrARN),
					resource.TestCheckResourceAttrPair("aws_networkfirewall_firewall_policy.tls", "firewall_policy.0.tls_inspection_configuration_arn", tlsInspectionConfigurationResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(tlsInspectionConfigurationResourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.decryption_action", "DECRYPT"),
				),
			},
			// The TLS inspection configuration updates in place while a proxy mode firewall uses it.
			{
				Config: testAccFirewallConfig_natProxyTLSInspection(rName, commonName.String(), certificateDomainName, "BYPASS"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(tlsInspectionConfigurationResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(tlsInspectionConfigurationResourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.decryption_action", "BYPASS"),
				),
			},
		},
	})
}

func testAccFirewall_NATProxy_updateAvailabilityZones(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_natProxyEIPs(rName, 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
				),
			},
			{
				Config:      testAccFirewallConfig_natProxyNATGatewayCreateMappingWithoutSubnetMapping(rName),
				ExpectError: regexache.MustCompile(`can only be added along with the subnet_mapping`),
			},
			{
				Config: testAccFirewallConfig_natProxyEIPs(rName, 2, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "nat_gateway_create_mapping.*.subnet_id", "aws_subnet.nat.1", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", "aws_subnet.firewall.1", names.AttrID),
				),
			},
			{
				Config: testAccFirewallConfig_natProxyEIPs(rName, 1, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					// Removing an Availability Zone detaches its NAT gateway without deleting it.
					testAccCheckFirewallDeleteDetachedNATGateway(ctx, t, "aws_subnet.nat.1"),
				),
			},
		},
	})
}

func testAccFirewall_NATProxy_capabilityForcesReplacement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_natProxyRouted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "proxy_settings.#", "0"),
				),
			},
			{
				Config: testAccFirewallConfig_natProxy(rName, 1),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_create_mapping.#", "1"),
				),
			},
		},
	})
}

// If step 1 fails, the NAT gateway left by the firewall blocks destroying its subnet and must be deleted manually.
func testAccFirewall_NATProxy_keepNATGateways(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_natProxyKeepNATGateways(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_associated_nat_gateways", acctest.CtFalse),
				),
			},
			// Destroying the firewall leaves its NAT gateway running.
			{
				Config: testAccFirewallConfig_baseNATProxy(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallDeleteDetachedNATGateway(ctx, t, "aws_subnet.nat.0"),
				),
			},
		},
	})
}

func testAccFirewall_NATProxy_allocationIDsWithoutPrivateIPs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccFirewallConfig_natProxyAllocationIDsWithoutPrivateIPs(rName),
				ExpectError: regexache.MustCompile(`specify one private_ips address for each of the 1 allocation_ids`),
			},
		},
	})
}

func testAccCheckFirewallDeleteDetachedNATGateway(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		input := ec2.DescribeNatGatewaysInput{
			Filter: []ec2types.Filter{
				{
					Name:   aws.String("subnet-id"),
					Values: []string{rs.Primary.ID},
				},
				{
					Name:   aws.String(names.AttrState),
					Values: enum.Slice(ec2types.NatGatewayStatePending, ec2types.NatGatewayStateAvailable),
				},
			},
		}
		natGateways, err := tfec2.FindNATGateways(ctx, conn, &input)

		if err != nil {
			return err
		}

		if len(natGateways) == 0 {
			return fmt.Errorf("no detached NAT gateway found in subnet (%s)", rs.Primary.ID)
		}

		for _, v := range natGateways {
			input := ec2.DeleteNatGatewayInput{
				NatGatewayId: v.NatGatewayId,
			}
			if _, err := conn.DeleteNatGateway(ctx, &input); err != nil {
				return err
			}
		}

		const (
			timeout = 10 * time.Minute
		)
		for _, v := range natGateways {
			if _, err := tfec2.WaitNATGatewayDeleted(ctx, conn, aws.ToString(v.NatGatewayId), timeout); err != nil {
				return err
			}
		}

		return nil
	}
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

func testAccFirewallConfig_baseNATProxy(rName string, eipCount int) string {
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

resource "aws_subnet" "firewall" {
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

# A public NAT gateway's subnet needs a route to an internet gateway.
resource "aws_route_table" "nat" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "nat" {
  count = 2

  subnet_id      = aws_subnet.nat[count.index].id
  route_table_id = aws_route_table.nat.id
}

resource "aws_eip" "test" {
  count = %[2]d

  domain = "vpc"

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
`, rName, eipCount))
}

func testAccFirewallConfig_natProxy(rName string, azCount int) string {
	return testAccFirewallConfig_natProxyEIPs(rName, azCount, azCount)
}

// testAccFirewallConfig_natProxyEIPs keeps eipCount Elastic IP addresses, which can exceed azCount
// so that an address isn't released while a detached NAT gateway still holds it.
func testAccFirewallConfig_natProxyEIPs(rName string, azCount, eipCount int) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseNATProxy(rName, eipCount), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                           = %[1]q
  firewall_policy_arn            = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                         = aws_vpc.test.id
  delete_associated_nat_gateways = true

  dynamic "subnet_mapping" {
    for_each = slice(aws_subnet.firewall, 0, %[2]d)

    content {
      subnet_id = subnet_mapping.value.id
    }
  }

  dynamic "nat_gateway_create_mapping" {
    for_each = range(%[2]d)

    content {
      subnet_id      = aws_subnet.nat[nat_gateway_create_mapping.value].id
      allocation_ids = [aws_eip.test[nat_gateway_create_mapping.value].allocation_id]
      private_ips    = [cidrhost(aws_subnet.nat[nat_gateway_create_mapping.value].cidr_block, 10)]
    }
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName, azCount))
}

func testAccFirewallConfig_natProxyTransparentFilteringNATScope(rName, scope string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseNATProxy(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                            = %[1]q
  firewall_policy_arn             = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                          = aws_vpc.test.id
  delete_associated_nat_gateways  = true
  transparent_filtering_nat_scope = %[2]q

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat[0].id
    allocation_ids = [aws_eip.test[0].allocation_id]
    private_ips    = [cidrhost(aws_subnet.nat[0].cidr_block, 10)]
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName, scope))
}

func testAccFirewallConfig_natProxyNATGatewayCreateMappingWithoutSubnetMapping(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseNATProxy(rName, 2), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                           = %[1]q
  firewall_policy_arn            = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                         = aws_vpc.test.id
  delete_associated_nat_gateways = true

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }

  dynamic "nat_gateway_create_mapping" {
    for_each = range(2)

    content {
      subnet_id      = aws_subnet.nat[nat_gateway_create_mapping.value].id
      allocation_ids = [aws_eip.test[nat_gateway_create_mapping.value].allocation_id]
      private_ips    = [cidrhost(aws_subnet.nat[nat_gateway_create_mapping.value].cidr_block, 10)]
    }
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName))
}

func testAccFirewallConfig_natProxyKeepNATGateways(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseNATProxy(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat[0].id
    allocation_ids = [aws_eip.test[0].allocation_id]
    private_ips    = [cidrhost(aws_subnet.nat[0].cidr_block, 10)]
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName))
}

func testAccFirewallConfig_natProxyAllocationIDsWithoutPrivateIPs(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseNATProxy(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                           = %[1]q
  firewall_policy_arn            = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                         = aws_vpc.test.id
  delete_associated_nat_gateways = true

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat[0].id
    allocation_ids = [aws_eip.test[0].allocation_id]
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName))
}

func testAccFirewallConfig_natProxyRouted(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseNATProxy(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }
}
`, rName))
}

func testAccFirewallConfig_natProxyPrivateNATGateway(rName string, hostNum int) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseNATProxy(rName, 0), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                            = %[1]q
  firewall_policy_arn             = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                          = aws_vpc.test.id
  delete_associated_nat_gateways  = true
  transparent_filtering_nat_scope = "NAT_NONE"

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }

  nat_gateway_create_mapping {
    subnet_id   = aws_subnet.nat[0].id
    private_ips = [cidrhost(aws_subnet.nat[0].cidr_block, %[2]d)]
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName, hostNum))
}

func testAccFirewallConfig_natProxyListenerProperties1(rName string, port int, listenerType string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseNATProxy(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                           = %[1]q
  firewall_policy_arn            = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                         = aws_vpc.test.id
  delete_associated_nat_gateways = true

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat[0].id
    allocation_ids = [aws_eip.test[0].allocation_id]
    private_ips    = [cidrhost(aws_subnet.nat[0].cidr_block, 10)]
  }

  proxy_settings {
    listener_property {
      port = %[2]d
      type = %[3]q
    }
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName, port, listenerType))
}

func testAccFirewallConfig_natProxyListenerProperties2(rName string, port1 int, listenerType1 string, port2 int, listenerType2 string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseNATProxy(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                           = %[1]q
  firewall_policy_arn            = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                         = aws_vpc.test.id
  delete_associated_nat_gateways = true

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat[0].id
    allocation_ids = [aws_eip.test[0].allocation_id]
    private_ips    = [cidrhost(aws_subnet.nat[0].cidr_block, 10)]
  }

  proxy_settings {
    listener_property {
      port = %[2]d
      type = %[3]q
    }

    listener_property {
      port = %[4]d
      type = %[5]q
    }
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName, port1, listenerType1, port2, listenerType2))
}

func testAccFirewallConfig_natProxyCustomDomain(rName, commonName, domainName string, httpPort, httpsPort int) string {
	return acctest.ConfigCompose(
		testAccFirewallConfig_baseNATProxy(rName, 1),
		testAccTLSInspectionConfigurationConfig_certificateBase(rName, commonName, domainName),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                           = %[1]q
  firewall_policy_arn            = aws_networkfirewall_firewall_policy.test.arn
  vpc_id                         = aws_vpc.test.id
  delete_associated_nat_gateways = true

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat[0].id
    allocation_ids = [aws_eip.test[0].allocation_id]
    private_ips    = [cidrhost(aws_subnet.nat[0].cidr_block, 10)]
  }

  proxy_settings {
    certificate_arn = aws_acm_certificate.test.arn
    domain_name     = %[2]q

    listener_property {
      port = %[3]d
      type = "HTTP"
    }

    listener_property {
      port = %[4]d
      type = "HTTPS"
    }
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName, domainName, httpPort, httpsPort))
}

func testAccFirewallConfig_natProxyTLSInspection(rName, commonName, certificateDomainName, decryptionAction string) string {
	return acctest.ConfigCompose(
		testAccFirewallConfig_baseNATProxy(rName, 1),
		testAccTLSInspectionConfigurationConfig_certificateCheckCertificateRevocationStatus(commonName, certificateDomainName),
		fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name = %[1]q

  tls_inspection_configuration {
    server_certificate_configuration {
      certificate_authority_arn   = aws_acm_certificate.test.arn
      default_decryption_behavior = "BYPASS"

      scope {
        decryption_action = %[2]q
        deployment_mode   = "EXPLICIT_PROXY_ONLY"
        protocols         = [6]

        destination {
          address_definition = "0.0.0.0/0"
        }

        domain_match {
          domains = ["example.com", "*.example.com"]
        }
      }
    }
  }
}

resource "aws_networkfirewall_firewall_policy" "tls" {
  name = "%[1]s-tls"

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    tls_inspection_configuration_arn   = aws_networkfirewall_tls_inspection_configuration.test.arn
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                           = %[1]q
  firewall_policy_arn            = aws_networkfirewall_firewall_policy.tls.arn
  vpc_id                         = aws_vpc.test.id
  delete_associated_nat_gateways = true

  subnet_mapping {
    subnet_id = aws_subnet.firewall[0].id
  }

  nat_gateway_create_mapping {
    subnet_id      = aws_subnet.nat[0].id
    allocation_ids = [aws_eip.test[0].allocation_id]
    private_ips    = [cidrhost(aws_subnet.nat[0].cidr_block, 10)]
  }

  proxy_settings {
    listener_property {
      port = 3128
      type = "HTTP"
    }

    listener_property {
      port = 8443
      type = "HTTPS"
    }
  }

  depends_on = [aws_route_table_association.nat]
}
`, rName, decryptionAction))
}
