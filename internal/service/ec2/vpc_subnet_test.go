// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCSubnet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`subnet/subnet-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "10.1.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "customer_owned_ipv4_pool", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "private_dns_hostname_type_on_launch", "ip-name"),
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

func TestAccVPCSubnet_tags_defaultAndIgnoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					testAccCheckSubnetUpdateTags(ctx, &subnet, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
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
					testAccVPCSubnetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
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
					testAccVPCSubnetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
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

func TestAccVPCSubnet_tags_ignoreTags(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					testAccCheckSubnetUpdateTags(ctx, &subnet, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			{
				Config: acctest.ConfigCompose(acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"), testAccVPCSubnetConfig_basic(rName)),
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
				Config: acctest.ConfigCompose(acctest.ConfigIgnoreTagsKeys("ignorekey1"), testAccVPCSubnetConfig_basic(rName)),
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

func TestAccVPCSubnet_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Config: testAccVPCSubnetConfig_ipv6(rName, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					testAccCheckSubnetIPv6CIDRBlockAssociationSet(&subnet),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName, tfjsonpath.New("assign_ipv6_address_on_creation"), knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Disable assign_ipv6_address_on_creation
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Config: testAccVPCSubnetConfig_ipv6(rName, 1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName, tfjsonpath.New("assign_ipv6_address_on_creation"), knownvalue.Bool(false),
					),
				},
			},
			{
				// Change IPv6 CIDR block
				// assign_ipv6_address_on_creation was false, so no replacement
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Config: testAccVPCSubnetConfig_ipv6(rName, 3, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName, tfjsonpath.New("assign_ipv6_address_on_creation"), knownvalue.Bool(true),
					),
				},
			},
			{
				// Force new by changing IPv6 CIDR block
				// since assign_ipv6_address_on_creation was true
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Config: testAccVPCSubnetConfig_ipv6(rName, 1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					testAccCheckSubnetIPv6CIDRBlockAssociationSet(&subnet),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceName, tfjsonpath.New("assign_ipv6_address_on_creation"), knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func TestAccVPCSubnet_enableIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_prev6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSubnetConfig_ipv6(rName, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtTrue),
				),
			},
			{
				Config: testAccVPCSubnetConfig_prev6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccVPCSubnet_availabilityZoneID(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_availabilityZoneID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone_id", "data.aws_availability_zones.available", "zone_ids.0"),
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

func TestAccVPCSubnet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceSubnet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCSubnet_customerOwnedIPv4Pool(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	coipDataSourceName := "data.aws_ec2_coip_pool.test"
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_customerOwnedv4Pool(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttrPair(resourceName, "customer_owned_ipv4_pool", coipDataSourceName, "pool_id"),
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

func TestAccVPCSubnet_mapCustomerOwnedIPOnLaunch(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_mapCustomerOwnedOnLaunch(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "map_customer_owned_ip_on_launch", acctest.CtTrue),
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

func TestAccVPCSubnet_mapPublicIPOnLaunch(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_mapPublicOnLaunch(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSubnetConfig_mapPublicOnLaunch(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtFalse),
				),
			},
			{
				Config: testAccVPCSubnetConfig_mapPublicOnLaunch(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccVPCSubnet_outpost(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, names.AttrARN),
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

func TestAccVPCSubnet_enableDNS64(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_enableDNS64(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSubnetConfig_enableDNS64(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtFalse),
				),
			},
			{
				Config: testAccVPCSubnetConfig_enableDNS64(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccVPCSubnet_ipv4ToIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv4ToIPv6Before(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccVPCSubnetConfig_ipv4ToIPv6After(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
				),
			},
		},
	})
}

func TestAccVPCSubnet_enableLNIAtDeviceIndex(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_enableLniAtDeviceIndex(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_lni_at_device_index", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSubnetConfig_enableLniAtDeviceIndex(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_lni_at_device_index", "1"),
				),
			},
			{
				Config: testAccVPCSubnetConfig_enableLniAtDeviceIndex(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_lni_at_device_index", "1"),
				),
			},
		},
	})
}

func TestAccVPCSubnet_privateDNSNameOptionsOnLaunch(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_privateDNSNameOptionsOnLaunch(rName, true, true, "resource-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_hostname_type_on_launch", "resource-name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSubnetConfig_privateDNSNameOptionsOnLaunch(rName, false, true, "ip-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_hostname_type_on_launch", "ip-name"),
				),
			},
			{
				Config: testAccVPCSubnetConfig_privateDNSNameOptionsOnLaunch(rName, true, false, "resource-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "private_dns_hostname_type_on_launch", "resource-name"),
				),
			},
		},
	})
}

func TestAccVPCSubnet_ipv6Native(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv6Native(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, ""),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ipv6_native", acctest.CtTrue),
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

func TestAccVPCSubnet_IPAM_ipv4Allocation(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv4IPAMAllocation(rName, 27),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttrPair(resourceName, "ipv4_ipam_pool_id", "aws_vpc_ipam_pool.vpc", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "ipv4_netmask_length", "27"),
					testAccCheckSubnetCIDRPrefix(&subnet, "27"),
					testAccCheckIPAMPoolAllocationExistsForSubnet(ctx, "aws_vpc_ipam_pool.vpc", &subnet),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ipv4_ipam_pool_id", "ipv4_netmask_length"},
			},
		},
	})
}

func TestAccVPCSubnet_IPAM_ipv4AllocationExplicitCIDR(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_subnet.test"
	cidr := "10.0.0.0/27"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv4IPAMAllocationExplicitCIDR(rName, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttrPair(resourceName, "ipv4_ipam_pool_id", "aws_vpc_ipam_pool.vpc", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, cidr),
					testAccCheckSubnetCIDRPrefix(&subnet, "27"),
					testAccCheckIPAMPoolAllocationExistsForSubnet(ctx, "aws_vpc_ipam_pool.vpc", &subnet),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ipv4_ipam_pool_id"},
			},
		},
	})
}

func TestAccVPCSubnet_IPAM_ipv6Allocation(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv6IPAMAllocation(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, resourceName, &subnet),
					resource.TestCheckResourceAttrPair(resourceName, "ipv6_ipam_pool_id", "aws_vpc_ipam_pool.vpc", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "ipv6_netmask_length", "60"),
					testAccCheckSubnetIPv6CIDRPrefix(&subnet, "60"),
					testAccCheckIPAMPoolAllocationExistsForSubnet(ctx, "aws_vpc_ipam_pool.vpc", &subnet),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ipv6_ipam_pool_id", "ipv6_netmask_length"},
			},
		},
	})
}

func TestAccVPCSubnet_IPAM_crossRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	poolResourceName := "aws_vpc_ipam_pool.vpc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckSubnetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipamCrossRegion(rName, 28),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubnetExistsWithProvider(ctx, resourceName, &subnet, acctest.RegionProviderFunc(ctx, acctest.AlternateRegion(), &providers)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "ipv4_ipam_pool_id", poolResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "ipv4_netmask_length", "28"),
					testAccCheckSubnetCIDRPrefix(&subnet, "28"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCIDRBlock),
					testAccCheckIPAMPoolAllocationExistsForSubnet(ctx, poolResourceName, &subnet, acctest.RegionProviderFunc(ctx, acctest.AlternateRegion(), &providers)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ipv4_ipam_pool_id", "ipv4_netmask_length"},
			},
		},
	})
}

func testAccCheckSubnetIPv6CIDRBlockAssociationSet(subnet *awstypes.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if subnet.Ipv6CidrBlockAssociationSet == nil {
			return fmt.Errorf("Expected IPV6 CIDR Block Association")
		}
		return nil
	}
}

func testAccCheckSubnetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_subnet" {
				continue
			}

			_, err := tfec2.FindSubnetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Subnet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSubnetExists(ctx context.Context, n string, v *awstypes.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Subnet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSubnetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSubnetExistsWithProvider(ctx context.Context, n string, v *awstypes.Subnet, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Subnet ID is set")
		}

		conn := providerF().Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSubnetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSubnetUpdateTags(ctx context.Context, subnet *awstypes.Subnet, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		return tfec2.UpdateTags(ctx, conn, aws.ToString(subnet.SubnetId), oldTags, newTags)
	}
}

func testAccCheckSubnetCIDRPrefix(subnet *awstypes.Subnet, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cidrBlock := aws.ToString(subnet.CidrBlock)
		parts := strings.Split(cidrBlock, "/")
		if len(parts) != 2 {
			return fmt.Errorf("Bad cidr format: got %s, expected format <ip>/<prefix>", cidrBlock)
		}
		if parts[1] != expected {
			return fmt.Errorf("Bad cidr prefix: got %s, expected /%s", cidrBlock, expected)
		}
		return nil
	}
}

func testAccCheckIPAMPoolAllocationExistsForSubnet(ctx context.Context, poolResourceName string, subnet *awstypes.Subnet, providerF ...func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[poolResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", poolResourceName)
		}

		poolID := rs.Primary.ID
		subnetID := aws.ToString(subnet.SubnetId)
		subnetCIDR := aws.ToString(subnet.CidrBlock)

		var conn *conns.AWSClient
		if len(providerF) > 0 && providerF[0] != nil {
			conn = providerF[0]().Meta().(*conns.AWSClient)
		} else {
			conn = acctest.Provider.Meta().(*conns.AWSClient)
		}

		allocations, err := tfec2.FindIPAMPoolAllocationsByIPAMPoolIDAndResourceID(ctx, conn.EC2Client(ctx), poolID, subnetID)
		if err != nil {
			return fmt.Errorf("error finding IPAM Pool (%s) allocations for subnet (%s): %w", poolID, subnetID, err)
		}

		if len(allocations) == 0 {
			return fmt.Errorf("no IPAM Pool allocation found for subnet %s in pool %s", subnetID, poolID)
		}

		allocation := allocations[0]
		if allocation.ResourceType != awstypes.IpamPoolAllocationResourceTypeSubnet {
			return fmt.Errorf("expected allocation resource type 'subnet', got %s", allocation.ResourceType)
		}

		allocationCIDR := aws.ToString(allocation.Cidr)

		// Check if allocation matches IPv4 CIDR
		if allocationCIDR == subnetCIDR && subnetCIDR != "" {
			return nil
		}

		// Check if allocation matches any IPv6 CIDR
		for _, association := range subnet.Ipv6CidrBlockAssociationSet {
			if association.Ipv6CidrBlockState.State == awstypes.SubnetCidrBlockStateCodeAssociated {
				subnetIPv6CIDR := aws.ToString(association.Ipv6CidrBlock)
				if allocationCIDR == subnetIPv6CIDR {
					return nil
				}
			}
		}

		return fmt.Errorf("allocation CIDR (%s) does not match subnet IPv4 CIDR (%s) or any associated IPv6 CIDR", allocationCIDR, subnetCIDR)
	}
}

func testAccCheckSubnetIPv6CIDRPrefix(subnet *awstypes.Subnet, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, association := range subnet.Ipv6CidrBlockAssociationSet {
			if association.Ipv6CidrBlockState != nil && association.Ipv6CidrBlockState.State == awstypes.SubnetCidrBlockStateCodeAssociated {
				if strings.Split(aws.ToString(association.Ipv6CidrBlock), "/")[1] != expected {
					return fmt.Errorf("Bad IPv6 cidr prefix: got %s, expected /%s", aws.ToString(association.Ipv6CidrBlock), expected)
				}
				return nil
			}
		}
		return fmt.Errorf("No associated IPv6 CIDR block found")
	}
}

func testAccVPCSubnetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id
}
`, rName)
}

func testAccVPCSubnetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCSubnetConfig_prev6(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.10.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSubnetConfig_ipv6(rName string, ipv6CidrSubnetIndex int, assignIPv6AddressOnCreation bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block                      = "10.10.1.0/24"
  vpc_id                          = aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, %[2]d)
  assign_ipv6_address_on_creation = %[3]t

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6CidrSubnetIndex, assignIPv6AddressOnCreation)
}

func testAccVPCSubnetConfig_availabilityZoneID(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block           = "10.1.1.0/24"
  vpc_id               = aws_vpc.test.id
  availability_zone_id = data.aws_availability_zones.available.zone_ids[0]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCSubnetConfig_customerOwnedv4Pool(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_ec2_local_gateway_route_tables" "test" {
  filter {
    name   = "outpost-arn"
    values = [data.aws_outposts_outpost.test.arn]
  }
}

data "aws_ec2_coip_pools" "test" {
  # Filtering by Local Gateway Route Table ID is documented but not working in EC2 API.
  # If there are multiple Outposts in the test account, this lookup can
  # be misaligned and cause downstream resource errors.
  #
  # filter {
  #   name   = "coip-pool.local-gateway-route-table-id"
  #   values = [tolist(data.aws_ec2_local_gateway_route_tables.test.ids)[0]]
  # }
}

data "aws_ec2_coip_pool" "test" {
  pool_id = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone               = data.aws_outposts_outpost.test.availability_zone
  cidr_block                      = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  customer_owned_ipv4_pool        = data.aws_ec2_coip_pool.test.pool_id
  map_customer_owned_ip_on_launch = true
  outpost_arn                     = data.aws_outposts_outpost.test.arn
  vpc_id                          = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSubnetConfig_mapCustomerOwnedOnLaunch(rName string, mapCustomerOwnedIpOnLaunch bool) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_ec2_local_gateway_route_tables" "test" {
  filter {
    name   = "outpost-arn"
    values = [data.aws_outposts_outpost.test.arn]
  }
}

data "aws_ec2_coip_pools" "test" {
  # Filtering by Local Gateway Route Table ID is documented but not working in EC2 API.
  # If there are multiple Outposts in the test account, this lookup can
  # be misaligned and cause downstream resource errors.
  #
  # filter {
  #   name   = "coip-pool.local-gateway-route-table-id"
  #   values = [tolist(data.aws_ec2_local_gateway_route_tables.test.ids)[0]]
  # }
}

data "aws_ec2_coip_pool" "test" {
  pool_id = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone               = data.aws_outposts_outpost.test.availability_zone
  cidr_block                      = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  customer_owned_ipv4_pool        = data.aws_ec2_coip_pool.test.pool_id
  map_customer_owned_ip_on_launch = %[2]t
  outpost_arn                     = data.aws_outposts_outpost.test.arn
  vpc_id                          = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, mapCustomerOwnedIpOnLaunch)
}

func testAccVPCSubnetConfig_mapPublicOnLaunch(rName string, mapPublicIpOnLaunch bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = %[2]t
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, mapPublicIpOnLaunch)
}

func testAccVPCSubnetConfig_enableDNS64(rName string, enableDns64 bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block                      = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  enable_dns64                    = %[2]t
  vpc_id                          = aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}
`, rName, enableDns64)
}

func testAccVPCSubnetConfig_enableLniAtDeviceIndex(rName string, deviceIndex int) string {
	return fmt.Sprintf(`


data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone          = data.aws_outposts_outpost.test.availability_zone
  cidr_block                 = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  enable_lni_at_device_index = %[2]d
  outpost_arn                = data.aws_outposts_outpost.test.arn
  vpc_id                     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, deviceIndex)
}

func testAccVPCSubnetConfig_privateDNSNameOptionsOnLaunch(rName string, enableDnsAAAA, enableDnsA bool, hostnameType string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block                      = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id                          = aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  enable_resource_name_dns_aaaa_record_on_launch = %[2]t
  enable_resource_name_dns_a_record_on_launch    = %[3]t
  private_dns_hostname_type_on_launch            = %[4]q

  tags = {
    Name = %[1]q
  }
}
`, rName, enableDnsAAAA, enableDnsA, hostnameType)
}

func testAccVPCSubnetConfig_ipv6Native(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                          = aws_vpc.test.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true
  ipv6_native                     = true

  enable_resource_name_dns_aaaa_record_on_launch = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSubnetConfig_outpost(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  cidr_block        = "10.1.1.0/24"
  outpost_arn       = data.aws_outposts_outpost.test.arn
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSubnetConfig_ipv4ToIPv6Before(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  assign_ipv6_address_on_creation                = false
  cidr_block                                     = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  enable_dns64                                   = false
  enable_resource_name_dns_aaaa_record_on_launch = false
  ipv6_cidr_block                                = null
  vpc_id                                         = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSubnetConfig_ipv4ToIPv6After(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.10.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  assign_ipv6_address_on_creation                = true
  cidr_block                                     = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  enable_dns64                                   = true
  enable_resource_name_dns_aaaa_record_on_launch = true
  ipv6_cidr_block                                = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
  vpc_id                                         = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

const testAccVPCSubnetConfig_ipamBase = `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.region
  }
}
`

func testAccVPCSubnetConfig_ipamIPv4(rName string) string {
	return acctest.ConfigCompose(testAccVPCSubnetConfig_ipamBase, fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.current.name
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "10.0.0.0/16"
}

resource "aws_vpc" "test" {
  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = 24

  depends_on = [aws_vpc_ipam_pool_cidr.test]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool" "vpc" {
  address_family      = "ipv4"
  ipam_scope_id       = aws_vpc_ipam.test.private_default_scope_id
  locale              = data.aws_region.current.name
  source_ipam_pool_id = aws_vpc_ipam_pool.test.id

  source_resource {
    resource_id     = aws_vpc.test.id
    resource_owner  = data.aws_caller_identity.current.account_id
    resource_region = data.aws_region.current.name
    resource_type   = "vpc"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCSubnetConfig_ipamIPv6(rName string) string {
	return acctest.ConfigCompose(testAccVPCSubnetConfig_ipamBase, fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv6"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.current.name
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id   = aws_vpc_ipam_pool.test.id
  netmask_length = 52
}

resource "aws_vpc" "test" {
  cidr_block          = "10.1.0.0/16"
  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = 56

  depends_on = [aws_vpc_ipam_pool_cidr.test]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool" "vpc" {
  address_family      = "ipv6"
  ipam_scope_id       = aws_vpc_ipam.test.private_default_scope_id
  locale              = data.aws_region.current.name
  source_ipam_pool_id = aws_vpc_ipam_pool.test.id

  source_resource {
    resource_id     = aws_vpc.test.id
    resource_owner  = data.aws_caller_identity.current.account_id
    resource_region = data.aws_region.current.name
    resource_type   = "vpc"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCSubnetConfig_ipv4IPAMAllocation(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVPCSubnetConfig_ipamIPv4(rName), fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc_ipam_pool_cidr" "vpc" {
  ipam_pool_id = aws_vpc_ipam_pool.vpc.id
  cidr         = aws_vpc.test.cidr_block
}

resource "aws_subnet" "test" {
  vpc_id              = aws_vpc.test.id
  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.vpc.id
  ipv4_netmask_length = %[1]d
  availability_zone   = data.aws_availability_zones.available.names[0]

  depends_on = [aws_vpc_ipam_pool_cidr.vpc]
}
`, netmaskLength))
}

func testAccVPCSubnetConfig_ipv4IPAMAllocationExplicitCIDR(rName string, cidr string) string {
	return acctest.ConfigCompose(testAccVPCSubnetConfig_ipamIPv4(rName), fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc_ipam_pool_cidr" "vpc" {
  ipam_pool_id = aws_vpc_ipam_pool.vpc.id
  cidr         = aws_vpc.test.cidr_block
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.vpc.id
  cidr_block        = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]

  depends_on = [aws_vpc_ipam_pool_cidr.vpc]
}
`, cidr))
}

func testAccVPCSubnetConfig_ipv6IPAMAllocation(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVPCSubnetConfig_ipamIPv6(rName), fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc_ipam_pool_cidr" "vpc" {
  ipam_pool_id = aws_vpc_ipam_pool.vpc.id
  cidr         = aws_vpc.test.ipv6_cidr_block
}

resource "aws_subnet" "test" {
  vpc_id                                         = aws_vpc.test.id
  ipv6_native                                    = true
  assign_ipv6_address_on_creation                = true
  ipv6_ipam_pool_id                              = aws_vpc_ipam_pool.vpc.id
  ipv6_netmask_length                            = %[1]d
  availability_zone                              = data.aws_availability_zones.available.names[0]
  enable_resource_name_dns_aaaa_record_on_launch = true

  depends_on = [aws_vpc_ipam_pool_cidr.vpc]
}
`, netmaskLength))
}

func testAccVPCSubnetConfig_ipamCrossRegion(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = awsalternate
}

data "aws_caller_identity" "current" {}

data "aws_availability_zones" "alternate" {
  provider = awsalternate
  state    = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc_ipam" "test" {
  operating_regions {
    region_name = data.aws_region.current.name
  }

  operating_regions {
    region_name = data.aws_region.alternate.name
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.alternate.name

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "10.0.0.0/16"
}

resource "aws_vpc" "test" {
  provider = awsalternate

  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = 24

  depends_on = [aws_vpc_ipam_pool_cidr.test]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool" "vpc" {
  address_family      = "ipv4"
  ipam_scope_id       = aws_vpc_ipam.test.private_default_scope_id
  locale              = data.aws_region.alternate.name
  source_ipam_pool_id = aws_vpc_ipam_pool.test.id

  source_resource {
    resource_id     = aws_vpc.test.id
    resource_owner  = data.aws_caller_identity.current.account_id
    resource_region = data.aws_region.alternate.name
    resource_type   = "vpc"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool_cidr" "vpc" {
  ipam_pool_id = aws_vpc_ipam_pool.vpc.id
  cidr         = aws_vpc.test.cidr_block
}

resource "aws_subnet" "test" {
  provider = awsalternate

  vpc_id              = aws_vpc.test.id
  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.vpc.id
  ipv4_netmask_length = %[2]d
  availability_zone   = data.aws_availability_zones.alternate.names[0]

  depends_on = [aws_vpc_ipam_pool_cidr.vpc]

  tags = {
    Name = %[1]q
  }
}
`, rName, netmaskLength))
}
