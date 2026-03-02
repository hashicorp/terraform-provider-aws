// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCEndpoint_gatewayBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "cidr_blocks.#", 0),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrSet(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "service_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Gateway"),
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

func TestAccVPCEndpoint_interfaceBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "service_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
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

func TestAccVPCEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceVPCEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCEndpoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
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
				Config: testAccVPCEndpointConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCEndpointConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_gatewayWithRouteTableAndPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayRouteTableAndPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_gatewayRouteTableAndPolicyModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
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

func TestAccVPCEndpoint_gatewayPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	// This policy checks the DiffSuppressFunc
	policy1 := `
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadOnly",
      "Principal": "*",
      "Action": [
        "dynamodb:DescribeTable",
        "dynamodb:ListTables"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
`

	policy2 := `
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}
`

	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayPolicy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConfig_gatewayPolicy(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccVPCEndpoint_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_orderPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccVPCEndpointConfig_newOrderPolicy(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccVPCEndpoint_ipAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_ipAddressType(rName, "ipv4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
			{
				Config: testAccVPCEndpointConfig_ipAddressType(rName, "dualstack"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "dualstack"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "dualstack"),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_interfaceWithSubnetAndSecurityGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_interfaceSubnetModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// There is a known issue with import verification of computed sets with
					// terraform-plugin-testing > 1.6. The only current workaround is ignoring
					// the impacted attribute.
					// Ref: https://github.com/hashicorp/terraform-plugin-testing/issues/269
					"network_interface_ids",
					"subnet_configuration",
				},
			},
		},
	})
}

func TestAccVPCEndpoint_interfaceNonAWSServiceAcceptOnCreate(t *testing.T) { // nosempgrep:aws-in-func-name
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceNonAWSService(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "available"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccVPCEndpoint_interfaceNonAWSServiceAcceptOnUpdate(t *testing.T) { // nosempgrep:aws-in-func-name
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceNonAWSService(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "pendingAcceptance"),
				),
			},
			{
				Config: testAccVPCEndpointConfig_interfaceNonAWSService(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "available"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccVPCEndpoint_interfaceUserDefinedIPv4(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	ipv4Address1 := "10.0.0.10"
	ipv4Address2 := "10.0.1.10"
	ipv4Address1Updated := "10.0.0.11"
	ipv4Address2Updated := "10.0.1.11"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceUserDefinedIPv4(rName, ipv4Address1, ipv4Address2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "subnet_configuration.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_configuration.*", map[string]string{
						"ipv4": ipv4Address1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_configuration.*", map[string]string{
						"ipv4": ipv4Address2,
					}),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConfig_interfaceUserDefinedIPv4(rName, ipv4Address1Updated, ipv4Address2Updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "subnet_configuration.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_configuration.*", map[string]string{
						"ipv4": ipv4Address1Updated,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_configuration.*", map[string]string{
						"ipv4": ipv4Address2Updated,
					}),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_interfaceUserDefinedIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	ipv6HostNum1 := 10
	ipv6HostNum2 := 11
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceUserDefinedIPv6(rName, ipv6HostNum1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "subnet_configuration.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConfig_interfaceUserDefinedIPv6(rName, ipv6HostNum2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "subnet_configuration.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_interfaceUserDefinedDualstack(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	ipv4Address1 := "10.0.0.10"
	ipv4Address2 := "10.0.1.10"
	ipv4Address1Updated := "10.0.0.11"
	ipv4Address2Updated := "10.0.1.11"
	ipv6HostNum1 := 10
	ipv6HostNum2 := 11
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceUserDefinedDualstackCombined(rName, ipv4Address1, ipv4Address2, ipv6HostNum1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "dualstack"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "dualstack"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "subnet_configuration.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_configuration.*", map[string]string{
						"ipv4": ipv4Address1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_configuration.*", map[string]string{
						"ipv4": ipv4Address2,
					}),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConfig_interfaceUserDefinedDualstackCombined(rName, ipv4Address1Updated, ipv4Address2Updated, ipv6HostNum2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "dualstack"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "dualstack"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "subnet_configuration.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_configuration.*", map[string]string{
						"ipv4": ipv4Address1Updated,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_configuration.*", map[string]string{
						"ipv4": ipv4Address2Updated,
					}),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_VPCEndpointType_gatewayLoadBalancer(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	vpcEndpointServiceResourceName := "aws_vpc_endpoint_service.test"
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckELBv2GatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayLoadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_type", vpcEndpointServiceResourceName, "service_type"),
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

func TestAccVPCEndpoint_crossRegionService(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	var svcCfg awstypes.ServiceConfiguration
	serviceResourceName := "aws_vpc_endpoint_service.test"
	endpointResourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	// record the initialized providers so that we can use them to
	// check for the vpc endpoints in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVPCEndpointDestroy(ctx),
			testAccCheckVPCEndpointServiceDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_crossRegionService(rName, acctest.Region()), // Cross-region
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExistsWithProvider(ctx, endpointResourceName, &endpoint, acctest.RegionProviderFunc(ctx, acctest.AlternateRegion(), &providers)),
					testAccCheckVPCEndpointServiceExists(ctx, serviceResourceName, &svcCfg),
					resource.TestCheckResourceAttr(endpointResourceName, "service_region", acctest.Region()),
				),
			},
			{
				ResourceName:      endpointResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCEndpoint_invalidCrossRegionService(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tfacctest") // 32 character limit

	// record the initialized providers so that we can use them to
	// check for the vpc endpoints in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVPCEndpointDestroy(ctx),
			testAccCheckVPCEndpointServiceDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCEndpointConfig_crossRegionService(rName, acctest.ThirdRegion()), // Cross-region
				ExpectError: regexache.MustCompile(`The Vpc Endpoint Service 'com\.amazonaws\.vpce\.([A-Za-z0-9]+(-[A-Za-z0-9]+)+)\.vpce-svc-[A-Za-z0-9]+' does not exist`),
			},
		},
	})
}

func TestAccVPCEndpoint_resourceConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	resourceConfigurationName := "aws_vpclattice_resource_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_resourceConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Resource"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_configuration_arn", resourceConfigurationName, names.AttrARN),
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

func TestAccVPCEndpoint_resourceConfigurationPrivateDNSOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	resourceConfigurationName := "aws_vpclattice_resource_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Config: testAccVPCEndpointConfig_resourceConfigurationPrivateDNS(rName, "SPECIFIED_DOMAINS_ONLY", []string{"example1.com", "example2.com"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_preference", "SPECIFIED_DOMAINS_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_options.0.private_dns_specified_domains.*", "example1.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_options.0.private_dns_specified_domains.*", "example2.com"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Resource"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_configuration_arn", resourceConfigurationName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Config: testAccVPCEndpointConfig_resourceConfigurationPrivateDNS(rName, "VERIFIED_DOMAINS_AND_SPECIFIED_DOMAINS", []string{"example1.com", "example2.com"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_preference", "VERIFIED_DOMAINS_AND_SPECIFIED_DOMAINS"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_options.0.private_dns_specified_domains.*", "example1.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_options.0.private_dns_specified_domains.*", "example2.com"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Resource"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_configuration_arn", resourceConfigurationName, names.AttrARN),
				),
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Config: testAccVPCEndpointConfig_resourceConfigurationPrivateDNSWithoutPrivateDNSSpecifiedDomains(rName, "ALL_DOMAINS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_preference", "ALL_DOMAINS"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Resource"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_configuration_arn", resourceConfigurationName, names.AttrARN),
				),
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Config: testAccVPCEndpointConfig_resourceConfigurationPrivateDNSWithoutPrivateDNSSpecifiedDomains(rName, "VERIFIED_DOMAINS_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_preference", "VERIFIED_DOMAINS_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Resource"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_configuration_arn", resourceConfigurationName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_serviceNetwork(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	serviceNetworkName := "aws_vpclattice_service_network.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_serviceNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_only_for_inbound_resolver_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // Default SG.
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "ServiceNetwork"),
					resource.TestCheckResourceAttrPair(resourceName, "service_network_arn", serviceNetworkName, names.AttrARN),
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

func TestAccVPCEndpoint_serviceNetworkWithExplicitIPAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_serviceNetworkWithIPAddressType(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					// AWS API returns ip_address_type as uppercase
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "IPV4"),
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

func TestAccVPCEndpoint_serviceNetworkPrivateDNSOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	serviceNetworkName := "aws_vpclattice_service_network.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Config: testAccVPCEndpointConfig_serviceNetworkPrivateDNSOptions(rName, "SPECIFIED_DOMAINS_ONLY", []string{"example1.com", "example2.com"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_preference", "SPECIFIED_DOMAINS_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_options.0.private_dns_specified_domains.*", "example1.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_options.0.private_dns_specified_domains.*", "example2.com"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "ServiceNetwork"),
					resource.TestCheckResourceAttrPair(resourceName, "service_network_arn", serviceNetworkName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Config: testAccVPCEndpointConfig_serviceNetworkPrivateDNSOptions(rName, "VERIFIED_DOMAINS_AND_SPECIFIED_DOMAINS", []string{"example1.com", "example2.com"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_preference", "VERIFIED_DOMAINS_AND_SPECIFIED_DOMAINS"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_options.0.private_dns_specified_domains.*", "example1.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "dns_options.0.private_dns_specified_domains.*", "example2.com"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "ServiceNetwork"),
					resource.TestCheckResourceAttrPair(resourceName, "service_network_arn", serviceNetworkName, names.AttrARN),
				),
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Config: testAccVPCEndpointConfig_serviceNetworkPrivateDNSOptionsWithoutPrivateDNSSpecifiedDomains(rName, "ALL_DOMAINS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_preference", "ALL_DOMAINS"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "ServiceNetwork"),
					resource.TestCheckResourceAttrPair(resourceName, "service_network_arn", serviceNetworkName, names.AttrARN),
				),
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Config: testAccVPCEndpointConfig_serviceNetworkPrivateDNSOptionsWithoutPrivateDNSSpecifiedDomains(rName, "VERIFIED_DOMAINS_ONLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc-endpoint/vpce-.+`)),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_preference", "VERIFIED_DOMAINS_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.private_dns_specified_domains.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "ServiceNetwork"),
					resource.TestCheckResourceAttrPair(resourceName, "service_network_arn", serviceNetworkName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccVPCEndpoint_crossRegionInterfaceAlternateRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_crossRegionInterfacePrivateDNS(rName, acctest.AlternateRegion(), "com.amazonaws."+acctest.AlternateRegion()+".s3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
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

func TestAccVPCEndpoint_crossRegionInterfaceGlobalService(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_crossRegionInterfacePrivateDNS(rName, endpoints.UsEast1RegionID, "com.amazonaws.route53"), // Global service
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "dns_entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_options.0.dns_record_ip_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", acctest.CtTrue),
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

func TestAccVPCEndpoint_gatewayIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_gatewayIPv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"dns_record_ip_type": tfknownvalue.StringExact(awstypes.DnsRecordIpTypeDualstack),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.IpAddressTypeDualstack)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("private_dns_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrServiceName), knownvalue.StringExact(fmt.Sprintf("com.amazonaws.%s.s3", acctest.Region()))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_type"), tfknownvalue.StringExact(awstypes.VpcEndpointTypeGateway)),
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

func TestAccVPCEndpoint_interfaceNoPrivateDNS(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceNoPrivateDNS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"dns_record_ip_type":                             tfknownvalue.StringExact(awstypes.DnsRecordIpTypeIpv4),
						"private_dns_only_for_inbound_resolver_endpoint": knownvalue.Bool(false),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.IpAddressTypeIpv4)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("private_dns_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrServiceName), knownvalue.StringExact(fmt.Sprintf("com.amazonaws.%s.ec2", acctest.Region()))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_type"), tfknownvalue.StringExact(awstypes.VpcEndpointTypeInterface)),
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

func TestAccVPCEndpoint_InterfacePrivateDNSNoGateway_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfacePrivateDNSNoGateway(rName, "s3", awstypes.IpAddressTypeIpv4, awstypes.DnsRecordIpTypeIpv4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"dns_record_ip_type": tfknownvalue.StringExact(awstypes.DnsRecordIpTypeIpv4),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.IpAddressTypeIpv4)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("private_dns_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrServiceName), knownvalue.StringExact(fmt.Sprintf("com.amazonaws.%s.s3", acctest.Region()))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_type"), tfknownvalue.StringExact(awstypes.VpcEndpointTypeInterface)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConfig_interfacePrivateDNSNoGateway(rName, "s3", awstypes.IpAddressTypeDualstack, awstypes.DnsRecordIpTypeIpv6),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"dns_record_ip_type": tfknownvalue.StringExact(awstypes.DnsRecordIpTypeIpv6),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.IpAddressTypeDualstack)),
				},
			},
		},
	})
}

func TestAccVPCEndpoint_InterfacePrivateDNSNoGateway_rds(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfacePrivateDNSNoGateway(rName, "rds", awstypes.IpAddressTypeIpv4, awstypes.DnsRecordIpTypeIpv4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"dns_record_ip_type": tfknownvalue.StringExact(awstypes.DnsRecordIpTypeIpv4),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.IpAddressTypeIpv4)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("private_dns_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrServiceName), knownvalue.StringExact(fmt.Sprintf("com.amazonaws.%s.rds", acctest.Region()))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_type"), tfknownvalue.StringExact(awstypes.VpcEndpointTypeInterface)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConfig_interfacePrivateDNSNoGateway(rName, "rds", awstypes.IpAddressTypeDualstack, awstypes.DnsRecordIpTypeIpv6),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"dns_record_ip_type": tfknownvalue.StringExact(awstypes.DnsRecordIpTypeIpv6),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.IpAddressTypeDualstack)),
				},
			},
		},
	})
}

func TestAccVPCEndpoint_InterfacePrivateDNSNoGateway_s3tables(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfacePrivateDNSNoGateway(rName, "s3tables", awstypes.IpAddressTypeIpv4, awstypes.DnsRecordIpTypeIpv4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"dns_record_ip_type": tfknownvalue.StringExact(awstypes.DnsRecordIpTypeIpv4),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.IpAddressTypeIpv4)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("private_dns_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrServiceName), knownvalue.StringExact(fmt.Sprintf("com.amazonaws.%s.s3tables", acctest.Region()))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_type"), tfknownvalue.StringExact(awstypes.VpcEndpointTypeInterface)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConfig_interfacePrivateDNSNoGateway(rName, "s3tables", awstypes.IpAddressTypeDualstack, awstypes.DnsRecordIpTypeIpv6),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"dns_record_ip_type": tfknownvalue.StringExact(awstypes.DnsRecordIpTypeIpv6),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.IpAddressTypeDualstack)),
				},
			},
		},
	})
}

func TestAccVPCEndpoint_interfaceWithPrivateDNSOnlyForInboundResolverEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var endpoint awstypes.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointConfig_interfaceWithPrivateDNSOnlyForInboundResolverEndpoint(rName, "s3", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"private_dns_only_for_inbound_resolver_endpoint": knownvalue.Bool(true),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddressType), tfknownvalue.StringExact(awstypes.IpAddressTypeIpv4)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("private_dns_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrServiceName), knownvalue.StringExact(fmt.Sprintf("com.amazonaws.%s.s3", acctest.Region()))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_type"), tfknownvalue.StringExact(awstypes.VpcEndpointTypeInterface)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCEndpointConfig_interfaceWithPrivateDNSOnlyForInboundResolverEndpoint(rName, "s3", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointExists(ctx, resourceName, &endpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dns_options"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"private_dns_only_for_inbound_resolver_endpoint": knownvalue.Bool(false),
					})})),
				},
			},
		},
	})
}

func testAccCheckVPCEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_endpoint" {
				continue
			}

			_, err := tfec2.FindVPCEndpointByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckVPCEndpointExists(ctx context.Context, n string, v *awstypes.VpcEndpoint) resource.TestCheckFunc {
	return testAccCheckVPCEndpointExistsWithProvider(ctx, n, v, acctest.DefaultProviderFunc)
}

func testAccCheckVPCEndpointExistsWithProvider(ctx context.Context, n string, v *awstypes.VpcEndpoint, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := providerF().Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVPCEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCEndpointConfig_gatewayBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"
}
`, rName)
}

func testAccVPCEndpointConfig_gatewayRouteTableAndPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  route_table_ids = [
    aws_route_table.test.id,
  ]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}
`, rName)
}

func testAccVPCEndpointConfig_gatewayRouteTableAndPolicyModified(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  route_table_ids = []

  policy = ""

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

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}
`, rName)
}

func testAccVPCEndpointConfig_interfaceBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.region}.ec2"
  vpc_endpoint_type = "Interface"
}
`, rName)
}

func testAccVPCEndpointConfig_ipAddressType(rName, addressType string) string {
	return acctest.ConfigCompose(testAccVPCEndpointServiceConfig_baseSupportedIPAddressTypes(rName), fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
  supported_ip_address_types = ["ipv4", "ipv6"]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = aws_vpc_endpoint_service.test.service_name
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false
  auto_accept         = true
  ip_address_type     = %[2]q

  dns_options {
    dns_record_ip_type = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, addressType))
}

func testAccVPCEndpointConfig_gatewayPolicy(rName, policy string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service      = "dynamodb"
  service_type = "Gateway"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  policy       = <<POLICY
%[2]s
POLICY
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, policy)
}

func testAccVPCEndpointConfig_vpcBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_interfaceSubnet(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.region}.ec2"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false

  subnet_ids = [
    aws_subnet.test[0].id,
  ]

  security_group_ids = [
    aws_security_group.test[0].id,
    aws_security_group.test[1].id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_interfaceSubnetModified(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.region}.ec2"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true

  subnet_ids = [
    aws_subnet.test[2].id,
    aws_subnet.test[1].id,
    aws_subnet.test[0].id,
  ]

  security_group_ids = [
    aws_security_group.test[1].id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_interfaceNonAWSService(rName string, autoAccept bool) string { // nosemgrep:ci.aws-in-func-name
	return acctest.ConfigCompose(
		testAccVPCEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test[0].id,
    aws_subnet.test[1].id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = true

  network_load_balancer_arns = [
    aws_lb.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = aws_vpc_endpoint_service.test.service_name
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false
  auto_accept         = %[2]t

  security_group_ids = [
    aws_security_group.test[0].id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName, autoAccept))
}

func testAccVPCEndpointConfig_interfaceUserDefinedDualstackCombined(rName, ipv4Address1, ipv4Address2 string, ipv6HostNum int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 2), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.region}.athena"
  vpc_endpoint_type = "Interface"
  ip_address_type   = "dualstack"

  subnet_configuration {
    ipv4      = %[2]q
    ipv6      = cidrhost(aws_subnet.test[0].ipv6_cidr_block, %[4]d)
    subnet_id = aws_subnet.test[0].id
  }

  subnet_configuration {
    ipv4      = %[3]q
    ipv6      = cidrhost(aws_subnet.test[1].ipv6_cidr_block, %[4]d)
    subnet_id = aws_subnet.test[1].id
  }

  subnet_ids = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv4Address1, ipv4Address2, ipv6HostNum))
}

func testAccVPCEndpointConfig_interfaceUserDefinedIPv4(rName, ipv4Address1, ipv4Address2 string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.region}.ec2"
  vpc_endpoint_type = "Interface"

  subnet_configuration {
    ipv4      = %[2]q
    subnet_id = aws_subnet.test[0].id
  }

  subnet_configuration {
    ipv4      = %[3]q
    subnet_id = aws_subnet.test[1].id
  }

  subnet_ids = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv4Address1, ipv4Address2))
}

func testAccVPCEndpointConfig_interfaceUserDefinedIPv6(rName string, ipv6HostNum int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id                                         = aws_vpc.test.id
  availability_zone                              = data.aws_availability_zones.available.names[count.index]
  assign_ipv6_address_on_creation                = true
  enable_resource_name_dns_aaaa_record_on_launch = true
  ipv6_cidr_block                                = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  ipv6_native                                    = true
  private_dns_hostname_type_on_launch            = "resource-name"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.region}.athena"
  vpc_endpoint_type = "Interface"
  ip_address_type   = "ipv6"

  subnet_configuration {
    ipv6      = cidrhost(aws_subnet.test[0].ipv6_cidr_block, %[2]d)
    subnet_id = aws_subnet.test[0].id
  }

  subnet_configuration {
    ipv6      = cidrhost(aws_subnet.test[1].ipv6_cidr_block, %[2]d)
    subnet_id = aws_subnet.test[1].id
  }

  subnet_ids = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6HostNum))
}

func testAccVPCEndpointConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCEndpointConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVPCEndpointConfig_gatewayLoadBalancer(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  allowed_principals         = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name      = aws_vpc_endpoint_service.test.service_name
  subnet_ids        = [aws_subnet.test.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.test.service_type
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_orderPolicy(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service      = "dynamodb"
  service_type = "Gateway"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "ReadOnly"
      Principal = "*"
      Action = [
        "dynamodb:DescribeTable",
        "dynamodb:ListTables",
        "dynamodb:ListTagsOfResource",
      ]
      Effect   = "Allow"
      Resource = "*"
    }]
  })
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCEndpointConfig_newOrderPolicy(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service      = "dynamodb"
  service_type = "Gateway"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "ReadOnly"
      Principal = "*"
      Action = [
        "dynamodb:ListTables",
        "dynamodb:ListTagsOfResource",
        "dynamodb:DescribeTable",
      ]
      Effect   = "Allow"
      Resource = "*"
    }]
  })
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCEndpointConfig_crossRegionService(rName, serviceRegion string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccVPCEndpointServiceConfig_crossRegion(rName, acctest.Region(), acctest.AlternateRegion()),
		fmt.Sprintf(`

resource "aws_vpc" "endpoint" {
  provider   = "awsalternate"
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  provider          = "awsalternate"
  vpc_id            = aws_vpc.endpoint.id
  vpc_endpoint_type = "Interface"
  service_name      = aws_vpc_endpoint_service.test.service_name
  service_region    = "%[2]s"

  tags = {
    Name = "%[1]s"
  }
}
`, rName, serviceRegion),
	)
}

func testAccVPCEndpointConfig_resourceConfiguration(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_vpc" "endpoint" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "endpoint" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.endpoint.id
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  resource_configuration_arn = aws_vpclattice_resource_configuration.test.arn
  subnet_ids                 = [aws_subnet.endpoint.id]
  vpc_endpoint_type          = "Resource"
  vpc_id                     = aws_vpc.endpoint.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_resourceConfigurationPrivateDNS(rName, privateDNSPreference string, privateDNSSpecifiedDomains []string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_vpc" "endpoint" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "endpoint" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.endpoint.id
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  resource_configuration_arn = aws_vpclattice_resource_configuration.test.arn
  subnet_ids                 = [aws_subnet.endpoint.id]
  vpc_endpoint_type          = "Resource"
  vpc_id                     = aws_vpc.endpoint.id

  private_dns_enabled = true

  dns_options {
    private_dns_preference        = %[2]q
    private_dns_specified_domains = ["%[3]s"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, privateDNSPreference, strings.Join(privateDNSSpecifiedDomains, `", "`)))
}

func testAccVPCEndpointConfig_resourceConfigurationPrivateDNSWithoutPrivateDNSSpecifiedDomains(rName, privateDNSPreference string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_vpc" "endpoint" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "endpoint" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.endpoint.id
  cidr_block        = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  resource_configuration_arn = aws_vpclattice_resource_configuration.test.arn
  subnet_ids                 = [aws_subnet.endpoint.id]
  vpc_endpoint_type          = "Resource"
  vpc_id                     = aws_vpc.endpoint.id

  private_dns_enabled = true

  dns_options {
    private_dns_preference = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, privateDNSPreference))
}

func testAccVPCEndpointConfig_serviceNetwork(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_network_arn = aws_vpclattice_service_network.test.arn
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "ServiceNetwork"
  vpc_id              = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_serviceNetworkWithIPAddressType(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_network_arn = aws_vpclattice_service_network.test.arn
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "ServiceNetwork"
  vpc_id              = aws_vpc.test.id
  ip_address_type     = "ipv4"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCEndpointConfig_serviceNetworkPrivateDNSOptions(rName, privateDNSPreference string, privateDNSSpecifiedDomains []string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_network_arn = aws_vpclattice_service_network.test.arn
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "ServiceNetwork"
  vpc_id              = aws_vpc.test.id

  private_dns_enabled = true

  dns_options {
    private_dns_preference        = %[2]q
    private_dns_specified_domains = ["%[3]s"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, privateDNSPreference, strings.Join(privateDNSSpecifiedDomains, `", "`)))
}

func testAccVPCEndpointConfig_serviceNetworkPrivateDNSOptionsWithoutPrivateDNSSpecifiedDomains(rName, privateDNSPreference string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_network_arn = aws_vpclattice_service_network.test.arn
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "ServiceNetwork"
  vpc_id              = aws_vpc.test.id

  private_dns_enabled = true

  dns_options {
    private_dns_preference = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, privateDNSPreference))
}

func testAccVPCEndpointConfig_crossRegionInterfacePrivateDNS(rName string, serviceRegion string, serviceName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = %[3]q
  private_dns_enabled = true
  vpc_endpoint_type   = "Interface"
  ip_address_type     = "ipv4"
  service_region      = %[2]q

  tags = {
    Name = %[1]q
  }

}
`, rName, serviceRegion, serviceName)
}

func testAccVPCEndpointConfig_gatewayIPv6(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block                       = "10.0.0.0/16"
  enable_dns_support               = true
  enable_dns_hostnames             = true
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  ip_address_type = "dualstack"
  service_name    = "com.amazonaws.${data.aws_region.current.region}.s3"

  dns_options {
    dns_record_ip_type = "dualstack"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCEndpointConfig_interfaceNoPrivateDNS(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.region}.ec2"
  private_dns_enabled = false
  vpc_endpoint_type   = "Interface"
}
`, rName)
}

func testAccVPCEndpointConfig_interfacePrivateDNSNoGateway(rName, service string, ipAddressType awstypes.IpAddressType, dnsRecordIPType awstypes.DnsRecordIpType) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.0.0.0/16"
  enable_dns_support               = true
  enable_dns_hostnames             = true
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.region}.%[2]s"
  private_dns_enabled = true
  vpc_endpoint_type   = "Interface"
  ip_address_type     = %[3]q

  dns_options {
    dns_record_ip_type = %[4]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, service, ipAddressType, dnsRecordIPType)
}

func testAccVPCEndpointConfig_interfaceWithPrivateDNSOnlyForInboundResolverEndpoint(rName, service string, privateDNSOnlyForInboundResolverEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "gateway" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.%[2]s"

  tags = {
    Name = "%[1]s-gateway"
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.region}.%[2]s"
  vpc_endpoint_type = "Interface"

  private_dns_enabled = true

  dns_options {
    private_dns_only_for_inbound_resolver_endpoint = %[3]t
  }

  tags = {
    Name = "%[1]s-interface"
  }

  depends_on = [aws_vpc_endpoint.gateway]
}
`, rName, service, privateDNSOnlyForInboundResolverEndpoint)
}
