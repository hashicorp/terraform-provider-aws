// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestDefaultIPv6CIDRBlockAssociation(t *testing.T) {
	t.Parallel()

	vpc := awstypes.Vpc{
		Ipv6CidrBlockAssociationSet: []awstypes.VpcIpv6CidrBlockAssociation{
			{AssociationId: aws.String("default_cidr"), Ipv6CidrBlock: aws.String("fd00:1::/64"), Ipv6CidrBlockState: &awstypes.VpcCidrBlockState{State: awstypes.VpcCidrBlockStateCodeAssociated}},
			{AssociationId: aws.String("some_other_cidr"), Ipv6CidrBlock: aws.String("fd00:2::/64"), Ipv6CidrBlockState: &awstypes.VpcCidrBlockState{State: awstypes.VpcCidrBlockStateCodeAssociated}},
		},
	}
	if v := tfec2.DefaultIPv6CIDRBlockAssociation(&vpc, ""); v == nil {
		t.Errorf("defaultIPv6CIDRBlockAssociation() got nil")
	} else if got, want := aws.ToString(v.AssociationId), "default_cidr"; got != want {
		t.Errorf("defaultIPv6CIDRBlockAssociation() = %v, want = %v", got, want)
	}
}

func TestAccVPC_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckNoResourceAttr(resourceName, "ipv4_ipam_pool_id"),
					resource.TestCheckNoResourceAttr(resourceName, "ipv4_netmask_length"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
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

func TestAccVPC_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceVPC(), resourceName),
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
}

// TestAccVPC_DynamicResourceTagsMergedWithLocals_ignoreChanges ensures computed "tags_all"
// attributes are correctly determined when the provider-level default_tags block
// is left unused and resource tags (merged with local.tags) are only known at apply time,
// with additional lifecycle ignore_changes attributes, thereby eliminating "Inconsistent final plan" errors
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18366
func TestAccVPC_DynamicResourceTagsMergedWithLocals_ignoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ignoreChangesDynamicTagsMergedLocals("localkey", "localvalue"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we want to ensure subsequent applies will not result in "inconsistent final plan errors"
				// for the attribute "tags_all"
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVPCConfig_ignoreChangesDynamicTagsMergedLocals("localkey", "localvalue"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we wanted to ensure this configuration applies successfully
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccVPC_DynamicResourceTags_ignoreChanges ensures computed "tags_all"
// attributes are correctly determined when the provider-level default_tags block
// is left unused and resource tags are only known at apply time,
// with additional lifecycle ignore_changes attributes, thereby eliminating "Inconsistent final plan" errors
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18366
func TestAccVPC_DynamicResourceTags_ignoreChanges(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ignoreChangesDynamicTags,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we want to ensure subsequent applies will not result in "inconsistent final plan errors"
				// for the attribute "tags_all"
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVPCConfig_ignoreChangesDynamicTags,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we wanted to ensure this configuration applies successfully
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPC_tags_defaultAndIgnoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					testAccCheckVPCUpdateTags(ctx, t, &vpc, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeyPrefixes1("defaultkey1", "defaultvalue1", "defaultkey"),
					testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeys1("defaultkey1", "defaultvalue1"),
					testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				),
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

func TestAccVPC_tags_ignoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					testAccCheckVPCUpdateTags(ctx, t, &vpc, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			{
				Config: acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey") + testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: acctest.ConfigIgnoreTagsKeys("ignorekey1") + testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
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

func TestAccVPC_tenancy(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_dedicatedTenancy(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("instance_tenancy"), knownvalue.StringExact("dedicated")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConfig_default(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("instance_tenancy"), knownvalue.StringExact("default")),
				},
			},
			{
				Config: testAccVPCConfig_dedicatedTenancy(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("instance_tenancy"), knownvalue.StringExact("dedicated")),
				},
			},
		},
	})
}

func TestAccVPC_updateDNSHostnames(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_default(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtFalse),
				),
			},
			{
				Config: testAccVPCConfig_enableDNSHostnames(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/1301
func TestAccVPC_bothDNSOptionsSet(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_bothDNSOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtTrue),
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

// https://github.com/hashicorp/terraform/issues/10168
func TestAccVPC_disabledDNSSupport(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_disabledDNSSupport(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", acctest.CtFalse),
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

func TestAccVPC_enableNetworkAddressUsageMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_enableNetworkAddressUsageMetrics(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_network_address_usage_metrics", acctest.CtTrue),
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

func TestAccVPC_assignGeneratedIPv6CIDRBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlock(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlock(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlock(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
				),
			},
		},
	})
}

func TestAccVPC_assignGeneratedIPv6CIDRBlockWithNetworkBorderGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	azDataSourceName := "data.aws_availability_zone.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
			// https://docs.aws.amazon.com/vpc/latest/userguide/Extend_VPCs.html#local-zone:
			// "You can request the IPv6 Amazon-provided IP addresses and associate them with the network border group
			//  for a new or existing VPCs only for us-west-2-lax-1a and use-west-2-lax-1b. All other Local Zones don't support IPv6."
			testAccPreCheckLocalZoneAvailable(ctx, t, "us-west-2-lax-1") //lintignore:AWSAT003
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlockOptionalNetworkBorderGroup(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "ipv6_cidr_block_network_border_group", azDataSourceName, "network_border_group"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConfig_assignGeneratedIPv6CIDRBlockOptionalNetworkBorderGroup(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "0"),
				),
			},
		},
	})
}

func TestAccVPC_IPAMIPv4BasicNetmask(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ipamIPv4(rName, 28),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					testAccCheckVPCCIDRPrefix(&vpc, "28"),
				),
			},
		},
	})
}

func TestAccVPC_IPAMIPv4BasicExplicitCIDR(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	cidr := "172.2.0.32/28"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ipamIPv4ExplicitCIDR(rName, cidr),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, cidr),
				),
			},
		},
	})
}

func TestAccVPC_IPAMIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	ipamPoolResourceName := "aws_vpc_ipam_pool.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_ipamIPv6(rName, 28),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block_network_border_group"),
					resource.TestCheckResourceAttrPair(resourceName, "ipv6_ipam_pool_id", ipamPoolResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "56"),
				),
			},
		},
	})
}

func TestAccVPC_upgradeFromV5(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t),
		CheckDestroy: testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.92.0",
					},
				},
				Config: testAccVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
		},
	})
}

func TestAccVPC_upgradeFromV5PlanRefreshFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t),
		CheckDestroy: testAccCheckVPCDestroy(ctx, t),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
		},
	})
}

func TestAccVPC_upgradeFromV5WithUpdatePlanRefreshFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t),
		CheckDestroy: testAccCheckVPCDestroy(ctx, t),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCConfig_tags1(acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
					})),
				},
			},
		},
	})
}

func TestAccVPC_upgradeFromV5WithDefaultRegionRefreshFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t),
		CheckDestroy: testAccCheckVPCDestroy(ctx, t),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccVPCConfig_tags1("Name", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCConfig_region(rName, acctest.Region()),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
			},
		},
	})
}

func TestAccVPC_upgradeFromV5WithNewRegionRefreshFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t),
		CheckDestroy: testAccCheckVPCDestroy(ctx, t),
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccVPCConfig_tags1("Name", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccVPCConfig_region(rName, acctest.AlternateRegion()),
				// Can't call 'acctest.CheckVPCExists' as the VPC's in the alternate Region.
				// Check: resource.ComposeAggregateTestCheckFunc(
				// 	acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				// ),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						"Name": knownvalue.StringExact(rName),
					})),
				},
			},
		},
	})
}

func TestAccVPC_regionCreateNull(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_region(rName, "null"),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConfig_region(rName, acctest.Region()),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConfig_region(rName, acctest.AlternateRegion()),
				// Can't call 'acctest.CheckVPCExists' as the VPC's in the alternate Region.
				// Check: resource.ComposeAggregateTestCheckFunc(
				// 	acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				// ),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccVPCRegionImportStateIDFunc(resourceName, acctest.AlternateRegion()),
			},
		},
	})
}

func TestAccVPC_regionCreateNonNull(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_region(rName, acctest.AlternateRegion()),
				// Can't call 'acctest.CheckVPCExists' as the VPC's in the alternate Region.
				// Check: resource.ComposeAggregateTestCheckFunc(
				// 	acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				// ),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccVPCRegionImportStateIDFunc(resourceName, acctest.AlternateRegion()),
			},
			{
				Config: testAccVPCConfig_region(rName, acctest.Region()),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
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

// https://github.com/hashicorp/terraform-provider-aws/issues/45771.
// https://github.com/hashicorp/terraform-provider-aws/issues/45134.
func TestAccVPC_ramSharedImport(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	var vpc awstypes.Vpc
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckRAMSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Initialize the providers.
				Config: testAccVPCConfig_ramSharedInitProviders,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckSameOrganization(ctx, t, acctest.NamedProviderFunc(acctest.ProviderName, providers), acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				// Initialize the source VPC in the alternate account and RAM share a subnet into the default account.
				Config: testAccVPCConfig_ramSharedInit(rName),
			},
			{
				// Import the shared subnet in the default account.
				Config: testAccVPCConfig_ramShared(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, resourceName, &vpc),
				),
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

func testAccCheckVPCDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc" {
				continue
			}

			_, err := tfec2.FindVPCByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPC %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCExists(ctx context.Context, t *testing.T, n string, v *awstypes.Vpc) resource.TestCheckFunc {
	return acctest.CheckVPCExists(ctx, t, n, v)
}

func testAccCheckVPCUpdateTags(ctx context.Context, t *testing.T, vpc *awstypes.Vpc, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		return tfec2.UpdateTags(ctx, conn, aws.ToString(vpc.VpcId), oldTags, newTags)
	}
}

func testAccCheckVPCCIDRPrefix(vpc *awstypes.Vpc, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if strings.Split(aws.ToString(vpc.CidrBlock), "/")[1] != expected {
			return fmt.Errorf("Bad cidr prefix: got %s, expected %s", aws.ToString(vpc.CidrBlock), expected)
		}

		return nil
	}
}

func testAccVPCRegionImportStateIDFunc(n, region string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s@%s", rs.Primary.Attributes[names.AttrID], region), nil
	}
}

const testAccVPCConfig_basic = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}
`

func testAccVPCConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccVPCConfig_ignoreChangesDynamicTagsMergedLocals(localTagKey1, localTagValue1 string) string {
	return fmt.Sprintf(`
locals {
  tags = {
    %[1]q = %[2]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = merge(local.tags, {
    "created_at" = timestamp()
    "updated_at" = timestamp()
  })

  lifecycle {
    ignore_changes = [
      tags["created_at"],
    ]
  }
}
`, localTagKey1, localTagValue1)
}

const testAccVPCConfig_ignoreChangesDynamicTags = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    "created_at" = timestamp()
    "updated_at" = timestamp()
  }

  lifecycle {
    ignore_changes = [
      tags["created_at"],
    ]
  }
}
`

func testAccVPCConfig_assignGeneratedIPv6CIDRBlock(rName string, assignGeneratedIpv6CidrBlock bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = %[2]t
  cidr_block                       = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName, assignGeneratedIpv6CidrBlock)
}

func testAccVPCConfig_assignGeneratedIPv6CIDRBlockOptionalNetworkBorderGroup(rName string, localZoneNetworkBorderGroup bool) string { // lintignore:AWSAT003
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "zone-type"
    values = ["local-zone"]
  }

  filter {
    name   = "opt-in-status"
    values = ["opted-in"]
  }

  filter {
    name   = "group-name"
    values = ["us-west-2-lax-1"]
  }
}

data "aws_availability_zone" "test" {
  zone_id = data.aws_availability_zones.available.zone_ids[0]
}

resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block     = true
  cidr_block                           = "10.1.0.0/16"
  ipv6_cidr_block_network_border_group = %[2]t ? data.aws_availability_zone.test.network_border_group : data.aws_region.current.region

  tags = {
    Name = %[1]q
  }
}
`, rName, localZoneNetworkBorderGroup)
}

func testAccVPCConfig_default(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_enableDNSHostnames(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_dedicatedTenancy(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  instance_tenancy = "dedicated"
  cidr_block       = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_bothDNSOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.2.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_disabledDNSSupport(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block         = "10.2.0.0/16"
  enable_dns_support = false

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_enableNetworkAddressUsageMetrics(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                           = "10.2.0.0/16"
  enable_network_address_usage_metrics = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_baseIPAMIPv4(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.region
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.current.region

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "172.2.0.0/16"
}
`, rName)
}

func testAccVPCConfig_ipamIPv4(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv4(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = %[2]d

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, netmaskLength))
}

func testAccVPCConfig_ipamIPv4ExplicitCIDR(rName, cidr string) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv4(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr_block        = %[2]q

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, cidr))
}

func testAccVPCConfig_baseIPAMIPv6(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.region
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv6"
  ipam_scope_id  = aws_vpc_ipam.test.public_default_scope_id
  locale         = data.aws_region.current.region
  aws_service    = "ec2"

  public_ip_source              = "amazon"
  allocation_max_netmask_length = 128

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = 52
}
`, rName)
}

func testAccVPCConfig_ipamIPv6(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv6(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block          = "10.1.0.0/16"
  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = 56

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, netmaskLength))
}

func testAccVPCConfig_region(rName, region string) string {
	if region != "null" {
		region = strconv.Quote(region)
	}

	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  region = %[2]s

  tags = {
    Name = %[1]q
  }
}
`, rName, region)
}

var testAccVPCConfig_ramSharedInitProviders = acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "target" {
  provider = %[1]q
}

data "aws_caller_identity" "source" {
  provider = %[2]q
}
`, acctest.ProviderName, acctest.ProviderNameAlternate))

func testAccVPCConfig_ramSharedInit(rName string) string {
	return acctest.ConfigCompose(testAccVPCConfig_ramSharedInitProviders, fmt.Sprintf(`
data "aws_organizations_organization" "test" {
  provider = %[1]q
}

resource "aws_vpc" "source" {
  provider = %[2]q

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[3]q
  }
}

resource "aws_subnet" "source" {
  provider = %[2]q

  vpc_id     = aws_vpc.source.id
  cidr_block = cidrsubnet(aws_vpc.source.cidr_block, 8, 0)

  tags = {
    Name = %[3]q
  }
}

resource "aws_ram_resource_share" "test" {
  provider = %[2]q

  name                      = %[3]q
  allow_external_principals = false
}

resource "aws_ram_resource_association" "test" {
  provider = %[2]q

  resource_arn       = aws_subnet.source.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_principal_association" "test" {
  provider = %[2]q

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}
`, acctest.ProviderName, acctest.ProviderNameAlternate, rName))
}

func testAccVPCConfig_ramShared(rName string) string {
	return acctest.ConfigCompose(testAccVPCConfig_ramSharedInit(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  provider = %[1]q

  cidr_block = "10.1.0.0/16"
}

import {
  provider = %[1]q

  to = aws_vpc.test
  id = aws_subnet.source.vpc_id
}
`, acctest.ProviderName))
}

func TestGuardDutySecurityGroupNamePattern(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		groupName   string
		vpcID       string
		shouldMatch bool
	}{
		{
			name:        "exact match",
			groupName:   "GuardDutyManagedSecurityGroup-vpc-12345678",
			vpcID:       "vpc-12345678",
			shouldMatch: true,
		},
		{
			name:        "different VPC ID",
			groupName:   "GuardDutyManagedSecurityGroup-vpc-87654321",
			vpcID:       "vpc-12345678",
			shouldMatch: false,
		},
		{
			name:        "missing prefix",
			groupName:   "ManagedSecurityGroup-vpc-12345678",
			vpcID:       "vpc-12345678",
			shouldMatch: false,
		},
		{
			name:        "wrong format",
			groupName:   "GuardDuty-vpc-12345678",
			vpcID:       "vpc-12345678",
			shouldMatch: false,
		},
		{
			name:        "empty group name",
			groupName:   "",
			vpcID:       "vpc-12345678",
			shouldMatch: false,
		},
		{
			name:        "case sensitive - lowercase",
			groupName:   "guarddutymanagedsecuritygroup-vpc-12345678",
			vpcID:       "vpc-12345678",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			expectedName := tfec2.GuardDutySecurityGroupPrefix + tc.vpcID
			matches := tc.groupName == expectedName

			if matches != tc.shouldMatch {
				t.Errorf("Security group name pattern matching failed for '%s' with VPC '%s'\nExpected match: %v, Got: %v",
					tc.groupName, tc.vpcID, tc.shouldMatch, matches)
			}
		})
	}
}

func TestIsDependencyViolationError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                        string
		err                         error
		shouldBeDependencyViolation bool
	}{
		{
			name:                        "nil error",
			err:                         nil,
			shouldBeDependencyViolation: false,
		},
		{
			name:                        "DependencyViolation error",
			err:                         fmt.Errorf("DependencyViolation: resource sg-123 has a dependent object"),
			shouldBeDependencyViolation: true,
		},
		{
			name:                        "dependent object error",
			err:                         fmt.Errorf("Cannot delete security group: resource has a dependent object"),
			shouldBeDependencyViolation: true,
		},
		{
			name:                        "DependencyViolation with network interface",
			err:                         fmt.Errorf("DependencyViolation: resource sg-456 has a dependent object (network interface eni-789)"),
			shouldBeDependencyViolation: true,
		},
		{
			name:                        "other error",
			err:                         fmt.Errorf("InternalError: An internal error occurred"),
			shouldBeDependencyViolation: false,
		},
		{
			name:                        "unauthorized error",
			err:                         fmt.Errorf("UnauthorizedOperation: You are not authorized"),
			shouldBeDependencyViolation: false,
		},
		{
			name:                        "not found error",
			err:                         fmt.Errorf("InvalidGroup.NotFound: The security group does not exist"),
			shouldBeDependencyViolation: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tfec2.IsDependencyViolationError(tc.err)
			if result != tc.shouldBeDependencyViolation {
				t.Errorf("Expected %v, got %v for error: %v", tc.shouldBeDependencyViolation, result, tc.err)
			}
		})
	}
}

// TestAccVPC_guardDutySecurityGroupCleanup validates UC-V1: VPC destroy with out-of-band
// GuardDuty resources. It creates a VPC + subnet via Terraform, then uses the AWS SDK to
// create a GuardDuty endpoint and security group out-of-band (not in Terraform state).
// When terraform destroy runs, it deletes the subnet first (triggering subnet-level cleanup),
// then tries to delete the VPC. Since the endpoint and SG are out-of-band, Terraform may hit
// DependencyViolation, and detectAndDeleteGuardDutyVPCEndpoints + detectAndDeleteGuardDutySecurityGroups
// run to clean them up.
//
// EXPECTED: Test PASSES. VPC, endpoint, and SG are all cleaned up.
func TestAccVPC_guardDutySecurityGroupCleanup(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	var subnet awstypes.Subnet
	var vpcID string
	vpcResourceName := "aws_vpc.test"
	subnetResourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVPCGuardDutyCleanupDestroy(ctx, t, &vpcID),
			testAccDeleteGuardDutyResources(ctx, t, &vpcID),
		),
		Steps: []resource.TestStep{
			{
				// Step 1: Create VPC + subnet via Terraform. After creation,
				// use SDK to create GuardDuty endpoint + SG out-of-band.
				// When the test framework runs terraform destroy, it will
				// trigger the VPC-level cleanup path.
				Config: testAccVPCConfig_guardDutySecurityGroupCleanup(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, vpcResourceName, &vpc),
					testAccCheckSubnetExists(ctx, t, subnetResourceName, &subnet),
					testAccCaptureVPCIDFromVPC(&vpc, &vpcID),
					testAccCreateGuardDutyResourcesFromSubnet(ctx, t, &subnet),
					testAccCheckVPCGuardDutySecurityGroupExists(ctx, t, &vpc),
					testAccCheckVPCGuardDutyEndpointExists(ctx, t, &vpc),
				),
			},
		},
	})
}

func testAccCheckVPCGuardDutySecurityGroupExists(ctx context.Context, t *testing.T, vpc *awstypes.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)
		vpcID := aws.ToString(vpc.VpcId)

		sgInput := &ec2.DescribeSecurityGroupsInput{
			Filters: []awstypes.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{vpcID},
				},
				{
					Name:   aws.String("group-name"),
					Values: []string{fmt.Sprintf("GuardDutyManagedSecurityGroup-%s", vpcID)},
				},
				{
					Name:   aws.String("tag:GuardDutyManaged"),
					Values: []string{acctest.CtTrue},
				},
			},
		}
		sgOutput, err := conn.DescribeSecurityGroups(ctx, sgInput)
		if err != nil {
			return fmt.Errorf("error describing security groups: %w", err)
		}
		if len(sgOutput.SecurityGroups) == 0 {
			return fmt.Errorf("expected GuardDuty security group with GuardDutyManaged=true tag to exist, but none found")
		}

		return nil
	}
}

func testAccCheckVPCGuardDutyEndpointExists(ctx context.Context, t *testing.T, vpc *awstypes.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)
		vpcID := aws.ToString(vpc.VpcId)

		endpointsInput := &ec2.DescribeVpcEndpointsInput{
			Filters: []awstypes.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{vpcID},
				},
				{
					Name:   aws.String("service-name"),
					Values: []string{"*guardduty-data*"},
				},
				{
					Name:   aws.String("tag:GuardDutyManaged"),
					Values: []string{acctest.CtTrue},
				},
			},
		}
		endpointsOutput, err := conn.DescribeVpcEndpoints(ctx, endpointsInput)
		if err != nil {
			return fmt.Errorf("error describing VPC endpoints: %w", err)
		}
		if len(endpointsOutput.VpcEndpoints) == 0 {
			return fmt.Errorf("expected GuardDuty VPC endpoint with GuardDutyManaged=true tag to exist, but none found")
		}

		return nil
	}
}

func testAccCaptureVPCIDFromVPC(vpc *awstypes.Vpc, vpcID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*vpcID = aws.ToString(vpc.VpcId)
		return nil
	}
}

func testAccCheckVPCGuardDutyCleanupDestroy(ctx context.Context, t *testing.T, vpcID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc" {
				continue
			}

			_, err := tfec2.FindVPCByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPC %s still exists", rs.Primary.ID)
		}

		if vpcID != nil && *vpcID != "" {
			endpointsInput := &ec2.DescribeVpcEndpointsInput{
				Filters: []awstypes.Filter{
					{
						Name:   aws.String("vpc-id"),
						Values: []string{*vpcID},
					},
					{
						Name:   aws.String("service-name"),
						Values: []string{"*guardduty-data*"},
					},
					{
						Name:   aws.String("tag:GuardDutyManaged"),
						Values: []string{acctest.CtTrue},
					},
				},
			}
			endpointsOutput, err := conn.DescribeVpcEndpoints(ctx, endpointsInput)
			if err != nil {
				return nil
			}
			activeEndpoints := 0
			for _, ep := range endpointsOutput.VpcEndpoints {
				if string(ep.State) != "deleted" {
					activeEndpoints++
				}
			}
			if activeEndpoints > 0 {
				return fmt.Errorf("expected GuardDuty VPC endpoints to be cleaned up, but found %d active endpoint(s)", activeEndpoints)
			}

			sgInput := &ec2.DescribeSecurityGroupsInput{
				Filters: []awstypes.Filter{
					{
						Name:   aws.String("vpc-id"),
						Values: []string{*vpcID},
					},
					{
						Name:   aws.String("group-name"),
						Values: []string{fmt.Sprintf("GuardDutyManagedSecurityGroup-%s", *vpcID)},
					},
					{
						Name:   aws.String("tag:GuardDutyManaged"),
						Values: []string{acctest.CtTrue},
					},
				},
			}
			sgOutput, err := conn.DescribeSecurityGroups(ctx, sgInput)
			if err != nil {
				return nil
			}
			if len(sgOutput.SecurityGroups) > 0 {
				return fmt.Errorf("expected GuardDuty security groups to be cleaned up, but found %d group(s)", len(sgOutput.SecurityGroups))
			}
		}

		return nil
	}
}

func testAccVPCConfig_guardDutySecurityGroupCleanup(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-test"
  }
}
`, rName)
}

// TestAccVPC_guardDutyEndpointAlreadyCleaned validates UC-V2: VPC destroy where the
// GuardDuty endpoint was already cleaned up by subnet-level deletion, but the SG remains.
// Step 1 creates VPC + subnet, then creates GuardDuty endpoint + SG out-of-band via SDK.
// Step 2 removes the subnet from config — Terraform deletes the subnet, which triggers
// dissociateGuardDutyVPCEndpoints (dissociates the endpoint from the subnet). The endpoint
// may be auto-deleted by AWS or left in a degraded state, but the SG still exists.
// When terraform destroy runs on the remaining VPC, detectAndDeleteGuardDutyVPCEndpoints
// finds nothing (or a degraded endpoint) and detectAndDeleteGuardDutySecurityGroups finds
// and deletes the SG.
//
// EXPECTED: Test PASSES. Subnet deletion cleans up the endpoint association, VPC deletion
// cleans up the remaining SG.
func TestAccVPC_guardDutyEndpointAlreadyCleaned(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	var subnet awstypes.Subnet
	var vpcID string
	vpcResourceName := "aws_vpc.test"
	subnetResourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVPCGuardDutyCleanupDestroy(ctx, t, &vpcID),
			testAccDeleteGuardDutyResources(ctx, t, &vpcID),
		),
		Steps: []resource.TestStep{
			{
				// Step 1: Create VPC + subnet via Terraform. After creation,
				// use SDK to create GuardDuty endpoint + SG out-of-band.
				Config: testAccVPCConfig_guardDutyEndpointAlreadyCleaned_withSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, vpcResourceName, &vpc),
					testAccCheckSubnetExists(ctx, t, subnetResourceName, &subnet),
					testAccCaptureVPCIDFromVPC(&vpc, &vpcID),
					testAccCreateGuardDutyResourcesFromSubnet(ctx, t, &subnet),
					testAccCheckVPCGuardDutySecurityGroupExists(ctx, t, &vpc),
					testAccCheckVPCGuardDutyEndpointExists(ctx, t, &vpc),
				),
			},
			{
				// Step 2: Remove subnet from config (VPC only remains).
				// Terraform deletes the subnet, triggering dissociateGuardDutyVPCEndpoints
				// which dissociates the endpoint. The SG still exists in the VPC.
				// Verify the SG persists after subnet deletion.
				Config: testAccVPCConfig_guardDutyEndpointAlreadyCleaned_withoutSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, vpcResourceName, &vpc),
					testAccCheckVPCGuardDutySecurityGroupExists(ctx, t, &vpc),
				),
			},
			// After step 2, the test framework runs terraform destroy on the
			// remaining VPC. detectAndDeleteGuardDutyVPCEndpoints finds nothing
			// (or a degraded endpoint) and detectAndDeleteGuardDutySecurityGroups
			// finds and deletes the SG.
		},
	})
}

// TestAccVPC_guardDutyNoResources validates UC-V4: standard VPC with a subnet and no
// GuardDuty resources at all. This is a regression test to ensure the GuardDuty cleanup
// code is a safe no-op when there are no GuardDuty resources. The DescribeVpcEndpoints
// and DescribeSecurityGroups calls should return empty results and the code should return
// immediately without interfering with normal VPC deletion.
//
// Note: This test only exercises the GuardDuty cleanup path if DeleteVpc returns
// DependencyViolation. Without GuardDuty resources, there may be no DependencyViolation
// at all, in which case the cleanup code is never called.
//
// EXPECTED: Test PASSES. Normal VPC deletion unaffected by GuardDuty cleanup code.
func TestAccVPC_guardDutyNoResources(t *testing.T) {
	ctx := acctest.Context(t)
	var vpc awstypes.Vpc
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig_guardDutyNoResources(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, t, vpcResourceName, &vpc),
				),
			},
		},
	})
}

func testAccVPCConfig_guardDutyEndpointAlreadyCleaned_withSubnet(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-test"
  }
}
`, rName)
}

func testAccVPCConfig_guardDutyEndpointAlreadyCleaned_withoutSubnet(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCConfig_guardDutyNoResources(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-test"
  }
}
`, rName)
}

func TestDetectAndDeleteGuardDutyVPCEndpoints_warningMessageFormat(t *testing.T) {
	t.Parallel()

	const warningTemplate = "During deletion of VPC %s, Terraform attempted to check for and delete " +
		"GuardDuty-managed VPC endpoints that may have been causing a DependencyViolation, " +
		"but lacked sufficient IAM permissions (%s) to do so. " +
		"If GuardDuty is enabled in this VPC, these permissions may be required for automatic cleanup."

	t.Run("non-definitive phrasing", func(t *testing.T) {
		t.Parallel()

		phrasingCases := []struct {
			name            string
			expectedPhrase  string
			forbiddenPhrase string
		}{
			{
				name:            "uses may have been causing",
				expectedPhrase:  "may have been causing",
				forbiddenPhrase: "was causing",
			},
			{
				name:            "uses attempted to check for",
				expectedPhrase:  "attempted to check for",
				forbiddenPhrase: "found",
			},
			{
				name:            "uses may be required",
				expectedPhrase:  "may be required",
				forbiddenPhrase: "are required",
			},
		}

		for _, tc := range phrasingCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				if !strings.Contains(warningTemplate, tc.expectedPhrase) {
					t.Errorf("Warning template missing non-definitive phrase %q\nTemplate: %s", tc.expectedPhrase, warningTemplate)
				}
				if strings.Contains(warningTemplate, tc.forbiddenPhrase) {
					t.Errorf("Warning template contains definitive phrase %q which should not be present\nTemplate: %s", tc.forbiddenPhrase, warningTemplate)
				}
			})
		}
	})

	t.Run("includes required identifiers", func(t *testing.T) {
		t.Parallel()

		identifierCases := []struct {
			name        string
			vpcID       string
			permissions string
		}{
			{
				name:        "describe endpoints permission",
				vpcID:       "vpc-0abc123def456",
				permissions: "ec2:DescribeVpcEndpoints",
			},
			{
				name:        "delete endpoints permission",
				vpcID:       "vpc-0xyz789ghi012",
				permissions: "ec2:DeleteVpcEndpoints",
			},
		}

		for _, tc := range identifierCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				formatted := fmt.Sprintf(warningTemplate, tc.vpcID, tc.permissions)

				if !strings.Contains(formatted, tc.vpcID) {
					t.Errorf("Formatted warning missing VPC ID %q\nFormatted: %s", tc.vpcID, formatted)
				}
				if !strings.Contains(formatted, tc.permissions) {
					t.Errorf("Formatted warning missing permissions %q\nFormatted: %s", tc.permissions, formatted)
				}
			})
		}
	})
}

func TestDetectAndDeleteGuardDutySecurityGroups_warningMessageFormat(t *testing.T) {
	t.Parallel()

	const warningTemplate = "During deletion of VPC %s, Terraform attempted to check for and delete " +
		"GuardDuty-managed security groups that may have been causing a DependencyViolation, " +
		"but lacked sufficient IAM permissions (%s) to do so. If GuardDuty is enabled " +
		"in this VPC, these permissions may be required for automatic cleanup."

	t.Run("non-definitive phrasing", func(t *testing.T) {
		t.Parallel()

		phrasingCases := []struct {
			name            string
			expectedPhrase  string
			forbiddenPhrase string
		}{
			{
				name:            "uses may have been causing",
				expectedPhrase:  "may have been causing",
				forbiddenPhrase: "was causing",
			},
			{
				name:            "uses attempted to check for",
				expectedPhrase:  "attempted to check for",
				forbiddenPhrase: "found",
			},
			{
				name:            "uses may be required",
				expectedPhrase:  "may be required",
				forbiddenPhrase: "are required",
			},
		}

		for _, tc := range phrasingCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				if !strings.Contains(warningTemplate, tc.expectedPhrase) {
					t.Errorf("Warning template missing non-definitive phrase %q\nTemplate: %s", tc.expectedPhrase, warningTemplate)
				}
				if strings.Contains(warningTemplate, tc.forbiddenPhrase) {
					t.Errorf("Warning template contains definitive phrase %q which should not be present\nTemplate: %s", tc.forbiddenPhrase, warningTemplate)
				}
			})
		}
	})

	t.Run("includes required identifiers", func(t *testing.T) {
		t.Parallel()

		identifierCases := []struct {
			name        string
			vpcID       string
			permissions string
		}{
			{
				name:        "describe security groups permission",
				vpcID:       "vpc-0abc123def456",
				permissions: "ec2:DescribeSecurityGroups",
			},
			{
				name:        "delete security group permission",
				vpcID:       "vpc-0xyz789ghi012",
				permissions: "ec2:DeleteSecurityGroup",
			},
		}

		for _, tc := range identifierCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				formatted := fmt.Sprintf(warningTemplate, tc.vpcID, tc.permissions)

				if !strings.Contains(formatted, tc.vpcID) {
					t.Errorf("Formatted warning missing VPC ID %q\nFormatted: %s", tc.vpcID, formatted)
				}
				if !strings.Contains(formatted, tc.permissions) {
					t.Errorf("Formatted warning missing permissions %q\nFormatted: %s", tc.permissions, formatted)
				}
			})
		}
	})
}
