// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNATGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttrSet(resourceName, "allocation_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeZonal)),
					resource.TestCheckResourceAttr(resourceName, "availability_zone_address.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "connectivity_type", "public"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
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

func TestAccVPCNATGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNATGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNATGateway_ConnectivityType_private(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_connectivityType(rName, "private"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "allocation_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrAssociationID, ""),
					resource.TestCheckResourceAttr(resourceName, "connectivity_type", "private"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "0"),
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

func TestAccVPCNATGateway_privateIP(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_privateIP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "allocation_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrAssociationID, ""),
					resource.TestCheckResourceAttr(resourceName, "connectivity_type", "private"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "10.0.0.8"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "0"),
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

func TestAccVPCNATGateway_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
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
				Config: testAccVPCNATGatewayConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccVPCNATGatewayConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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

func TestAccVPCNATGateway_secondaryAllocationIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	eipResourceName := "aws_eip.secondary"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secondary_allocation_ids.*", eipResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "1"),
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
				Config: testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secondary_allocation_ids.*", eipResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "1"),
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

func TestAccVPCNATGateway_addSecondaryAllocationIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	eipResourceName := "aws_eip.secondary"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secondary_allocation_ids.*", eipResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "0"),
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

func TestAccVPCNATGateway_secondaryPrivateIPAddressCount(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	secondaryPrivateIpAddressCount := 3

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddressCount(rName, secondaryPrivateIpAddressCount),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", strconv.Itoa(secondaryPrivateIpAddressCount)),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", strconv.Itoa(secondaryPrivateIpAddressCount)),
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

func TestAccVPCNATGateway_secondaryPrivateIPAddressCountToSpecific(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	secondaryPrivateIpAddressCount := 3

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddressCount(rName, secondaryPrivateIpAddressCount),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", strconv.Itoa(secondaryPrivateIpAddressCount)),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", strconv.Itoa(secondaryPrivateIpAddressCount)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses_private(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.7"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
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

func TestAccVPCNATGateway_secondaryPrivateIPAddresses(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	eipResourceName := "aws_eip.secondary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secondary_allocation_ids.*", eipResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
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
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secondary_allocation_ids.*", eipResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
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

func TestAccVPCNATGateway_SecondaryPrivateIPAddresses_private(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses_private(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.9"),
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
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses_private(rName, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.9"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.10"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.11"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses_private(rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "4"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.8"),
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

func TestAccVPCNATGateway_availabilityModeRegionalAuto(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalAuto(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", names.AttrEnabled),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
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

func TestAccVPCNATGateway_availabilityModeRegionalManual_AddAndRemove(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// key: Availability Zone index
	// value: Elastic IP indexes
	testSets := []map[int][]int{
		{
			0: {0},
			1: {1, 2},
		},
		// Add one more EIP
		{
			0: {0, 3},
			1: {1, 2},
		},
		// Remove one EIP each from different AZs simultaneously
		{
			0: {0},
			1: {1},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName, 4, testSets[0]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[0]),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"availability_zone_address.0.availability_zone_id",
					"availability_zone_address.1.availability_zone_id",
				},
			},
			{
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName, 4, testSets[1]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[1]),
				),
			},
			{
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName, 4, testSets[2]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[2]),
				),
			},
		},
	})
}

func TestAccVPCNATGateway_availabilityModeRegionalManual_ReplaceAndRemove(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// key: Availability Zone index
	// value: Elastic IP indexes
	testSets := []map[int][]int{
		{
			0: {0},
			1: {1},
		},
		// Replace EIP in one AZ
		{
			0: {2},
			1: {1},
		},
		// Remove EIP in one AZ
		{
			0: {2},
		},
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName, 3, testSets[0]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[0]),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"availability_zone_address.0.availability_zone_id",
					"availability_zone_address.1.availability_zone_id",
				},
			},
			{
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName, 3, testSets[1]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[1]),
				),
			},
			{
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName, 3, testSets[2]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[2]),
				),
			},
		},
	})
}

func TestAccVPCNATGateway_availabilityModeRegionalManual_AZNameToAZID(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// key: Availability Zone index
	// value: Elastic IP indexes
	testSets := []map[int][]int{
		{
			0: {0},
		},
		// Add one EIP in a new AZ and remove one in the existing AZ
		// availability_zone_id is used instead of availability_zone
		{
			1: {1},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Create with availability_zone specified
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName, 2, testSets[0]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[0]),
				),
			},
			{
				// Update with availability_zone_id specified
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManualByAZID(rName, 2, testSets[1]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
				),
			},
			{
				// Refresh state to store availability_zone in state
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[1]),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"availability_zone_address.0.availability_zone",
				},
			},
		},
	})
}

func TestAccVPCNATGateway_availabilityModeRegionalManual_AZIDToAZName(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// key: Availability Zone index
	// value: Elastic IP indexes
	testSets := []map[int][]int{
		// availability_zone_id is used
		{
			0: {0},
		},
		// Add one EIP in a new AZ and remove one in the existing AZ.
		// availability_zone is used instead of availability_zone_id
		{
			1: {1},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Create a resource with availability_zone_id specified
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManualByAZID(rName, 2, testSets[0]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
				),
			},
			{
				// Refresh state to store availability_zone in state
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[0]),
				),
			},
			{
				// Update the resource with availability_zone specified instead of availability_zone_id
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName, 2, testSets[1]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSets[1]),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"availability_zone_address.0.availability_zone_id",
				},
			},
		},
	})
}

func TestAccVPCNATGateway_availabilityModeRegionalSwitchMode(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// key: Availability Zone index
	// value: Elastic IP indexes
	testSet := map[int][]int{
		0: {0},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Create a resource in auto mode
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalAuto(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", names.AttrEnabled),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
				),
			},
			{
				// Switch to manual mode
				// Resource recreation is expected
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName, 1, testSet),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", "disabled"),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
					testAccCheckNATGatewayRegionalManualModeWithConfig(ctx, resourceName, testSet),
				),
			},
			{
				// Switch back to auto mode
				// Resource recreation is expected
				Config: testAccVPCNATGatewayConfig_availabilityModeRegionalAuto(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "availability_mode", string(awstypes.AvailabilityModeRegional)),
					resource.TestCheckResourceAttr(resourceName, "auto_provision_zones", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_ips", names.AttrEnabled),
					resource.TestCheckResourceAttrSet(resourceName, "route_table_id"),
				),
			},
		},
	})
}

func testAccCheckNATGatewayDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_nat_gateway" {
				continue
			}

			_, err := tfec2.FindNATGatewayByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 NAT Gateway %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNATGatewayExists(ctx context.Context, n string, v *awstypes.NatGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindNATGatewayByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckNATGatewayRegionalManualModeWithConfig(_ context.Context, n string, expectedConfig map[int][]int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		// Get availability zone data source
		azData, ok := s.RootModule().Resources["data.aws_availability_zones.available"]
		if !ok {
			return fmt.Errorf("availability zones data source not found")
		}

		// Build expected AZ to EIP Allocation ID mapping
		azToAllocationIDs := make(map[string][]string)
		expectedRegionalNATGatewayAddressCount := 0
		for azIndex, eipIndexes := range expectedConfig {
			azName := azData.Primary.Attributes[fmt.Sprintf("names.%d", azIndex)]
			var allocationIDs []string
			for _, eipIndex := range eipIndexes {
				eipResource, ok := s.RootModule().Resources[fmt.Sprintf("aws_eip.test.%d", eipIndex)]
				if !ok {
					return fmt.Errorf("eip resource aws_eip.test.%d not found", eipIndex)
				}
				allocationIDs = append(allocationIDs, eipResource.Primary.ID)
				expectedRegionalNATGatewayAddressCount++
			}
			azToAllocationIDs[azName] = allocationIDs
		}

		// Validate availability_zone_address blocks
		availableZoneAddressCountStr, exists := rs.Primary.Attributes["availability_zone_address.#"]
		if !exists {
			return fmt.Errorf("availability_zone_address block not found")
		}
		availableZoneAddressCount, err := strconv.Atoi(availableZoneAddressCountStr)
		if err != nil {
			return fmt.Errorf("invalid availability_zone_address count: %s", availableZoneAddressCountStr)
		}
		if availableZoneAddressCount != len(expectedConfig) {
			return fmt.Errorf("expected %d availability_zone_address blocks, got %d", len(expectedConfig), availableZoneAddressCount)
		}
		for i := range availableZoneAddressCount {
			az, azExists := rs.Primary.Attributes[fmt.Sprintf("availability_zone_address.%d.availability_zone", i)]
			allocationIDsCountStr, allocationIDsExists := rs.Primary.Attributes[fmt.Sprintf("availability_zone_address.%d.allocation_ids.#", i)]

			if !azExists || !allocationIDsExists {
				return fmt.Errorf("incomplete availability_zone_address block at index %d", i)
			}

			allocationIDsCount, err := strconv.Atoi(allocationIDsCountStr)
			if err != nil {
				return fmt.Errorf("invalid allocation count: %s", allocationIDsCountStr)
			}

			expectedAllocationIDs, hasExpected := azToAllocationIDs[az]
			if !hasExpected {
				return fmt.Errorf("unexpected AZ %s found in availability_zone_address", az)
			}
			actualAllocationIDs := make([]string, 0, allocationIDsCount)
			for j := range allocationIDsCount {
				allocationID := rs.Primary.Attributes[fmt.Sprintf("availability_zone_address.%d.allocation_ids.%d", i, j)]
				actualAllocationIDs = append(actualAllocationIDs, allocationID)
			}
			slices.Sort(actualAllocationIDs)
			slices.Sort(expectedAllocationIDs)
			if !cmp.Equal(expectedAllocationIDs, actualAllocationIDs) {
				return fmt.Errorf("allocation IDs mismatch in AZ %s. Expected: %v, Actual: %v, Index: %d", az, expectedAllocationIDs, actualAllocationIDs, i)
			}
		}

		// Validate regional_nat_gateway_address blocks
		regionalNATGatewayAddressCountStr, exists := rs.Primary.Attributes["regional_nat_gateway_address.#"]
		if !exists {
			return fmt.Errorf("regional_nat_gateway_address block not found")
		}
		regionalNATGatewayAddressCount, err := strconv.Atoi(regionalNATGatewayAddressCountStr)
		if err != nil {
			return fmt.Errorf("invalid regional_nat_gateway_address count: %s", regionalNATGatewayAddressCountStr)
		}
		if regionalNATGatewayAddressCount != expectedRegionalNATGatewayAddressCount {
			return fmt.Errorf("expected %d regional_nat_gateway_address blocks, got %d", expectedRegionalNATGatewayAddressCount, regionalNATGatewayAddressCount)
		}
		for i := range regionalNATGatewayAddressCount {
			az, azExists := rs.Primary.Attributes[fmt.Sprintf("regional_nat_gateway_address.%d.availability_zone", i)]

			if !azExists {
				return fmt.Errorf("incomplete regional_nat_gateway_address block at index %d", i)
			}

			expectedAllocationIDs, hasExpected := azToAllocationIDs[az]

			if !hasExpected {
				return fmt.Errorf("unexpected AZ %s found in regional_nat_gateway_address", az)
			}

			actualAllocationID := rs.Primary.Attributes[fmt.Sprintf("regional_nat_gateway_address.%d.allocation_id", i)]

			if !slices.Contains(expectedAllocationIDs, actualAllocationID) {
				return fmt.Errorf("expected allocation ID %s not found in AZ %s", actualAllocationID, az)
			}

			if v, ok := rs.Primary.Attributes[fmt.Sprintf("regional_nat_gateway_address.%d.availability_zone_id", i)]; !ok || v == "" {
				return fmt.Errorf("availability_zone_id not set for regional_nat_gateway_address in AZ %s", az)
			}
			if v, ok := rs.Primary.Attributes[fmt.Sprintf("regional_nat_gateway_address.%d.association_id", i)]; !ok || v == "" {
				return fmt.Errorf("association_id not set for regional_nat_gateway_address in AZ %s", az)
			}
			if v, ok := rs.Primary.Attributes[fmt.Sprintf("regional_nat_gateway_address.%d.public_ip", i)]; !ok || v == "" {
				return fmt.Errorf("public_ip not set for regional_nat_gateway_address in AZ %s", az)
			}
			if v, ok := rs.Primary.Attributes[fmt.Sprintf("regional_nat_gateway_address.%d.status", i)]; !ok || v != "succeeded" {
				return fmt.Errorf("status not set to succeeded for regional_nat_gateway_address in AZ %s", az)
			}
		}
		return nil
	}
}

func testAccNATGatewayConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "private" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.2.0/24"
  map_public_ip_on_launch = true

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

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNATGatewayConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), `
resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  depends_on = [aws_internet_gateway.test]
}
`)
}

func testAccVPCNATGatewayConfig_connectivityType(rName, connectivityType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  connectivity_type = %[2]q
  subnet_id         = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, connectivityType))
}

func testAccVPCNATGatewayConfig_privateIP(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  connectivity_type = "private"
  private_ip        = "10.0.0.8"
  subnet_id         = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNATGatewayConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    %[1]q = %[2]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, tagKey1, tagValue1))
}

func testAccVPCNATGatewayConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName string, hasSecondary bool) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_eip" "secondary" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id            = aws_eip.test.id
  subnet_id                = aws_subnet.public.id
  secondary_allocation_ids = %[2]t ? [aws_eip.secondary.id] : []

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, hasSecondary))
}

func testAccVPCNATGatewayConfig_secondaryPrivateIPAddressCount(rName string, secondaryPrivateIpAddressCount int) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  connectivity_type                  = "private"
  subnet_id                          = aws_subnet.private.id
  secondary_private_ip_address_count = %[2]d

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, secondaryPrivateIpAddressCount))
}

func testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses(rName string, hasSecondary bool) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_eip" "secondary" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id                  = aws_eip.test.id
  subnet_id                      = aws_subnet.private.id
  secondary_allocation_ids       = %[2]t ? [aws_eip.secondary.id] : []
  secondary_private_ip_addresses = %[2]t ? ["10.0.1.5"] : []

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, hasSecondary))
}

func testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses_private(rName string, n int) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  connectivity_type              = "private"
  subnet_id                      = aws_subnet.private.id
  secondary_private_ip_addresses = [for n in range(%[2]d) : "10.0.1.${5 + n}"]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, n))
}

func testAccVPCNATGatewayConfig_availabilityModeRegionalBase(rName string, eipCount int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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

resource "aws_eip" "test" {
  count  = %[2]d
  domain = "vpc"

  tags = {
    Name = "%[1]s-${count.index}"
  }
}
`, rName, eipCount))
}

func testAccVPCNATGatewayConfig_availabilityModeRegionalAuto(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCNATGatewayConfig_availabilityModeRegionalBase(rName, 0),
		fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  vpc_id            = aws_vpc.test.id
  availability_mode = "regional"

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

`, rName))
}

func testAccVPCNATGatewayConfig_availabilityModeRegionalManual(rName string, eipCount int, config map[int][]int) string {
	localsStr := "locals {\n  config = [\n"
	for azIndex, eipIndexes := range config {
		localsStr += fmt.Sprintf("    {az = %d, eip = [%s]},\n", azIndex, strings.Trim(strings.Replace(fmt.Sprint(eipIndexes), " ", ", ", -1), "[]"))
	}
	localsStr += "  ]\n}\n\n"

	return acctest.ConfigCompose(
		testAccVPCNATGatewayConfig_availabilityModeRegionalBase(rName, eipCount),
		fmt.Sprintf(`
%[2]s
resource "aws_nat_gateway" "test" {
  vpc_id            = aws_vpc.test.id
  availability_mode = "regional"

  dynamic "availability_zone_address" {
    for_each = local.config
    content {
      allocation_ids    = [for i in availability_zone_address.value.eip : aws_eip.test[i].id]
      availability_zone = data.aws_availability_zones.available.names[availability_zone_address.value.az]
    }
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

`, rName, localsStr))
}

func testAccVPCNATGatewayConfig_availabilityModeRegionalManualByAZID(rName string, eipCount int, config map[int][]int) string {
	localsStr := "locals {\n  config = [\n"
	for azIndex, eipIndexes := range config {
		localsStr += fmt.Sprintf("    {az = %d, eip = [%s]},\n", azIndex, strings.Trim(strings.Replace(fmt.Sprint(eipIndexes), " ", ", ", -1), "[]"))
	}
	localsStr += "  ]\n}\n\n"

	return acctest.ConfigCompose(
		testAccVPCNATGatewayConfig_availabilityModeRegionalBase(rName, eipCount),
		fmt.Sprintf(`
%[2]s
resource "aws_nat_gateway" "test" {
  vpc_id            = aws_vpc.test.id
  availability_mode = "regional"

  dynamic "availability_zone_address" {
    for_each = local.config
    content {
      allocation_ids       = [for i in availability_zone_address.value.eip : aws_eip.test[i].id]
      availability_zone_id = data.aws_availability_zones.available.zone_ids[availability_zone_address.value.az]
    }
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

`, rName, localsStr))
}
