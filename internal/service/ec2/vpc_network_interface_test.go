package ec2_test

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCNetworkInterface_basic(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	subnetResourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`network-interface/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					checkResourceAttrPrivateDNSName(resourceName, "private_dns_name", &conf.PrivateIpAddress),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
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
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_tags(t *testing.T) {
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var conf ec2.NetworkInterface

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ipv6Count(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Count(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
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
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Count(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Count(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_disappears(t *testing.T) {
	var networkInterface ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkInterface(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkInterface_description(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	subnetResourceName := "aws_subnet.test"
	securityGroupResourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_description(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "description 1"),
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
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_groups.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "description 2"),
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
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_groups.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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

	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_attachment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "attachment.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attachment.*", map[string]string{
						"device_index": "1",
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
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_externalAttachment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					testAccCheckENIMakeExternalAttachment("aws_instance.test", &conf),
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

func TestAccVPCNetworkInterface_sourceDestCheck(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_sourceDestCheck(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "false"),
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
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_sourceDestCheck(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "false"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_privateIPsCount(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_privateIPsCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
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
					testAccCheckENIExists(resourceName, &conf),
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
					testAccCheckENIExists(resourceName, &conf),
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
					testAccCheckENIExists(resourceName, &conf),
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
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_type(rName, "efa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
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
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
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
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv4PrefixCount(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
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
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "0"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv4PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv6Prefix(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
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
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6Prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_ENI_ipv6PrefixCount(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
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
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "2"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "0"),
				),
			},
			{
				Config: testAccVPCNetworkInterfaceConfig_ipv6PrefixCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_privateIPSet(t *testing.T) {
	var networkInterface, lastInterface ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{ // Configuration with three private_ips
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.44", "172.16.10.59", "172.16.10.123"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENIExists(resourceName, &lastInterface),
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
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Add secondaries to private_ips
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.123", "172.16.10.12", "172.16.10.44", "172.16.10.59"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.12", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Remove secondary to private_ips
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.123", "172.16.10.44", "172.16.10.59"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Remove primary
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.123", "172.16.10.59", "172.16.10.57"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.57", "172.16.10.59", "172.16.10.123"}, &networkInterface),
					testAccCheckENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Use count to add IPs
				Config: testAccVPCNetworkInterfaceConfig_privateIPSetCount(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Change list, retain primary
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.44", "172.16.10.57"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.44", "172.16.10.57"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // New list
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.17"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17"}, &networkInterface),
					testAccCheckENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterface_privateIPList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var networkInterface, lastInterface ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{ // Build a set incrementally in order
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.17"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17"}, &networkInterface),
					testAccCheckENIExists(resourceName, &lastInterface),
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
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17", "172.16.10.45"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Add to set
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.17", "172.16.10.45", "172.16.10.89"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17", "172.16.10.45", "172.16.10.89"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Add to set
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.17", "172.16.10.45", "172.16.10.89", "172.16.10.122"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.17", "172.16.10.45", "172.16.10.89", "172.16.10.122"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Change from set to list using same order
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.45", "172.16.10.89", "172.16.10.122"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.45", "172.16.10.89", "172.16.10.122"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Change order of private_ip_list
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.45", "172.16.10.122"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.45", "172.16.10.122"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Remove secondaries from end
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.45"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.45"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Add secondaries to end
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.45", "172.16.10.123"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.45", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Add secondaries to middle
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.77", "172.16.10.45", "172.16.10.123"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.77", "172.16.10.45", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Remove secondaries from middle
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.17", "172.16.10.89", "172.16.10.123"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.17", "172.16.10.89", "172.16.10.123"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Use count to add IPs
				Config: testAccVPCNetworkInterfaceConfig_privateIPSetCount(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Change to specific list - forces new
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.59", "172.16.10.123", "172.16.10.38"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.59", "172.16.10.123", "172.16.10.38"}, &networkInterface),
					testAccCheckENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Change first of private_ip_list - forces new
				Config: testAccVPCNetworkInterfaceConfig_privateIPList(rName, []string{"172.16.10.123", "172.16.10.59", "172.16.10.38"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPList([]string{"172.16.10.123", "172.16.10.59", "172.16.10.38"}, &networkInterface),
					testAccCheckENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			},
			{ // Change from list to set using same set
				Config: testAccVPCNetworkInterfaceConfig_privateIPSet(rName, []string{"172.16.10.123", "172.16.10.59", "172.16.10.38"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckENIPrivateIPSet([]string{"172.16.10.123", "172.16.10.59", "172.16.10.38"}, &networkInterface),
					testAccCheckENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
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

func testAccCheckENIExists(n string, v *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Network Interface ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindNetworkInterfaceByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckENIDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_network_interface" {
			continue
		}

		_, err := tfec2.FindNetworkInterfaceByID(conn, rs.Primary.ID)

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

func testAccCheckENIMakeExternalAttachment(n string, conf *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("Not found: %s", n)
		}

		input := &ec2.AttachNetworkInterfaceInput{
			DeviceIndex:        aws.Int64(1),
			InstanceId:         aws.String(rs.Primary.ID),
			NetworkInterfaceId: conf.NetworkInterfaceId,
		}

		_, err := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn.AttachNetworkInterface(input)

		if err != nil {
			return fmt.Errorf("error attaching ENI: %w", err)
		}
		return nil
	}
}

func testAccCheckENIPrivateIPSet(ips []string, iface *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iIPs := tfec2.FlattenNetworkInterfacePrivateIPAddresses(iface.PrivateIpAddresses)

		if !stringSlicesEqualIgnoreOrder(ips, iIPs) {
			return fmt.Errorf("expected private IP set %s, got %s", strings.Join(ips, ","), strings.Join(iIPs, ","))
		}

		return nil
	}
}

func testAccCheckENIPrivateIPList(ips []string, iface *ec2.NetworkInterface) resource.TestCheckFunc {
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

func testAccCheckENISame(iface1 *ec2.NetworkInterface, iface2 *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(iface1.NetworkInterfaceId) != aws.StringValue(iface2.NetworkInterfaceId) {
			return fmt.Errorf("interface %s should not have been replaced with %s", aws.StringValue(iface1.NetworkInterfaceId), aws.StringValue(iface2.NetworkInterfaceId))
		}
		return nil
	}
}

func testAccCheckENIDifferent(iface1 *ec2.NetworkInterface, iface2 *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(iface1.NetworkInterfaceId) == aws.StringValue(iface2.NetworkInterfaceId) {
			return fmt.Errorf("interface %s should have been replaced, have %s", aws.StringValue(iface1.NetworkInterfaceId), aws.StringValue(iface2.NetworkInterfaceId))
		}
		return nil
	}
}

func testAccENIIPV4BaseConfig(rName string) string {
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

func testAccENIIPV6BaseConfig(rName string) string {
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
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), `
resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}
`)
}

func testAccVPCNetworkInterfaceConfig_ipv6(rName string) string {
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName), fmt.Sprintf(`
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		testAccENIIPV4BaseConfig(rName),
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
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		testAccENIIPV4BaseConfig(rName),
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
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName), fmt.Sprintf(`
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
		testAccENIIPV6BaseConfig(rName),
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
		testAccENIIPV6BaseConfig(rName),
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
		testAccENIIPV6BaseConfig(rName),
		fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id               = aws_subnet.test.id
  security_groups         = [aws_security_group.test.id]
  private_ip_list_enabled = true
  private_ip_list         = ["%[1]s"]
}
`, strings.Join(privateIPs, `", "`)))
}
