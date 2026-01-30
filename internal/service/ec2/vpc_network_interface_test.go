// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkInterface_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
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
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`network-interface/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					checkResourceAttrPrivateDNSName(resourceName, "private_dns_name", &conf.PrivateIpAddress),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ipv6Primary(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Primary(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ipv6PrimaryEnable(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Primary(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ipv6PrimaryDisable(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Primary(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Primary(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var conf awstypes.NetworkInterface

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
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ipv6Count(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Count(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Count(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var networkInterface awstypes.NetworkInterface
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
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceNetworkInterface(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkInterface_description(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 1"),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "private_ips.*", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_groups.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 2"),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "private_dns_name"),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "private_ips.*", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_groups.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attachment.*", map[string]string{
						"device_index":       "1",
						"network_card_index": "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
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

// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html#network-cards.
// This test requires an expensive instance type that supports multiple network cards, such as "c6in.32xlarge" or "c6in.metal".
// Set the environment variable `VPC_NETWORK_INTERFACE_TEST_MULTIPLE_NETWORK_CARDS` to run this test.
func TestAccVPCNetworkInterface_attachmentNetworkCardIndex(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, "VPC_NETWORK_INTERFACE_TEST_MULTIPLE_NETWORK_CARDS")
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_attachmentNetworkCardIndex(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attachment.*", map[string]string{
						"device_index":       "1",
						"network_card_index": "1",
					}),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "172.16.10.100"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
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
	var conf awstypes.NetworkInterface
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
	var conf awstypes.NetworkInterface
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
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "2"),
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
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "private_ips_count", "1"),
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
	var conf awstypes.NetworkInterface
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
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv4PrefixCount(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "0"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv6Prefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv6PrefixCount(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
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
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "0"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_privateIPSet(t *testing.T) {
	ctx := acctest.Context(t)
	var networkInterface, lastInterface awstypes.NetworkInterface
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
	var networkInterface, lastInterface awstypes.NetworkInterface
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

// To support ABAC (attribute based access control) when deleting network interfaces, the provider
// attempts to verify that the interface has an attachment before attempting to detach
//
// This test expects the TF_ACC_ASSUME_ROLE_ARN to be set to a role ARN that has a principal "Owner" tag
// that matches the "Owner" tag set on the resources.
//
// Use the following configuration to create the role to assume:
/*
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = "assume-role-ec2"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = ["sts:AssumeRole", "sts:SetSourceIdentity"],
      Principal = {
        AWS = data.aws_caller_identity.current.arn,
      }
      Effect = "Allow"
      Sid    = ""
    }]
  })

  tags = {
    Owner = "terraform-provider-aws"
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "ec2:DetachNetworkInterface",
      "ec2:DeleteNetworkInterface",
      "ec2:TerminateInstances",
      "ec2:DeleteVolume",
      "ec2:DetachVolume"
    ]
    effect    = "Allow"
    resources = ["*"]

    condition {
      test     = "StringEquals"
      values   = ["$${aws:PrincipalTag/Owner}"]
      variable = "aws:ResourceTag/Owner"
    }
  }

  # 2. Creation Actions (Allow creation freely so we can get to the destroy step)
  statement {
    actions = [
      "ec2:RunInstances",
      "ec2:CreateNetworkInterface",
      "ec2:CreateVolume",
      "ec2:CreateTags",
      "ec2:CreateSecurityGroup",
      "ec2:Describe*",
      # Add network creation permissions for THIS workspace to work
      "ec2:CreateVpc",
      "ec2:DeleteVpc",
      "ec2:CreateSubnet",
      "ec2:DeleteSubnet"
    ]
    effect    = "Allow"
    resources = ["*"]
  }
}

resource "aws_iam_policy" "test" {
  name   = "tfc-reproduce-abac-issue"
  policy = data.aws_iam_policy_document.test.json

  tags = {
    Owner = "terraform-provider-aws"
  }
}

resource "aws_iam_role_policy_attachment" "tfc_policy_attachment" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}
*/
//	Once provisioned, use the role_arn output and run this test as follows:
//
//	TF_ACC_ASSUME_ROLE_ARN=<output> make t K=ec2 T=TestAccVPCNetworkInterface_deleteWithLimitedAssumeRole
func TestAccVPCNetworkInterface_deleteWithLimitedAssumeRole(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAssumeRoleARN(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckENIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigAssumeRole(),
					testAccVPCNetworkInterfaceConfig_deleteWithLimitedAssumeRole_attached(rName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`network-interface/.+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled", "attachment"},
			},
			{
				// test that deleting the network interface does not produce an error while detaching
				Config: acctest.ConfigCompose(
					acctest.ConfigAssumeRole(),
					testAccVPCNetworkInterfaceConfig_deleteWithLimitedAssumeRole_removed(rName),
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
	if region == endpoints.UsEast1RegionID {
		return "ec2.internal"
	}

	return fmt.Sprintf("%s.compute.internal", region)
}

func testAccCheckENIExists(ctx context.Context, n string, v *awstypes.NetworkInterface) resource.TestCheckFunc {
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

			if retry.NotFound(err) {
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

func testAccCheckENIMakeExternalAttachment(ctx context.Context, n string, networkInterface *awstypes.NetworkInterface, attachmentId *string) resource.TestCheckFunc {
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

func testAccCheckENIPrivateIPSet(ips []string, iface *awstypes.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iIPs := tfec2.FlattenNetworkInterfacePrivateIPAddresses(iface.PrivateIpAddresses)

		if !stringSlicesEqualIgnoreOrder(ips, iIPs) {
			return fmt.Errorf("expected private IP set %s, got %s", strings.Join(ips, ","), strings.Join(iIPs, ","))
		}

		return nil
	}
}

func testAccCheckENIPrivateIPList(ips []string, iface *awstypes.NetworkInterface) resource.TestCheckFunc {
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

	slices.Sort(s1)
	slices.Sort(s2)

	return reflect.DeepEqual(s1, s2)
}

func stringSlicesEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	return reflect.DeepEqual(s1, s2)
}

func testAccCheckENISame(iface1 *awstypes.NetworkInterface, iface2 *awstypes.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(iface1.NetworkInterfaceId) != aws.ToString(iface2.NetworkInterfaceId) {
			return fmt.Errorf("interface %s should not have been replaced with %s", aws.ToString(iface1.NetworkInterfaceId), aws.ToString(iface2.NetworkInterfaceId))
		}
		return nil
	}
}

func testAccCheckENIDifferent(iface1 *awstypes.NetworkInterface, iface2 *awstypes.NetworkInterface) resource.TestCheckFunc {
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

func testAccVPCNetworkInterfaceConfig_ipv6Primary(rName string, enable bool) string {
	return acctest.ConfigCompose(testAccVPCNetworkInterfaceConfig_baseIPV6(rName), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id           = aws_subnet.test.id
  private_ips         = ["172.16.10.100"]
  enable_primary_ipv6 = %[1]t
  ipv6_addresses      = [cidrhost(aws_subnet.test.ipv6_cidr_block, 4)]
  security_groups     = [aws_security_group.test.id]

  tags = {
    Name = %[2]q
  }
}
`, enable, rName))
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

func testAccVPCNetworkInterfaceConfig_attachmentNetworkCardIndex(rName string, networkCardIndex int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("c6in.32xlarge", "c6in.metal"),
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
    instance           = aws_instance.test.id
    device_index       = 1
    network_card_index = %[2]d
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, networkCardIndex))
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

func testAccVPCNetworkInterfaceConfig_deleteWithLimitedAssumeRole_attached(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name  = %[1]q
    Owner = "terraform-provider-aws"
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = false

  tags = {
    Name  = %[1]q
    Owner = "terraform-provider-aws"
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  primary_network_interface {
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name  = %[1]q
    Owner = "terraform-provider-aws"
  }
}

resource "aws_network_interface" "test" {
  subnet_id   = aws_subnet.test.id
  private_ips = ["10.1.1.42"]

  tags = {
    Name  = %[1]q
    Owner = "terraform-provider-aws"
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceConfig_deleteWithLimitedAssumeRole_removed(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name  = %[1]q
    Owner = "terraform-provider-aws"
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = false

  tags = {
    Name  = %[1]q
    Owner = "terraform-provider-aws"
  }
}
`, rName))
}
