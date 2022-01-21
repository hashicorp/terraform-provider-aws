package ec2_test

import (
	"fmt"
	"regexp"
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

func TestAccEC2NetworkInterface_basic(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	subnetResourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIConfig(rName),
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

func TestAccEC2NetworkInterface_ipv6(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIIPV6Config(rName),
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
				Config: testAccENIIPV6MultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "2"),
				),
			},
			{
				Config: testAccENIIPV6Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_addresses.#", "1"),
				),
			},
		},
	})
}

func TestAccEC2NetworkInterface_tags(t *testing.T) {
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var conf ec2.NetworkInterface

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENITags1Config(rName, "key1", "value1"),
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
				Config: testAccENITags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccENITags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2NetworkInterface_ipv6Count(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIIPV6CountConfig(rName, 1),
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
				Config: testAccENIIPV6CountConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "2"),
				),
			},
			{
				Config: testAccENIIPV6CountConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "0"),
				),
			},
			{
				Config: testAccENIIPV6CountConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
				),
			},
		},
	})
}

func TestAccEC2NetworkInterface_disappears(t *testing.T) {
	var networkInterface ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkInterface(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2NetworkInterface_description(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	subnetResourceName := "aws_subnet.test"
	securityGroupResourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIDescriptionConfig(rName, "description 1"),
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
				Config: testAccENIDescriptionConfig(rName, "description 2"),
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

func TestAccEC2NetworkInterface_attachment(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIAttachmentConfig(rName),
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

func TestAccEC2NetworkInterface_ignoreExternalAttachment(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIExternalAttachmentConfig(rName),
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

func TestAccEC2NetworkInterface_sourceDestCheck(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENISourceDestCheckConfig(rName, false),
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
				Config: testAccENISourceDestCheckConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "true"),
				),
			},
			{
				Config: testAccENISourceDestCheckConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "false"),
				),
			},
		},
	})
}

func TestAccEC2NetworkInterface_privateIPsCount(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIPrivateIPsCountConfig(rName, 1),
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
				Config: testAccENIPrivateIPsCountConfig(rName, 2),
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
				Config: testAccENIPrivateIPsCountConfig(rName, 0),
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
				Config: testAccENIPrivateIPsCountConfig(rName, 1),
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

func TestAccEC2NetworkInterface_ENIInterfaceType_efa(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIInterfaceTypeConfig(rName, "efa"),
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

func TestAccEC2NetworkInterface_ENI_ipv4Prefix(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIIPV4PrefixConfig(rName),
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
				Config: testAccENIIPV4PrefixMultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", "2"),
				),
			},
			{
				Config: testAccENIIPV4PrefixConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefixes.#", "1"),
				),
			},
		},
	})
}

func TestAccEC2NetworkInterface_ENI_ipv4PrefixCount(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIIPV4PrefixCountConfig(rName, 1),
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
				Config: testAccENIIPV4PrefixCountConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "2"),
				),
			},
			{
				Config: testAccENIIPV4PrefixCountConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "0"),
				),
			},
			{
				Config: testAccENIIPV4PrefixCountConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv4_prefix_count", "1"),
				),
			},
		},
	})
}

func TestAccEC2NetworkInterface_ENI_ipv6Prefix(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIIPV6PrefixConfig(rName),
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
				Config: testAccENIIPV6PrefixMultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", "2"),
				),
			},
			{
				Config: testAccENIIPV6PrefixConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefixes.#", "1"),
				),
			},
		},
	})
}

func TestAccEC2NetworkInterface_ENI_ipv6PrefixCount(t *testing.T) {
	var conf ec2.NetworkInterface
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccENIIPV6PrefixCountConfig(rName, 1),
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
				Config: testAccENIIPV6PrefixCountConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "2"),
				),
			},
			{
				Config: testAccENIIPV6PrefixCountConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "0"),
				),
			},
			{
				Config: testAccENIIPV6PrefixCountConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ipv6_prefix_count", "1"),
				),
			},
		},
	})
}

type privateIpListTestConfigData struct {
	private_ips             []string
	private_ips_count       string
	private_ip_list_enabled string
	private_ip_list         []string
	replacesInterface       bool
}

func TestAccAWSENI_PrivateIpsSet(t *testing.T) {
	var networkInterface, lastInterface ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	testConfigs := []privateIpListTestConfigData{

		{[]string{"44", "59", "123"}, "", "", []string{}, false},       // Configuration with three private_ips
		{[]string{"123", "44", "59"}, "", "", []string{}, false},       // Change order of private_ips
		{[]string{"123", "12", "59", "44"}, "", "", []string{}, false}, // Add secondaries to private_ips
		{[]string{"123", "59", "44"}, "", "", []string{}, false},       // Remove secondaries from private_ips
		{[]string{"123", "59", "57"}, "", "", []string{}, true},        // Remove primary
		{[]string{}, "4", "", []string{}, false},                       // Use count to add IPs
		{[]string{"44", "57"}, "", "", []string{}, false},              // Change list, retain primary
		{[]string{"44", "57", "123", "12"}, "", "", []string{}, false}, // Add to secondaries
		{[]string{"17"}, "", "", []string{}, true},                     // New list
		{[]string{"17", "45", "89"}, "", "", []string{}, false},        // Add secondaries
	}

	testSteps := make([]resource.TestStep, len(testConfigs)*2)
	testSteps[0] = resource.TestStep{
		Config: testAccAWSENIConfigPrivateIpList(testConfigs[0], resourceName),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckENIExists(resourceName, &networkInterface),
			testAccCheckAWSENIPrivateIpList(testConfigs[0], &networkInterface),
			testAccCheckENIExists(resourceName, &lastInterface),
		),
	}
	testSteps[1] = resource.TestStep{
		ResourceName:            resourceName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
	}

	for i, testConfig := range testConfigs {
		if i == 0 {
			continue
		}
		if testConfig.replacesInterface {
			testSteps[i*2] = resource.TestStep{
				Config: testAccAWSENIConfigPrivateIpList(testConfigs[i], resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckAWSENIPrivateIpList(testConfigs[i], &networkInterface),
					testAccCheckAWSENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			}
		} else {
			testSteps[i*2] = resource.TestStep{
				Config: testAccAWSENIConfigPrivateIpList(testConfigs[i], resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckAWSENIPrivateIpList(testConfigs[i], &networkInterface),
					testAccCheckAWSENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			}
		}
		// import check
		testSteps[i*2+1] = testSteps[1]
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps:        testSteps,
	})
}

func TestAccAWSENI_PrivateIpList(t *testing.T) {
	var networkInterface, lastInterface ec2.NetworkInterface
	resourceName := "aws_network_interface.test"

	// private_ips, private_ips_count, private_ip_list_enabed, private_ip_list, replacesInterface
	testConfigs := []privateIpListTestConfigData{
		{[]string{"17"}, "", "", []string{}, true},                               // Build a set incrementally in order
		{[]string{"17", "45"}, "", "", []string{}, false},                        //   Add to set
		{[]string{"17", "45", "89"}, "", "", []string{}, false},                  //   Add to set
		{[]string{"17", "45", "89", "122"}, "", "", []string{}, false},           //   Add to set
		{[]string{}, "", "true", []string{"17", "45", "89", "122"}, false},       // Change from set to list using same order
		{[]string{}, "", "true", []string{"17", "89", "45", "122"}, false},       // Change order of private_ip_list
		{[]string{}, "", "true", []string{"17", "89", "45"}, false},              // Remove secondaries from end
		{[]string{}, "", "true", []string{"17", "89", "45", "123"}, false},       // Add secondaries to end
		{[]string{}, "", "true", []string{"17", "89", "77", "45", "123"}, false}, // Add secondaries to middle
		{[]string{}, "", "true", []string{"17", "89", "123"}, false},             // Remove secondaries from middle
		{[]string{}, "4", "", []string{}, false},                                 // Use count to add IPs
		{[]string{}, "", "true", []string{"59", "123", "44"}, true},              // Change to specific list - forces new
		{[]string{}, "", "true", []string{"123", "59", "44"}, true},              // Change first of private_ip_list - forces new
		{[]string{"123", "59", "44"}, "", "", []string{}, false},                 // Change from list to set using same set
	}

	testSteps := make([]resource.TestStep, len(testConfigs)*2)
	testSteps[0] = resource.TestStep{
		Config: testAccAWSENIConfigPrivateIpList(testConfigs[0], resourceName),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckENIExists(resourceName, &networkInterface),
			testAccCheckAWSENIPrivateIpList(testConfigs[0], &networkInterface),
			testAccCheckENIExists(resourceName, &lastInterface),
		),
	}
	testSteps[1] = resource.TestStep{
		ResourceName:            resourceName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"private_ip_list_enabled", "ipv6_address_list_enabled"},
	}

	for i, testConfig := range testConfigs {
		if i == 0 {
			continue
		}
		if testConfig.replacesInterface {
			testSteps[i*2] = resource.TestStep{
				Config: testAccAWSENIConfigPrivateIpList(testConfigs[i], resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckAWSENIPrivateIpList(testConfigs[i], &networkInterface),
					testAccCheckAWSENIDifferent(&lastInterface, &networkInterface), // different
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			}
		} else {
			testSteps[i*2] = resource.TestStep{
				Config: testAccAWSENIConfigPrivateIpList(testConfigs[i], resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckENIExists(resourceName, &networkInterface),
					testAccCheckAWSENIPrivateIpList(testConfigs[i], &networkInterface),
					testAccCheckAWSENISame(&lastInterface, &networkInterface), // same
					testAccCheckENIExists(resourceName, &lastInterface),
				),
			}
		}
		// import check
		testSteps[i*2+1] = testSteps[1]
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckENIDestroy,
		Steps:        testSteps,
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

func testAccCheckAWSENIPrivateIpList(testConfig privateIpListTestConfigData, iface *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		havePrivateIps := tfec2.FlattenNetworkInterfacePrivateIpAddresses(iface.PrivateIpAddresses)
	PRIVATE_IPS_LOOP:
		// every IP from private_ips should be present on the interface
		for _, needIp := range testConfig.private_ips {
			for _, haveIp := range havePrivateIps {
				if haveIp == "172.16.10."+needIp {
					continue PRIVATE_IPS_LOOP
				}
			}
			return fmt.Errorf("expected ip 172.16.10.%s to be in interface set %s", needIp, strings.Join(havePrivateIps, ","))
		}
		// every configured IP should be present on the interface
		for needIdx, needIp := range testConfig.private_ip_list {
			if len(havePrivateIps) <= needIdx || "172.16.10."+needIp != havePrivateIps[needIdx] {
				return fmt.Errorf("expected ip 172.16.10.%s to be at %d in the list %s", needIp, needIdx, strings.Join(havePrivateIps, ","))
			}
		}
		// number of ips configured should match interface
		if len(testConfig.private_ips) > 0 && len(testConfig.private_ips) != len(havePrivateIps) {
			return fmt.Errorf("expected %s got %s", strings.Join(testConfig.private_ips, ","), strings.Join(havePrivateIps, ","))
		}
		if len(testConfig.private_ip_list) > 0 && len(testConfig.private_ip_list) != len(havePrivateIps) {
			return fmt.Errorf("expected %s got %s", strings.Join(testConfig.private_ip_list, ","), strings.Join(havePrivateIps, ","))
		}
		return nil
	}
}

func testAccCheckAWSENISame(iface1 *ec2.NetworkInterface, iface2 *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(iface1.NetworkInterfaceId) != aws.StringValue(iface2.NetworkInterfaceId) {
			return fmt.Errorf("Interface %s should not have been replaced with %s", aws.StringValue(iface1.NetworkInterfaceId), aws.StringValue(iface2.NetworkInterfaceId))
		}
		return nil
	}
}

func testAccCheckAWSENIDifferent(iface1 *ec2.NetworkInterface, iface2 *ec2.NetworkInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(iface1.NetworkInterfaceId) == aws.StringValue(iface2.NetworkInterfaceId) {
			return fmt.Errorf("Interface %s should have been replaced, have %s", aws.StringValue(iface1.NetworkInterfaceId), aws.StringValue(iface2.NetworkInterfaceId))
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

func testAccENIConfig(rName string) string {
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName), `
resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}
`)
}

func testAccENIIPV6Config(rName string) string {
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

func testAccENIIPV6MultipleConfig(rName string) string {
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

func testAccENIIPV6CountConfig(rName string, ipv6Count int) string {
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName) + fmt.Sprintf(`
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

func testAccENIDescriptionConfig(rName, description string) string {
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

func testAccENISourceDestCheckConfig(rName string, sourceDestCheck bool) string {
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName) + fmt.Sprintf(`
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

func testAccENIAttachmentConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
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

func testAccENIExternalAttachmentConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
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

func testAccENIPrivateIPsCountConfig(rName string, privateIpsCount int) string {
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

func testAccENITags1Config(rName, tagKey1, tagValue1 string) string {
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

func testAccENITags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccENIInterfaceTypeConfig(rName, interfaceType string) string {
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

func testAccENIIPV4PrefixConfig(rName string) string {
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

func testAccENIIPV4PrefixMultipleConfig(rName string) string {
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

func testAccENIIPV4PrefixCountConfig(rName string, ipv4PrefixCount int) string {
	return acctest.ConfigCompose(testAccENIIPV4BaseConfig(rName) + fmt.Sprintf(`
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

func testAccENIIPV6PrefixConfig(rName string) string {
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

func testAccENIIPV6PrefixMultipleConfig(rName string) string {
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

func testAccENIIPV6PrefixCountConfig(rName string, ipv6PrefixCount int) string {
	return acctest.ConfigCompose(testAccENIIPV6BaseConfig(rName) + fmt.Sprintf(`
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

func testAccAWSENIConfigPrivateIpList(testConfig privateIpListTestConfigData, rName string) string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf(`
%s "aws_network_interface" "test" {
  subnet_id          = aws_subnet.test.id
  security_groups    = [aws_security_group.test.id]
  description        = "Managed by Terraform"
`, "resource"))

	if len(testConfig.private_ips) > 0 {
		config.WriteString("  private_ips = [\n")
		for _, ip := range testConfig.private_ips {
			config.WriteString(fmt.Sprintf("  \"172.16.10.%s\",\n", ip))
		}
		config.WriteString("]\n")
	}

	if testConfig.private_ips_count != "" {
		config.WriteString(fmt.Sprintf("  private_ips_count = %s\n", testConfig.private_ips_count))
	}

	if testConfig.private_ip_list_enabled != "" {
		config.WriteString(fmt.Sprintf("  private_ip_list_enabled = %s\n", testConfig.private_ip_list_enabled))
	}
	config.WriteString("  ipv6_address_list_enabled = false\n")

	if len(testConfig.private_ip_list) > 0 {
		config.WriteString("  private_ip_list = [\n")
		for _, ip := range testConfig.private_ip_list {
			config.WriteString(fmt.Sprintf("  \"172.16.10.%s\",\n", ip))
		}
		config.WriteString("]\n")
	}

	config.WriteString("}\n")
	return testAccENIIPV6BaseConfig(rName) + config.String()
}
