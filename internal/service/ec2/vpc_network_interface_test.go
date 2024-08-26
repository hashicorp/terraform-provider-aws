// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkInterface_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	subnetResourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`network-interface/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					checkResourceAttrPrivateDNSName(resourceName, "private_dns_name", &conf.PrivateIpAddress),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
		},
	})
}

func TestAccVPCNetworkInterface_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var conf types.NetworkInterface

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ipv6Count(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Count(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Count(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Count(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct0),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Count(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var networkInterface types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNetworkInterface(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkInterface_description(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	subnetResourceName := "aws_subnet.test"
	securityGroupResourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_description(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 1"),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "private_ips.*", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_groups.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_description(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 2"),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "private_ips.*", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_groups.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_attachment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_attachment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attachment.*", map[string]string{
						"device_index": acctest.Ct1,
					}),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "private_ips.*", "172.16.10.100"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
		},
	})
}

func TestAccVPCNetworkInterface_ignoreExternalAttachment(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	var attachmentId string
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_externalAttachment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					testAccCheckENIMakeExternalAttachment(ctx, "aws_instance.test", &conf, &attachmentId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"attachment",
					"ipv6_address_list_enabled",
					"private_ip_list_enabled",
				},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_externalAttachment(rName),
				Check: resource.ComposeTestCheckFunc(
					// Detach the external network interface attachment for the post-destroy to be able to the delete network interface
					testAccCheckENIRemoveExternalAttachment(ctx, &attachmentId),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_sourceDestCheck(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_sourceDestCheck(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_sourceDestCheck(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtTrue),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_sourceDestCheck(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_privateIPsCount(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_privateIPsCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_privateIPsCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_privateIPsCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_privateIPsCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENIInterfaceType_efa(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_type(rName, "efa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "efa"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv4Prefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv4PrefixCount(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", acctest.Ct0),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv6Prefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv6PrefixCount(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", acctest.Ct0),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_privateIPSet(t *testing.T) {
	ctx := acctest.Context(t)
	var networkInterface, lastInterface types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{ // Configuration with three private_ips
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.44", "172.16.10.59", "172.16.10.123"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{ // Change order of private_ips
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.123", "172.16.10.44", "172.16.10.59"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Add secondaries to private_ips
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.123", "172.16.10.12", "172.16.10.44", "172.16.10.59"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.12", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Remove secondary to private_ips
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.123", "172.16.10.44", "172.16.10.59"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Remove primary
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.123", "172.16.10.59", "172.16.10.57"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.57", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Use count to add IPs
				Config: testAccVPCNetworkInterfaceConfig_privateIPSetCount(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Change list, retain primary
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.44", "172.16.10.57"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.57"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // New list
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.17"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17"}, &networkInterface),
					testAccCheckENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_privateIPList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var networkInterface, lastInterface types.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{ // Build a set incrementally in order
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.17"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17"}, &networkInterface),
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{ // Add to set
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.17", "172.16.10.45"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17", "172.16.10.45"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Add to set
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.17", "172.16.10.45", "172.16.10.89"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17", "172.16.10.45", "172.16.10.89"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Add to set
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.17", "172.16.10.45", "172.16.10.89", "172.16.10.122"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17", "172.16.10.45", "172.16.10.89", "172.16.10.122"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Change from set to list using same order
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.45", "172.16.10.89", "172.16.10.122"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.45", "172.16.10.89", "172.16.10.122"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Change order of private_ip_list
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.45", "172.16.10.122"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.45", "172.16.10.122"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Remove secondaries from end
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.45"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.45"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Add secondaries to end
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.45", "172.16.10.123"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.45", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Add secondaries to middle
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.77", "172.16.10.45", "172.16.10.123"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.77", "172.16.10.45", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Remove secondaries from middle
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.123"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Use count to add IPs
				Config: testAccVPCNetworkInterfaceConfig_privateIPSetCount(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Change to specific list - forces new
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.59", "172.16.10.123", "172.16.10.38"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.59", "172.16.10.123", "172.16.10.38"}, &networkInterface),
					testAccCheckENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Change first of private_ip_list - forces new
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.123", "172.16.10.59", "172.16.10.38"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.123", "172.16.10.59", "172.16.10.38"}, &networkInterface),
					testAccCheckENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
			{ // Change from list to set using same set
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.123", "172.16.10.59", "172.16.10.38"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.123", "172.16.10.59", "172.16.10.38"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(ctx, resourceName, &lastInterface),
				),
			},
		},
	})
}

// checkResourceAttrPrivateDNSName ensures the Terraform state exactly matches a private DNS name
//
// For example: ip-172-16-10-100.us-west-2.compute.internal
func checkResourceAttrPrivateDNSName(resourceName, attributeName string, privateIpAddress **string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		privateDnsName := fmt.Sprintf("ip-%s.%s", convertIPToDashIP(**privateIpAddress), regionalPrivateDNSSuffix(acctest.Region()))

		return resource.TestCheckResourceAttr(resourceName, attributeName, privateDnsName)(s)
	}
}

func convertIPToDashIP(ip string) string {
	return strings.Replace(ip, ".", "-", -1)
}

func regionalPrivateDNSSuffix(region string) string {
	if region == names.USEast1RegionID {
		return "ec2.internal"
	}

	return fmt.Sprintf("%s.compute.internal", region)
}

func testAccCheckENIExists(ctx context.Context, n string, v *types.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Network Interface ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindNetworkInterfaceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckENIDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_network_interface" {
				continue
			}

			_, err := tfec2.FindNetworkInterfaceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Network Interface %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckENIMakeExternalAttachment(ctx context.Context, n string, networkInterface *types.NetworkInterface, attachmentId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("Not found: %s", n)
		}

		input := &ec2.AttachNetworkInterfaceInput{
			DeviceIndex:        aws.Int32(1),
			InstanceId:         aws.String(rs.Primary.ID),
			NetworkInterfaceId: networkInterface.NetworkInterfaceId,
		}

		output, err := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx).AttachNetworkInterface(ctx, input)
		*attachmentId = *output.AttachmentId

		if err != nil {
			return fmt.Errorf("error attaching ENI: %w", err)
		}
		return nil
	}
}

func testAccCheckENIRemoveExternalAttachment(ctx context.Context, attachmentId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		input := &ec2.DetachNetworkInterfaceInput{
			AttachmentId: attachmentId,
		}

		_, err := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx).DetachNetworkInterface(ctx, input)

		if err != nil {
			return fmt.Errorf("error detaching ENI: %w", err)
		}
		return nil
	}
}

func testAccCheckENIPrivateIPSet(ips []string, iface *types.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iIPs := tfec2.FlattenNetworkInterfacePrivateIPAddresses(iface.PrivateIpAddresses)

		if !stringSlicesEqualIgnoreOrder(ips, iIPs) {
			return fmt.Errorf("expected private IP set %s, got %s", strings.Join(ips, ","), strings.Join(iIPs, ","))
		}

		return nil
	}
}

func testAccCheckENIPrivateIPList(ips []string, iface *types.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iIPs := tfec2.FlattenNetworkInterfacePrivateIPAddresses(iface.PrivateIpAddresses)

		if !stringSlicesEqual(ips, iIPs) {
			return fmt.Errorf("expected private IP set %s, got %s", strings.Join(ips, ","), strings.Join(iIPs, ","))
		}

		return nil
	}
}

func stringSlicesEqualIgnoreOrder(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	sort.Strings(s1)
	sort.Strings(s2)

	return reflect.DeepEqual(s1, s2)
}

func stringSlicesEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	return reflect.DeepEqual(s1, s2)
}

func testAccCheckENISame(iface1 *types.NetworkInterface, iface2 *types.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(iface1.NetworkInterfaceId) != aws.ToString(iface2.NetworkInterfaceId) {
			return fmt.Errorf("interface %s should not have been replaced with %s", aws.ToString(iface1.NetworkInterfaceId), aws.ToString(iface2.NetworkInterfaceId))
		}
		return nil
	}
}

func testAccCheckENIDifferent(iface1 *types.NetworkInterface, iface2 *types.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(iface1.NetworkInterfaceId) == aws.ToString(iface2.NetworkInterfaceId) {
			return fmt.Errorf("interface %s should have been replaced, have %s", aws.ToString(iface1.NetworkInterfaceId), aws.ToString(iface2.NetworkInterfaceId))
		}
		return nil
	}
}

func testAccVPCNetworkInterfaceConfig_baseIPV4(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_baseIPV6(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "172.16.0.0/16"
  assign_generated_ipv6_cidr_block = true
  enable_dns_hostnames             = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.10.0/24"
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 16)
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV4(rName), `
resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}
`)
}

func testAccVPCNetworkInterfaceConfig_ipv6(rName string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV6(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  ipv6_addresses  = [cidrhost(aws_subnet.test.ipv6_cidr_block, 4)]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_ipv6Multiple(rName string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV6(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  ipv6_addresses  = [cidrhost(aws_subnet.test.ipv6_cidr_block, 4), cidrhost(aws_subnet.test.ipv6_cidr_block, 8)]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_ipv6Count(rName string, ipv6Count int) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV6(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id          = aws_subnet.test.id
  private_ips        = ["172.16.10.100"]
  ipv6_address_count = %[2]d
  security_groups    = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6Count))
}

func testAccVPCNetworkInterfaceConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV4(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]
  description     = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, description))
}

func testAccVPCNetworkInterfaceConfig_sourceDestCheck(rName string, sourceDestCheck bool) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV6(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id         = aws_subnet.test.id
  source_dest_check = %[2]t
  private_ips       = ["172.16.10.100"]

  tags = {
    Name = %[1]q
  }
}
`, rName, sourceDestCheck))
}

func testAccVPCNetworkInterfaceConfig_attachment(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		testAccVPCNetworkInterfaceConfig_baseIPV4(rName),
		fmt.Sprintf(`
resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.11.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test2.id
  associate_public_ip_address = false
  private_ip                  = "172.16.11.50"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  attachment {
    instance     = aws_instance.test.id
    device_index = 1
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_externalAttachment(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		testAccVPCNetworkInterfaceConfig_baseIPV4(rName),
		fmt.Sprintf(`
resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.11.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test2.id
  associate_public_ip_address = false
  private_ip                  = "172.16.11.50"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_privateIPsCount(rName string, privateIpsCount int) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV4(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  private_ips_count = %[2]d
  subnet_id         = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, privateIpsCount))
}

func testAccVPCNetworkInterfaceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV4(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccVPCNetworkInterfaceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV4(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCNetworkInterfaceConfig_type(rName, interfaceType string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV4(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.test.id]
  interface_type  = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, interfaceType))
}

func testAccVPCNetworkInterfaceConfig_ipv4Prefix(rName string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV4(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  ipv4_prefixes   = ["172.16.10.16/28"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_ipv4PrefixMultiple(rName string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV4(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  ipv4_prefixes   = ["172.16.10.16/28", "172.16.10.32/28"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName string, ipv4PrefixCount int) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV4(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id         = aws_subnet.test.id
  ipv4_prefix_count = %[2]d
  security_groups   = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv4PrefixCount))
}

func testAccVPCNetworkInterfaceConfig_ipv6Prefix(rName string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV6(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  ipv6_prefixes   = [cidrsubnet(aws_subnet.test.ipv6_cidr_block, 16, 2)]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_ipv6PrefixMultiple(rName string) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV6(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["172.16.10.100"]
  ipv6_prefixes   = [cidrsubnet(aws_subnet.test.ipv6_cidr_block, 16, 2), cidrsubnet(aws_subnet.test.ipv6_cidr_block, 16, 3)]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName string, ipv6PrefixCount int) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV6(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id         = aws_subnet.test.id
  private_ips       = ["172.16.10.100"]
  ipv6_prefix_count = %[2]d
  security_groups   = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6PrefixCount))
}

func testAccVPCNetworkInterfaceConfig_privateIPSet(rName string, privateIPs []string) string {
	return acctest.ConfigCompose(
		testAccVPCNetworkInterfaceConfig_baseIPV6(rName),
		fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  security_groups = [aws_security_group.test.id]
  private_ips     = ["%[1]s"]
}
`, strings.Join(privateIPs, `", "`)))
}

func testAccVPCNetworkInterfaceConfig_privateIPSetCount(rName string, count int) string {
	return acctest.ConfigCompose(
		testAccVPCNetworkInterfaceConfig_baseIPV6(rName),
		fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id         = aws_subnet.test.id
  security_groups   = [aws_security_group.test.id]
  private_ips_count = %[1]d
}
`, count))
}

func testAccVPCNetworkInterfaceConfig_privateIPList(rName string, privateIPs []string) string {
	return acctest.ConfigCompose(
		testAccVPCNetworkInterfaceConfig_baseIPV6(rName),
		fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id               = aws_subnet.test.id
  security_groups         = [aws_security_group.test.id]
  private_ip_list_enabled = true
  private_ip_list         = ["%[1]s"]
}
`, strings.Join(privateIPs, `", "`)))
}
