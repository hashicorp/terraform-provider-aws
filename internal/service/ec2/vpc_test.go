package ec2_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPC_basic(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc/vpc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "default_network_acl_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "dhcp_options_id"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink_dns_support", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					resource.TestCheckNoResourceAttr(resourceName, "ipv4_ipam_pool_id"),
					resource.TestCheckNoResourceAttr(resourceName, "ipv4_netmask_length"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block_network_border_group", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_ipam_pool_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "main_route_table_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPC(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPC_tags(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTags_providerOnly(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("providerkey1", "providervalue1", "providerkey2", "providervalue2"),
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "value1"),
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "value1"),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTags_updateToProviderOnly(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("key1", "value1"),
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
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

func TestAccVPC_DefaultTags_updateToResourceOnly(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("key1", "value1"),
					testAccVpcConfig,
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccVPCTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
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

func TestAccVPC_DefaultTagsProviderAndResource_nonOverlappingTag(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccVPCTags1Config("resourcekey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", "resourcevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey1", "providervalue1"),
					testAccVPCTags2Config("resourcekey1", "resourcevalue1", "resourcekey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey2", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey1", "providervalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey2", "resourcevalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("providerkey2", "providervalue2"),
					testAccVPCTags1Config("resourcekey3", "resourcevalue3"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.resourcekey3", "resourcevalue3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.providerkey2", "providervalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.resourcekey3", "resourcevalue3"),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTagsProviderAndResource_overlappingTag(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey1", "providervalue1"),
					testAccVPCTags1Config("overlapkey1", "resourcevalue1"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("overlapkey1", "providervalue1", "overlapkey2", "providervalue2"),
					testAccVPCTags2Config("overlapkey1", "resourcevalue1", "overlapkey2", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey2", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", "resourcevalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey2", "resourcevalue2"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey1", "providervalue1"),
					testAccVPCTags1Config("overlapkey1", "resourcevalue2"),
				),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.overlapkey1", "resourcevalue2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.overlapkey1", "resourcevalue2"),
				),
			},
		},
	})
}

func TestAccVPC_DefaultTagsProviderAndResource_duplicateTag(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("overlapkey", "overlapvalue"),
					testAccVPCTags1Config("overlapkey", "overlapvalue"),
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`"tags" are identical to those in the "default_tags" configuration block`),
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
	var providers []*schema.Provider
	var vpc ec2.Vpc

	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCWithIgnoreChangesConfig_DynamicTagsMergedWithLocals("localkey", "localvalue"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "3"),
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
				Config: testAccVPCWithIgnoreChangesConfig_DynamicTagsMergedWithLocals("localkey", "localvalue"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.localkey", "localvalue"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "3"),
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
	var providers []*schema.Provider
	var vpc ec2.Vpc

	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCWithIgnoreChangesConfig_DynamicTags,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags_all.updated_at"),
				),
				// Dynamic tag "updated_at" will cause a perpetual diff but that's OK for this test
				// as we want to ensure subsequent applies will not result in "inconsistent final plan errors"
				// for the attribute "tags_all"
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVPCWithIgnoreChangesConfig_DynamicTags,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "tags.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
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

func TestAccVPC_defaultAndIgnoreTags(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					testAccCheckVpcUpdateTags(&vpc, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeyPrefixes1("defaultkey1", "defaultvalue1", "defaultkey"),
					testAccVPCTags1Config("key1", "value1"),
				),
				PlanOnly: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeys1("defaultkey1", "defaultvalue1"),
					testAccVPCTags1Config("key1", "value1"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPC_ignoreTags(t *testing.T) {
	var providers []*schema.Provider
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					testAccCheckVpcUpdateTags(&vpc, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:   acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey") + testAccVPCTags1Config("key1", "value1"),
				PlanOnly: true,
			},
			{
				Config:   acctest.ConfigIgnoreTagsKeys("ignorekey1") + testAccVPCTags1Config("key1", "value1"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPC_tenancy(t *testing.T) {
	var vpcDedicated ec2.Vpc
	var vpcDefault ec2.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigDedicatedTenancy(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpcDedicated),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "dedicated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcConfigDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpcDefault),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "default"),
					testAccCheckVpcIdsEqual(&vpcDedicated, &vpcDefault),
				),
			},
			{
				Config: testAccVpcConfigDedicatedTenancy(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpcDedicated),
					resource.TestCheckResourceAttr(resourceName, "instance_tenancy", "dedicated"),
					testAccCheckVpcIdsNotEqual(&vpcDedicated, &vpcDefault),
				),
			},
		},
	})
}

func TestAccVPC_updateDNSHostnames(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "false"),
				),
			},
			{
				Config: testAccVpcConfigEnableDNSHostnames(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "true"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/1301
func TestAccVPC_bothDNSOptionsSet(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigBothDnsOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_hostnames", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "true"),
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
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigDisabledDnsSupport(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_dns_support", "false"),
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

func TestAccVPC_classicLinkOptionSet(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigClassicLinkOption(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink", "true"),
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

func TestAccVPC_classicLinkDNSSupportOptionSet(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigClassicLinkDnsSupportOption(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "enable_classiclink_dns_support", "true"),
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
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigAssignGeneratedIpv6CidrBlock(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "true"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcConfigAssignGeneratedIpv6CidrBlock(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccVpcConfigAssignGeneratedIpv6CidrBlock(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "true"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_association_id"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
				),
			},
		},
	})
}

func TestAccVPC_assignGeneratedIPv6CIDRBlockWithNetworkBorderGroup(t *testing.T) {
	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"
	azDataSourceName := "data.aws_availability_zone.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckLocalZoneAvailable(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigAssignGeneratedIpv6CidrBlockWithOptionalNetworkBorderGroup(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "true"),
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
				Config: testAccVpcConfigAssignGeneratedIpv6CidrBlockWithOptionalNetworkBorderGroup(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "assign_generated_ipv6_cidr_block", "true"),
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

func TestAccVPC_IpamIpv4BasicNetmask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcIpamIpv4(rName, 28),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					testAccCheckVpcCidrPrefix(&vpc, "28"),
				),
			},
		},
	})
}

func TestAccVPC_IpamIpv4BasicExplicitCidr(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpc ec2.Vpc
	resourceName := "aws_vpc.test"
	cidr := "172.2.0.32/28"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcIpamIpv4ExplicitCidr(rName, cidr),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(resourceName, &vpc),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", cidr),
				),
			},
		},
	})
}

func testAccCheckVpcDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc" {
			continue
		}

		_, err := tfec2.FindVPCByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 VPC %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVpcUpdateTags(vpc *ec2.Vpc, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		return tfec2.UpdateTags(conn, aws.StringValue(vpc.VpcId), oldTags, newTags)
	}
}

func testAccCheckVpcCidrPrefix(vpc *ec2.Vpc, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if strings.Split(aws.StringValue(vpc.CidrBlock), "/")[1] != expected {
			return fmt.Errorf("Bad cidr prefix: got %s, expected %s", aws.StringValue(vpc.CidrBlock), expected)
		}

		return nil
	}
}

func testAccCheckVpcIdsEqual(vpc1, vpc2 *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(vpc1.VpcId) != aws.StringValue(vpc2.VpcId) {
			return fmt.Errorf("VPC IDs are not equal")
		}

		return nil
	}
}

func testAccCheckVpcIdsNotEqual(vpc1, vpc2 *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(vpc1.VpcId) == aws.StringValue(vpc2.VpcId) {
			return fmt.Errorf("VPC IDs are equal")
		}

		return nil
	}
}

const testAccVpcConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}
`

func testAccVPCTags1Config(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccVPCTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVPCWithIgnoreChangesConfig_DynamicTagsMergedWithLocals(localTagKey1, localTagValue1 string) string {
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

const testAccVPCWithIgnoreChangesConfig_DynamicTags = `
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

func testAccVpcConfigAssignGeneratedIpv6CidrBlock(rName string, assignGeneratedIpv6CidrBlock bool) string {
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

func testAccVpcConfigAssignGeneratedIpv6CidrBlockWithOptionalNetworkBorderGroup(rName string, localZoneNetworkBorderGroup bool) string {
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
}

data "aws_availability_zone" "test" {
  zone_id = data.aws_availability_zones.available.zone_ids[0]
}

resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block     = true
  cidr_block                           = "10.1.0.0/16"
  ipv6_cidr_block_network_border_group = %[2]t ? data.aws_availability_zone.test.network_border_group : data.aws_region.current.name

  tags = {
    Name = %[1]q
  }
}
`, rName, localZoneNetworkBorderGroup)
}

func testAccVpcConfigDefault(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVpcConfigEnableDNSHostnames(rName string) string {
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

func testAccVpcConfigDedicatedTenancy(rName string) string {
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

func testAccVpcConfigBothDnsOptions(rName string) string {
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

func testAccVpcConfigDisabledDnsSupport(rName string) string {
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

func testAccVpcConfigClassicLinkOption(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block         = "172.2.0.0/16"
  enable_classiclink = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVpcConfigClassicLinkDnsSupportOption(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                     = "172.2.0.0/16"
  enable_classiclink             = true
  enable_classiclink_dns_support = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVpcIPv4IPAMConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.current.name

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

func testAccVpcIpamIpv4(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVpcIPv4IPAMConfigBase(rName), fmt.Sprintf(`
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

func testAccVpcIpamIpv4ExplicitCidr(rName, cidr string) string {
	return acctest.ConfigCompose(testAccVpcIPv4IPAMConfigBase(rName), fmt.Sprintf(`
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
