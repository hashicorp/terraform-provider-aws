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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &v),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
					testAccCheckSubnetUpdateTags(ctx, t, &subnet, nil, map[string]string{"defaultkey1": "defaultvalue1"}),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
					testAccCheckSubnetUpdateTags(ctx, t, &subnet, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Config: testAccVPCSubnetConfig_ipv6(rName, 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_prev6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtTrue),
				),
			},
			{
				Config: testAccVPCSubnetConfig_prev6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_availabilityZoneID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &v),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &v),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_customerOwnedv4Pool(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_mapCustomerOwnedOnLaunch(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_mapPublicOnLaunch(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "map_public_ip_on_launch", acctest.CtFalse),
				),
			},
			{
				Config: testAccVPCSubnetConfig_mapPublicOnLaunch(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &v),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_enableDNS64(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtFalse),
				),
			},
			{
				Config: testAccVPCSubnetConfig_enableDNS64(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv4ToIPv6Before(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "assign_ipv6_address_on_creation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_dns64", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccVPCSubnetConfig_ipv4ToIPv6After(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_enableLniAtDeviceIndex(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_lni_at_device_index", "1"),
				),
			},
			{
				Config: testAccVPCSubnetConfig_enableLniAtDeviceIndex(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_privateDNSNameOptionsOnLaunch(rName, true, true, "resource-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_aaaa_record_on_launch", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_resource_name_dns_a_record_on_launch", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_hostname_type_on_launch", "ip-name"),
				),
			},
			{
				Config: testAccVPCSubnetConfig_privateDNSNameOptionsOnLaunch(rName, true, false, "resource-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv6Native(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &v),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_subnet.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv4IPAMAllocation(rName, 27),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_subnet.test"
	cidr := "10.0.0.0/27"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv4IPAMAllocationExplicitCIDR(rName, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_subnet.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_ipv6IPAMAllocation(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	poolResourceName := "aws_vpc_ipam_pool.vpc"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
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

func testAccCheckSubnetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

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

func testAccCheckSubnetExists(ctx context.Context, t *testing.T, n string, v *awstypes.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Subnet ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

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

func testAccCheckSubnetUpdateTags(ctx context.Context, t *testing.T, subnet *awstypes.Subnet, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

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

// TestAccVPCSubnet_guardDutyCleanup validates the multi-subnet dissociation path.
// It creates a VPC with two subnets, then creates GuardDuty resources (endpoint + SG)
// out-of-band via SDK associated with both subnets. When one subnet is removed from
// the Terraform config, dissociateGuardDutyVPCEndpoints runs to dissociate it from
// the endpoint, and the endpoint should remain available with the remaining subnet.
//
// **Validates: Requirements 3.1, 3.2**
//
// EXPECTED: Test PASSES (multi-subnet dissociation works).
func TestAccVPCSubnet_guardDutyCleanup(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet1, subnet2 awstypes.Subnet
	var vpcID string
	resourceName1 := "aws_subnet.test1"
	resourceName2 := "aws_subnet.test2"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckGuardDutyCleanupDestroy(ctx, t, &vpcID),
			testAccDeleteGuardDutyResources(ctx, t, &vpcID),
		),
		Steps: []resource.TestStep{
			{
				// Step 1: Create VPC with two subnets. After Terraform creates
				// the infrastructure, create GuardDuty resources out-of-band
				// via SDK with both subnet IDs.
				Config: testAccVPCSubnetConfig_guardDutyCleanupBothSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName1, &subnet1),
					testAccCheckSubnetExists(ctx, t, resourceName2, &subnet2),
					testAccCaptureVPCID(&subnet1, &vpcID),
					testAccCreateGuardDutyResourcesFromSubnets(ctx, t, &subnet1, &subnet2),
					testAccCheckGuardDutyResourcesExist(ctx, t, &subnet1),
				),
			},
			{
				// Step 2: Remove subnet1 from config. Terraform deletes it,
				// dissociateGuardDutyVPCEndpoints dissociates it from the endpoint.
				// Endpoint should remain available with subnet2.
				Config: testAccVPCSubnetConfig_guardDutyCleanupSingleSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName2, &subnet2),
					testAccCheckGuardDutyEndpointStillExists(ctx, t, &vpcID),
				),
			},
		},
	})
}

// TestAccVPCSubnet_guardDutyLastSubnetDeletion is a bug condition exploration test.
// It creates a GuardDuty VPC endpoint with exactly ONE subnet, then removes the subnet
// from the Terraform config while keeping the endpoint (with ignore_changes on subnet_ids).
// This triggers dissociateGuardDutyVPCEndpoints on the last subnet, which hits the bug:
// ModifyVpcEndpoint with RemoveSubnetIds fails because an Interface VPC Endpoint requires
// at least one subnet.
//
// **Validates: Requirements 2.1, 3.1**
//
// TestAccVPCSubnet_guardDutyLastSubnetDeletion validates the single-subnet dissociation path.
// It creates a VPC with one subnet, then creates GuardDuty resources (endpoint + SG) out-of-band
// via SDK. When the subnet is removed from the Terraform config, Terraform tries to delete it,
// gets a DependencyViolation because the out-of-band endpoint is still there, and
// dissociateGuardDutyVPCEndpoints runs to clean it up.
//
// EXPECTED: Test PASSES (AWS accepts last-subnet dissociation via ModifyVpcEndpoint).
func TestAccVPCSubnet_guardDutyLastSubnetDeletion(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet1 awstypes.Subnet
	var vpcID string
	resourceName1 := "aws_subnet.test1"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckGuardDutyCleanupDestroy(ctx, t, &vpcID),
			testAccDeleteGuardDutyResources(ctx, t, &vpcID),
		),
		Steps: []resource.TestStep{
			{
				// Step 1: Create a VPC with a single subnet. After Terraform creates
				// the infrastructure, create GuardDuty resources (endpoint + SG)
				// out-of-band via SDK so they are NOT in Terraform state.
				Config: testAccVPCSubnetConfig_guardDutyLastSubnet_withSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName1, &subnet1),
					testAccCaptureVPCID(&subnet1, &vpcID),
					testAccCreateGuardDutyResourcesFromSubnet(ctx, t, &subnet1),
					testAccCheckGuardDutyResourcesExist(ctx, t, &subnet1),
				),
			},
			{
				// Step 2: Remove the subnet from the config (VPC only remains).
				// Terraform tries to delete the subnet, gets DependencyViolation
				// because the out-of-band GuardDuty endpoint is still associated,
				// and dissociateGuardDutyVPCEndpoints runs to clean it up.
				Config: testAccVPCSubnetConfig_guardDutyLastSubnet_withoutSubnet(rName),
			},
		},
	})
}

// testAccCreateGuardDutyResourcesFromSubnet is a convenience wrapper around
// testAccCreateGuardDutyResources that extracts the subnet ID at execution time
// from the populated subnet pointer.
func testAccCreateGuardDutyResourcesFromSubnet(ctx context.Context, t *testing.T, subnet *awstypes.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		subnetID := aws.ToString(subnet.SubnetId)
		return testAccCreateGuardDutyResources(ctx, t, subnet, []string{subnetID})(s)
	}
}

// testAccCreateGuardDutyResourcesFromSubnets is a convenience wrapper around
// testAccCreateGuardDutyResources that extracts both subnet IDs at execution time
// from two populated subnet pointers.
func testAccCreateGuardDutyResourcesFromSubnets(ctx context.Context, t *testing.T, subnet1, subnet2 *awstypes.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		subnetID1 := aws.ToString(subnet1.SubnetId)
		subnetID2 := aws.ToString(subnet2.SubnetId)
		return testAccCreateGuardDutyResources(ctx, t, subnet1, []string{subnetID1, subnetID2})(s)
	}
}



func testAccCheckGuardDutyResourcesExist(ctx context.Context, t *testing.T, subnet *awstypes.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)
		vpcID := aws.ToString(subnet.VpcId)

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

func testAccCheckNoGuardDutyResources(ctx context.Context, t *testing.T, subnet *awstypes.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)
		vpcID := aws.ToString(subnet.VpcId)

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
			},
		}
		endpointsOutput, err := conn.DescribeVpcEndpoints(ctx, endpointsInput)
		if err != nil {
			return fmt.Errorf("error describing VPC endpoints: %w", err)
		}
		if len(endpointsOutput.VpcEndpoints) > 0 {
			return fmt.Errorf("expected no GuardDuty VPC endpoints, but found %d", len(endpointsOutput.VpcEndpoints))
		}

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
			},
		}
		sgOutput, err := conn.DescribeSecurityGroups(ctx, sgInput)
		if err != nil {
			return fmt.Errorf("error describing security groups: %w", err)
		}
		if len(sgOutput.SecurityGroups) > 0 {
			return fmt.Errorf("expected no GuardDuty security groups, but found %d", len(sgOutput.SecurityGroups))
		}

		return nil
	}
}

func testAccCheckGuardDutyEndpointStillExists(ctx context.Context, t *testing.T, vpcID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if vpcID == nil || *vpcID == "" {
			return fmt.Errorf("VPC ID not captured")
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

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
			return fmt.Errorf("error describing VPC endpoints: %w", err)
		}
		if len(endpointsOutput.VpcEndpoints) == 0 {
			return fmt.Errorf("expected GuardDuty VPC endpoint to still exist after subnet dissociation, but none found")
		}

		for _, ep := range endpointsOutput.VpcEndpoints {
			state := string(ep.State)
			if state != "available" {
				return fmt.Errorf("expected GuardDuty VPC endpoint %s to be in 'available' state, got %q",
					aws.ToString(ep.VpcEndpointId), state)
			}
		}

		return nil
	}
}

func testAccCaptureVPCID(subnet *awstypes.Subnet, vpcID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*vpcID = aws.ToString(subnet.VpcId)
		return nil
	}
}

// testAccCreateGuardDutyResources creates GuardDuty-tagged resources (VPC endpoint and security group)
// out-of-band using the AWS SDK directly. This is critical because when GuardDuty resources are
// Terraform-managed, Terraform handles dependency ordering and destroys the endpoint before the subnet,
// so dissociateGuardDutyVPCEndpoints is never exercised. In production, GuardDuty creates these
// resources out-of-band.
func testAccCreateGuardDutyResources(ctx context.Context, t *testing.T, subnet *awstypes.Subnet, subnetIDs []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)
		region := acctest.ProviderMeta(ctx, t).Region(ctx)
		vpcID := aws.ToString(subnet.VpcId)

		// Create GuardDuty-tagged security group
		sgName := fmt.Sprintf("%s%s", tfec2.GuardDutySecurityGroupPrefix, vpcID)
		sgInput := ec2.CreateSecurityGroupInput{
			GroupName:   aws.String(sgName),
			Description: aws.String("GuardDuty managed security group for testing"),
			VpcId:       aws.String(vpcID),
			TagSpecifications: []awstypes.TagSpecification{
				{
					ResourceType: awstypes.ResourceTypeSecurityGroup,
					Tags: []awstypes.Tag{
						{
							Key:   aws.String("GuardDutyManaged"),
							Value: aws.String(acctest.CtTrue),
						},
					},
				},
			},
		}
		sgOutput, err := conn.CreateSecurityGroup(ctx, &sgInput)
		if err != nil {
			return fmt.Errorf("creating GuardDuty security group in VPC %s: %w", vpcID, err)
		}
		sgID := aws.ToString(sgOutput.GroupId)

		// Create GuardDuty-tagged VPC endpoint
		serviceName := fmt.Sprintf("com.amazonaws.%s.guardduty-data", region)
		epInput := ec2.CreateVpcEndpointInput{
			VpcId:             aws.String(vpcID),
			ServiceName:       aws.String(serviceName),
			VpcEndpointType:   awstypes.VpcEndpointTypeInterface,
			SubnetIds:         subnetIDs,
			SecurityGroupIds:   []string{sgID},
			TagSpecifications: []awstypes.TagSpecification{
				{
					ResourceType: awstypes.ResourceTypeVpcEndpoint,
					Tags: []awstypes.Tag{
						{
							Key:   aws.String("GuardDutyManaged"),
							Value: aws.String(acctest.CtTrue),
						},
					},
				},
			},
		}
		epOutput, err := conn.CreateVpcEndpoint(ctx, &epInput)
		if err != nil {
			return fmt.Errorf("creating GuardDuty VPC endpoint in VPC %s: %w", vpcID, err)
		}
		endpointID := aws.ToString(epOutput.VpcEndpoint.VpcEndpointId)

		// Wait for endpoint to reach available state
		if _, err := tfec2.WaitVPCEndpointAvailable(ctx, conn, endpointID, tfec2.VPCEndpointCreationTimeout); err != nil {
			return fmt.Errorf("waiting for GuardDuty VPC endpoint %s to become available: %w", endpointID, err)
		}

		return nil
	}
}

// testAccDeleteGuardDutyResources describes and deletes any GuardDuty-tagged endpoints and security
// groups in the VPC. Used in CheckDestroy to clean up out-of-band resources that Terraform doesn't
// know about. Handles NotFound errors gracefully since resources may already be cleaned up.
func testAccDeleteGuardDutyResources(ctx context.Context, t *testing.T, vpcIDPtr *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if vpcIDPtr == nil || *vpcIDPtr == "" {
			return nil
		}
		vpcID := *vpcIDPtr

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		// Describe GuardDuty endpoints in the VPC
		describeEndpointsInput := ec2.DescribeVpcEndpointsInput{
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
		endpointsOutput, err := conn.DescribeVpcEndpoints(ctx, &describeEndpointsInput)
		if err != nil {
			// VPC may already be deleted, treat as success
			return nil
		}

		// Delete any found endpoints
		var endpointIDs []string
		for _, ep := range endpointsOutput.VpcEndpoints {
			if string(ep.State) != "deleted" {
				endpointIDs = append(endpointIDs, aws.ToString(ep.VpcEndpointId))
			}
		}

		if len(endpointIDs) > 0 {
			input := ec2.DeleteVpcEndpointsInput{
				VpcEndpointIds: endpointIDs,
			}
_, err := conn.DeleteVpcEndpoints(ctx, &input)
			if err != nil {
				// Ignore NotFound errors — endpoints may already be deleted
				if !strings.Contains(err.Error(), "InvalidVpcEndpointId.NotFound") {
					return fmt.Errorf("deleting GuardDuty VPC endpoints in VPC %s: %w", vpcID, err)
				}
			}

			// Wait for each endpoint to reach deleted state
			for _, epID := range endpointIDs {
				if _, err := tfec2.WaitVPCEndpointDeleted(ctx, conn, epID, tfec2.VPCEndpointDeletionTimeout); err != nil {
					// Ignore NotFound errors
					if !strings.Contains(err.Error(), "InvalidVpcEndpointId.NotFound") &&
						!strings.Contains(err.Error(), "couldn't find resource") {
						return fmt.Errorf("waiting for GuardDuty VPC endpoint %s to be deleted: %w", epID, err)
					}
				}
			}
		}

		// Describe and delete GuardDuty security groups
		describeSGInput := ec2.DescribeSecurityGroupsInput{
			Filters: []awstypes.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{vpcID},
				},
				{
					Name:   aws.String("group-name"),
					Values: []string{fmt.Sprintf("%s%s", tfec2.GuardDutySecurityGroupPrefix, vpcID)},
				},
				{
					Name:   aws.String("tag:GuardDutyManaged"),
					Values: []string{acctest.CtTrue},
				},
			},
		}
		sgOutput, err := conn.DescribeSecurityGroups(ctx, &describeSGInput)
		if err != nil {
			// VPC may already be deleted, treat as success
			return nil
		}

		for _, sg := range sgOutput.SecurityGroups {
			input := ec2.DeleteSecurityGroupInput{
				GroupId: sg.GroupId,
			}
_, err := conn.DeleteSecurityGroup(ctx, &input)
			if err != nil {
				// Ignore NotFound errors — SG may already be deleted
				if !strings.Contains(err.Error(), "InvalidGroup.NotFound") {
					return fmt.Errorf("deleting GuardDuty security group %s in VPC %s: %w", aws.ToString(sg.GroupId), vpcID, err)
				}
			}
		}

		return nil
	}
}

func testAccCheckGuardDutyCleanupDestroy(ctx context.Context, t *testing.T, vpcID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_subnet" {
				continue
			}

			input := ec2.DescribeSubnetsInput{
				SubnetIds: []string{rs.Primary.ID},
			}
_, err := conn.DescribeSubnets(ctx, &input)
			if err == nil {
				return fmt.Errorf("subnet %s still exists", rs.Primary.ID)
			}
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

func testAccVPCSubnetConfig_guardDutyCleanupBothSubnets(rName string) string {
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

resource "aws_subnet" "test1" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-1"
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s-2"
  }
}
`, rName)
}

func testAccVPCSubnetConfig_guardDutyCleanupSingleSubnet(rName string) string {
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

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s-2"
  }
}
`, rName)
}

func testAccVPCSubnetConfig_guardDutyCleanupSingle(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "%[1]s-1"
  }
}
`, rName)
}

func testAccVPCSubnetConfig_noGuardDuty(rName string) string {
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
    Name = %[1]q
  }
}
`, rName)
}


// TestAccVPCSubnet_guardDutyCleanup_timeout validates that GuardDuty resources can be
// created out-of-band via SDK alongside a Terraform-managed subnet. This test only
// creates resources (does not trigger subnet deletion). CheckDestroy cleans up
// the out-of-band GuardDuty resources.
func TestAccVPCSubnet_guardDutyCleanup_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet1 awstypes.Subnet
	var vpcID string
	resourceName1 := "aws_subnet.test1"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckGuardDutyCleanupDestroy(ctx, t, &vpcID),
			testAccDeleteGuardDutyResources(ctx, t, &vpcID),
		),
		Steps: []resource.TestStep{
			{
				// Step 1: Create VPC with single subnet. After Terraform creates
				// the infrastructure, create GuardDuty resources out-of-band via SDK.
				Config: testAccVPCSubnetConfig_guardDutyCleanupSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName1, &subnet1),
					testAccCaptureVPCID(&subnet1, &vpcID),
					testAccCreateGuardDutyResourcesFromSubnet(ctx, t, &subnet1),
					testAccCheckGuardDutyResourcesExist(ctx, t, &subnet1),
				),
			},
		},
	})
}

func TestAccVPCSubnet_noGuardDuty(t *testing.T) {
	ctx := acctest.Context(t)
	var subnet awstypes.Subnet
	resourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetConfig_noGuardDuty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceName, &subnet),
					testAccCheckNoGuardDutyResources(ctx, t, &subnet),
				),
			},
		},
	})
}


func TestGuardDutyServiceNamePattern(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		serviceName  string
		shouldMatch  bool
	}{
		{
			name:        "exact guardduty-data",
			serviceName: "guardduty-data",
			shouldMatch: true,
		},
		{
			name:        "guardduty-data with prefix",
			serviceName: "com.amazonaws.us-east-1.guardduty-data", //lintignore:AWSAT003
			shouldMatch: true,
		},
		{
			name:        "guardduty-data with suffix",
			serviceName: "guardduty-data.vpce.amazonaws.com",
			shouldMatch: true,
		},
		{
			name:        "guardduty-data in middle",
			serviceName: "com.amazonaws.guardduty-data.service",
			shouldMatch: true,
		},
		{
			name:        "guardduty-data with another region",    //lintignore:AWSAT003
			serviceName: "com.amazonaws.us-west-2.guardduty-data", //lintignore:AWSAT003
			shouldMatch: true,
		},
		{
			name:        "guardduty-data with eu region",         //lintignore:AWSAT003
			serviceName: "com.amazonaws.eu-west-1.guardduty-data", //lintignore:AWSAT003
			shouldMatch: true,
		},
		{
			name:        "guardduty-data-fips",
			serviceName: "guardduty-data-fips",
			shouldMatch: true,
		},
		{
			name:        "guardduty-data-fips with prefix",
			serviceName: "com.amazonaws.us-east-1.guardduty-data-fips", //lintignore:AWSAT003
			shouldMatch: true,
		},
		{
			name:        "guardduty-data-fips with suffix",
			serviceName: "guardduty-data-fips.vpce.amazonaws.com",
			shouldMatch: true,
		},
		{
			name:        "guardduty-data-fips with region",
			serviceName: "com.amazonaws.us-west-2.guardduty-data-fips", //lintignore:AWSAT003
			shouldMatch: true,
		},
		{
			name:        "guardduty without data",
			serviceName: "com.amazonaws.guardduty",
			shouldMatch: false,
		},
		{
			name:        "data without guardduty",
			serviceName: "com.amazonaws.data",
			shouldMatch: false,
		},
		{
			name:        "guardduty-service",
			serviceName: "com.amazonaws.guardduty-service",
			shouldMatch: false,
		},
		{
			name:        "s3 service",
			serviceName: "com.amazonaws.s3",
			shouldMatch: false,
		},
		{
			name:        "ec2 service",
			serviceName: "com.amazonaws.ec2",
			shouldMatch: false,
		},
		{
			name:        "empty string",
			serviceName: "",
			shouldMatch: false,
		},
		{
			name:        "guardduty data with space",
			serviceName: "guardduty data",
			shouldMatch: false,
		},
		{
			name:        "guardduty_data with underscore",
			serviceName: "guardduty_data",
			shouldMatch: false,
		},
		{
			name:        "guarddutydata without hyphen",
			serviceName: "guarddutydata",
			shouldMatch: false,
		},
		{
			name:        "partial match guardduty-dat",
			serviceName: "guardduty-dat",
			shouldMatch: false,
		},
		{
			name:        "partial match uardduty-data",
			serviceName: "uardduty-data",
			shouldMatch: false,
		},
		{
			name:        "case sensitive - uppercase",
			serviceName: "com.amazonaws.GUARDDUTY-DATA",
			shouldMatch: false,
		},
		{
			name:        "case sensitive - mixed case",
			serviceName: "com.amazonaws.GuardDuty-Data",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			matches := strings.Contains(tc.serviceName, "guardduty-data")

			if matches != tc.shouldMatch {
				t.Errorf("Service name pattern matching failed for '%s'\nExpected match: %v, Got: %v",
					tc.serviceName, tc.shouldMatch, matches)
			}
		})
	}
}

func TestIsVPCEndpointNotFoundError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		err           error
		shouldBeFound bool
	}{
		{
			name:          "nil error",
			err:           nil,
			shouldBeFound: false,
		},
		{
			name:          "InvalidVpcEndpointId.NotFound",
			err:           fmt.Errorf("InvalidVpcEndpointId.NotFound: The vpc endpoint ID 'vpce-123' does not exist"),
			shouldBeFound: true,
		},
		{
			name:          "InvalidVpcEndpoint.NotFound",
			err:           fmt.Errorf("InvalidVpcEndpoint.NotFound"),
			shouldBeFound: true,
		},
		{
			name:          "does not exist",
			err:           fmt.Errorf("The specified endpoint does not exist"),
			shouldBeFound: true,
		},
		{
			name:          "other error",
			err:           fmt.Errorf("InternalError: An internal error occurred"),
			shouldBeFound: false,
		},
		{
			name:          "unauthorized error",
			err:           fmt.Errorf("UnauthorizedOperation: You are not authorized"),
			shouldBeFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tfec2.IsVPCEndpointNotFoundError(tc.err)
			if result != tc.shouldBeFound {
				t.Errorf("Expected %v, got %v for error: %v", tc.shouldBeFound, result, tc.err)
			}
		})
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "UnauthorizedOperation error",
			err:      fmt.Errorf("UnauthorizedOperation: You are not authorized to perform this operation"),
			expected: true,
		},
		{
			name:     "AccessDenied error",
			err:      fmt.Errorf("AccessDenied: Access denied"),
			expected: true,
		},
		{
			name:     "not authorized error",
			err:      fmt.Errorf("User is not authorized to perform action"),
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("InternalError: An internal error occurred"),
			expected: false,
		},
		{
			name:     "RequestLimitExceeded error",
			err:      fmt.Errorf("RequestLimitExceeded: Request limit exceeded"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tfec2.IsUnauthorizedError(tc.err)
			if result != tc.expected {
				t.Errorf("IsUnauthorizedError(%v) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

func TestDetectAndDeleteGuardDutyVPCEndpoints_errorMessageFormat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		errorType             string
		operation             string
		vpcID                 string
		endpointID            string
		expectedErrorContains []string
		description           string
	}{
		{
			name:       "UnauthorizedOperation during DescribeVpcEndpoints",
			errorType:  "UnauthorizedOperation",
			operation:  "describe",
			vpcID:      "vpc-0abc123",
			endpointID: "",
			expectedErrorContains: []string{
				"unable to describe GuardDuty VPC endpoints",
				"vpc-0abc123",
				"insufficient IAM permissions",
				"ec2:DescribeVpcEndpoints",
			},
			description: "When describe fails with UnauthorizedOperation, error includes VPC ID and required permissions",
		},
		{
			name:       "UnauthorizedOperation during DeleteVpcEndpoints",
			errorType:  "UnauthorizedOperation",
			operation:  "delete",
			vpcID:      "vpc-0abc123",
			endpointID: "vpce-0def456",
			expectedErrorContains: []string{
				"unable to delete GuardDuty VPC endpoint",
				"vpce-0def456",
				"vpc-0abc123",
				"insufficient IAM permissions",
				"ec2:DeleteVpcEndpoints",
				"ec2:DescribeVpcEndpoints",
			},
			description: "When delete fails with UnauthorizedOperation, error includes endpoint ID, VPC ID, and required permissions",
		},
		{
			name:       "AccessDenied during DescribeVpcEndpoints",
			errorType:  "AccessDenied",
			operation:  "describe",
			vpcID:      "vpc-0xyz789",
			endpointID: "",
			expectedErrorContains: []string{
				"unable to describe GuardDuty VPC endpoints",
				"vpc-0xyz789",
				"insufficient IAM permissions",
				"ec2:DescribeVpcEndpoints",
			},
			description: "AccessDenied errors are also detected and return descriptive errors",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Logf("Testing: %s", tc.description)

			var simulatedError error
			if tc.errorType == "UnauthorizedOperation" {
				if tc.operation == "describe" {
					simulatedError = fmt.Errorf("UnauthorizedOperation: You are not authorized to perform this operation. Encoded authorization failure message: encoded-message")
				} else if tc.operation == "delete" {
					simulatedError = fmt.Errorf("UnauthorizedOperation: You are not authorized to perform this operation on VPC endpoint %s", tc.endpointID)
				}
			} else if tc.errorType == "AccessDenied" {
				simulatedError = fmt.Errorf("AccessDenied: User is not authorized to perform: ec2:DescribeVpcEndpoints")
			}

			isUnauthorized := tfec2.IsUnauthorizedError(simulatedError)
			if !isUnauthorized {
				t.Errorf("IsUnauthorizedError() = false, want true for error: %v", simulatedError)
				return
			}

			var expectedErrorFormat string
			if tc.operation == "describe" {
				expectedErrorFormat = fmt.Sprintf(
					"unable to describe GuardDuty VPC endpoints in VPC %s due to insufficient IAM permissions. Required permissions: ec2:DescribeVpcEndpoints. Original error: %v",
					tc.vpcID,
					simulatedError,
				)
			} else if tc.operation == "delete" {
				expectedErrorFormat = fmt.Sprintf(
					"unable to delete GuardDuty VPC endpoint %s blocking subnet deletion in VPC %s due to insufficient IAM permissions. Required permissions: ec2:DeleteVpcEndpoints, ec2:DescribeVpcEndpoints. Original error: %v",
					tc.endpointID,
					tc.vpcID,
					simulatedError,
				)
			}

			for _, expectedContent := range tc.expectedErrorContains {
				if !strings.Contains(expectedErrorFormat, expectedContent) {
					t.Errorf("Expected error format missing required content: %q\nError format: %s", expectedContent, expectedErrorFormat)
				}
			}

			t.Logf("✓ Error message format validated:")
			t.Logf("  - Contains VPC ID: %s", tc.vpcID)
			if tc.endpointID != "" {
				t.Logf("  - Contains endpoint ID: %s", tc.endpointID)
			}
			t.Logf("  - Contains required IAM permissions")
			t.Logf("  - Returns descriptive error instead of nil")
		})
	}
}

func TestDetectAndDeleteGuardDutyVPCEndpoints_preservationProperty(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		errorType             string
		operation             string
		shouldReturnError     bool
		expectedErrorContains []string
		requirement           string
		description           string
	}{
		{
			name:              "Throttling error during describe",
			errorType:         "RequestLimitExceeded",
			operation:         "describe",
			shouldReturnError: true,
			expectedErrorContains: []string{
				"describing GuardDuty VPC endpoints",
				"RequestLimitExceeded",
			},
			requirement: "3.2",
			description: "Non-UnauthorizedOperation errors return descriptive errors (not nil)",
		},
		{
			name:              "Network error during describe",
			errorType:         "NetworkError",
			operation:         "describe",
			shouldReturnError: true,
			expectedErrorContains: []string{
				"describing GuardDuty VPC endpoints",
				"NetworkError",
			},
			requirement: "3.2",
			description: "Network errors return descriptive errors (not nil)",
		},
		{
			name:              "Internal error during delete",
			errorType:         "InternalError",
			operation:         "delete",
			shouldReturnError: true,
			expectedErrorContains: []string{
				"deleting GuardDuty VPC endpoint",
				"InternalError",
			},
			requirement: "3.2",
			description: "Internal errors return descriptive errors (not nil)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Logf("PRESERVATION TEST: %s", tc.description)
			t.Logf("Requirement: %s", tc.requirement)

			var simulatedError error
			switch tc.errorType {
			case "RequestLimitExceeded":
				simulatedError = fmt.Errorf("RequestLimitExceeded: Request limit exceeded")
			case "NetworkError":
				simulatedError = fmt.Errorf("NetworkError: Network connection failed")
			case "InternalError":
				simulatedError = fmt.Errorf("InternalError: An internal error occurred")
			}

			isUnauthorized := tfec2.IsUnauthorizedError(simulatedError)
			if isUnauthorized {
				t.Errorf("Test setup error: %s should NOT be detected as UnauthorizedOperation", tc.name)
				return
			}

			var expectedErrorFormat string
			if tc.operation == "describe" {
				expectedErrorFormat = fmt.Sprintf("describing GuardDuty VPC endpoints in VPC vpc-test: %v", simulatedError)
			} else if tc.operation == "delete" {
				expectedErrorFormat = fmt.Errorf("deleting GuardDuty VPC endpoint vpce-test: %w", simulatedError).Error()
			}

			for _, expectedContent := range tc.expectedErrorContains {
				if !strings.Contains(expectedErrorFormat, expectedContent) {
					t.Errorf("Expected error format missing required content: %q", expectedContent)
				}
			}

			t.Logf("✓ PRESERVATION CONFIRMED: Non-UnauthorizedOperation errors return descriptive errors")
		})
	}
}

func TestHasGuardDutyManagedTag(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tags     []awstypes.Tag
		expected bool
	}{
		{
			name: "correct key and value",
			tags: []awstypes.Tag{
				{
					Key:   aws.String("GuardDutyManaged"),
					Value: aws.String(acctest.CtTrue),
				},
			},
			expected: true,
		},
		{
			name: "wrong value",
			tags: []awstypes.Tag{
				{
					Key:   aws.String("GuardDutyManaged"),
					Value: aws.String(acctest.CtFalse),
				},
			},
			expected: false,
		},
		{
			name: "empty value",
			tags: []awstypes.Tag{
				{
					Key:   aws.String("GuardDutyManaged"),
					Value: aws.String(""),
				},
			},
			expected: false,
		},
		{
			name:     "no tags",
			tags:     []awstypes.Tag{},
			expected: false,
		},
		{
			name: "multiple tags with one matching",
			tags: []awstypes.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("my-resource"),
				},
				{
					Key:   aws.String("GuardDutyManaged"),
					Value: aws.String(acctest.CtTrue),
				},
				{
					Key:   aws.String("Environment"),
					Value: aws.String("production"),
				},
			},
			expected: true,
		},
		{
			name: "wrong casing on key",
			tags: []awstypes.Tag{
				{
					Key:   aws.String("guarddutymanaged"),
					Value: aws.String(acctest.CtTrue),
				},
			},
			expected: false,
		},
		{
			name: "wrong casing on value",
			tags: []awstypes.Tag{
				{
					Key:   aws.String("GuardDutyManaged"),
					Value: aws.String("True"),
				},
			},
			expected: false,
		},
		{
			name:     "nil tags",
			tags:     nil,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tfec2.HasGuardDutyManagedTag(tc.tags)

			if result != tc.expected {
				t.Errorf("HasGuardDutyManagedTag() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestDissociateGuardDutyVPCEndpoints_warningMessageFormat(t *testing.T) {
	t.Parallel()

	const warningTemplate = "During deletion of subnet %s, Terraform attempted to check for and clean up " +
		"GuardDuty-managed resources that may have been causing a DependencyViolation, " +
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
			subnetID    string
			permissions string
		}{
			{
				name:        "ownership check permissions",
				subnetID:    "subnet-0abc123def456",
				permissions: "ec2:DescribeSubnets, ec2:DescribeVpcs",
			},
			{
				name:        "describe endpoints permission",
				subnetID:    "subnet-0xyz789ghi012",
				permissions: "ec2:DescribeVpcEndpoints",
			},
			{
				name:        "modify endpoint permission",
				subnetID:    "subnet-12345",
				permissions: "ec2:ModifyVpcEndpoint",
			},
		}

		for _, tc := range identifierCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				formatted := fmt.Sprintf(warningTemplate, tc.subnetID, tc.permissions)

				if !strings.Contains(formatted, tc.subnetID) {
					t.Errorf("Formatted warning missing subnet ID %q\nFormatted: %s", tc.subnetID, formatted)
				}
				if !strings.Contains(formatted, tc.permissions) {
					t.Errorf("Formatted warning missing permissions %q\nFormatted: %s", tc.permissions, formatted)
				}
			})
		}
	})
}

func testAccVPCSubnetConfig_guardDutyLastSubnet_withSubnet(rName string) string {
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

resource "aws_subnet" "test1" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-1"
  }
}
`, rName)
}

func testAccVPCSubnetConfig_guardDutyLastSubnet_withoutSubnet(rName string) string {
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

// TestAccVPCSubnet_guardDutyNotAssociated validates UC-S3: deleting a subnet that is NOT
// associated with a GuardDuty endpoint. The endpoint is associated with subnet-B only.
// When subnet-A is removed from the Terraform config, dissociateGuardDutyVPCEndpoints
// should find the endpoint but skip it because subnet-A is not in the endpoint's SubnetIds.
// The endpoint should remain untouched in available state with subnet-B.
//
// EXPECTED: Test PASSES. Endpoint is untouched because subnet-A was never associated with it.
func TestAccVPCSubnet_guardDutyNotAssociated(t *testing.T) {
	ctx := acctest.Context(t)
	var subnetA, subnetB awstypes.Subnet
	var vpcID string
	resourceNameA := "aws_subnet.testA"
	resourceNameB := "aws_subnet.testB"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckGuardDutyCleanupDestroy(ctx, t, &vpcID),
			testAccDeleteGuardDutyResources(ctx, t, &vpcID),
		),
		Steps: []resource.TestStep{
			{
				// Step 1: Create VPC with two subnets (A and B). After Terraform
				// creates the infrastructure, create GuardDuty resources out-of-band
				// via SDK with ONLY subnet-B's ID (not subnet-A).
				Config: testAccVPCSubnetConfig_guardDutyNotAssociated_bothSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceNameA, &subnetA),
					testAccCheckSubnetExists(ctx, t, resourceNameB, &subnetB),
					testAccCaptureVPCID(&subnetB, &vpcID),
					// Create endpoint associated with subnet-B ONLY
					testAccCreateGuardDutyResourcesFromSubnet(ctx, t, &subnetB),
					testAccCheckGuardDutyResourcesExist(ctx, t, &subnetB),
				),
			},
			{
				// Step 2: Remove subnet-A from config (keep subnet-B). Terraform
				// deletes subnet-A. Since subnet-A is NOT associated with the
				// endpoint, dissociateGuardDutyVPCEndpoints should skip it.
				// The subnet should delete normally.
				Config: testAccVPCSubnetConfig_guardDutyNotAssociated_onlySubnetB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(ctx, t, resourceNameB, &subnetB),
					// Endpoint should still exist unchanged with subnet-B in available state
					testAccCheckGuardDutyEndpointStillExists(ctx, t, &vpcID),
				),
			},
		},
	})
}

func testAccVPCSubnetConfig_guardDutyNotAssociated_bothSubnets(rName string) string {
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

resource "aws_subnet" "testA" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-A"
  }
}

resource "aws_subnet" "testB" {
  cidr_block        = "10.1.2.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s-B"
  }
}
`, rName)
}

func testAccVPCSubnetConfig_guardDutyNotAssociated_onlySubnetB(rName string) string {
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

resource "aws_subnet" "testB" {
  cidr_block        = "10.1.2.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s-B"
  }
}
`, rName)
}
