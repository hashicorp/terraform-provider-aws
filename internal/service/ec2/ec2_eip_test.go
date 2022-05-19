package ec2_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// This will currently skip EIPs with associations,
// although we depend on aws_vpc to potentially have
// the majority of those associations removed.

func TestAccEC2EIP_basic(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					testAccCheckEIPPublicDNS(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
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

func TestAccEC2EIP_disappears(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceEIP(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EIP_instance(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPInstanceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
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

// Regression test for https://github.com/hashicorp/terraform/issues/3429 (now
// https://github.com/hashicorp/terraform-provider-aws/issues/42)
func TestAccEC2EIP_Instance_reassociate(t *testing.T) {
	instanceResourceName := "aws_instance.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPInstanceReassociateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, "id"),
				),
			},
			{
				Config: testAccEIPInstanceReassociateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, "id"),
				),
				Taint: []string{resourceName},
			},
		},
	})
}

// This test is an expansion of TestAccEC2EIP_Instance_associatedUserPrivateIP, by testing the
// associated Private EIPs of two instances
func TestAccEC2EIP_Instance_associatedUserPrivateIP(t *testing.T) {
	var one ec2.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPInstanceAssociatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &one),
					testAccCheckEIPAttributes(&one),
					testAccCheckEIPAssociated(&one),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_with_private_ip"},
			},
			{
				Config: testAccEIPInstanceAssociatedSwitchConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &one),
					testAccCheckEIPAttributes(&one),
					testAccCheckEIPAssociated(&one),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
				),
			},
		},
	})
}

func TestAccEC2EIP_Instance_notAssociated(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPInstanceAssociateNotAssociatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEIPInstanceAssociateAssociatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					testAccCheckEIPAssociated(&conf),
				),
			},
		},
	})
}

func TestAccEC2EIP_Instance_ec2Classic(t *testing.T) {
	resourceName := "aws_eip.test"
	var conf ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_instanceClassic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, true, &conf),
					testAccCheckEIPAttributes(&conf),
					testAccCheckEIPPublicDNSClassic(resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeStandard),
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

func TestAccEC2EIP_networkInterface(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPNetworkInterfaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					testAccCheckEIPAssociated(&conf),
					testAccCheckEIPPrivateDNS(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "allocation_id"),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
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

func TestAccEC2EIP_NetworkInterface_twoEIPsOneInterface(t *testing.T) {
	var one, two ec2.Address
	resourceName := "aws_eip.test"
	resourceName2 := "aws_eip.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPMultiNetworkInterfaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &one),
					testAccCheckEIPAttributes(&one),
					testAccCheckEIPAssociated(&one),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),

					testAccCheckEIPExists(resourceName2, false, &two),
					testAccCheckEIPAttributes(&two),
					testAccCheckEIPAssociated(&two),
					resource.TestCheckResourceAttr(resourceName2, "domain", ec2.DomainTypeVpc),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_with_private_ip"},
			},
		},
	})
}

func TestAccEC2EIP_TagsEC2VPC_withVPCTrue(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckVPCOnly(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_tagsVPC(rName, "vpc = true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEIPConfig_tagsVPC(rName2, "vpc = true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName2),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform-provider-aws/issues/18756
func TestAccEC2EIP_TagsEC2VPC_withoutVPCTrue(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckVPCOnly(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_tagsVPC(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEIPConfig_tagsVPC(rName2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName2),
				),
			},
		},
	})
}

func TestAccEC2EIP_TagsEC2Classic_withVPCTrue(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_tagsClassic(rName, "vpc = true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, true, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEIPConfig_tagsClassic(rName2, "vpc = true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, true, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", rName2),
				),
			},
		},
	})
}

func TestAccEC2EIP_TagsEC2Classic_withoutVPCTrue(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccEIPConfig_tagsClassic(rName, ""),
				ExpectError: regexp.MustCompile(`tags cannot be set for a standard-domain EIP - must be a VPC-domain EIP`),
			},
		},
	})
}

func TestAccEC2EIP_PublicIPv4Pool_default(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPPublicIPv4PoolDefaultConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", "amazon"),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),
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

func TestAccEC2EIP_PublicIPv4Pool_custom(t *testing.T) {
	if os.Getenv("AWS_EC2_EIP_PUBLIC_IPV4_POOL") == "" {
		t.Skip("Environment variable AWS_EC2_EIP_PUBLIC_IPV4_POOL is not set")
	}

	var conf ec2.Address
	resourceName := "aws_eip.test"

	poolName := os.Getenv("AWS_EC2_EIP_PUBLIC_IPV4_POOL")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPPublicIPv4PoolCustomConfig(poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", poolName),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeStandard),
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

func TestAccEC2EIP_customerOwnedIPv4Pool(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPCustomerOwnedIPv4PoolConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					resource.TestMatchResourceAttr(resourceName, "customer_owned_ipv4_pool", regexp.MustCompile(`^ipv4pool-coip-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "customer_owned_ip", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
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

func TestAccEC2EIP_networkBorderGroup(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPNetworkBorderGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", "amazon"),
					resource.TestCheckResourceAttr(resourceName, "network_border_group", acctest.Region()),
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

func TestAccEC2EIP_carrierIP(t *testing.T) {
	var conf ec2.Address
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWavelengthZoneAvailable(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPCarrierIPConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "carrier_ip"),
					resource.TestCheckResourceAttrSet(resourceName, "network_border_group"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
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

func TestAccEC2EIP_BYOIPAddress_default(t *testing.T) {
	// Test case address not set
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_BYOIPAddress_custom_default,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
				),
			},
		},
	})
}

func TestAccEC2EIP_BYOIPAddress_custom(t *testing.T) {
	// Test Case for address being set

	if os.Getenv("AWS_EC2_EIP_BYOIP_ADDRESS") == "" {
		t.Skip("Environment variable AWS_EC2_EIP_BYOIP_ADDRESS is not set")
	}

	var conf ec2.Address
	resourceName := "aws_eip.test"

	address := os.Getenv("AWS_EC2_EIP_BYOIP_ADDRESS")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_BYOIPAddress_custom(address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ip", address),
				),
			},
		},
	})
}

func TestAccEC2EIP_BYOIPAddress_customWithPublicIPv4Pool(t *testing.T) {
	// Test Case for both address and public_ipv4_pool being set
	if os.Getenv("AWS_EC2_EIP_BYOIP_ADDRESS") == "" {
		t.Skip("Environment variable AWS_EC2_EIP_BYOIP_ADDRESS is not set")
	}

	if os.Getenv("AWS_EC2_EIP_PUBLIC_IPV4_POOL") == "" {
		t.Skip("Environment variable AWS_EC2_EIP_PUBLIC_IPV4_POOL is not set")
	}

	var conf ec2.Address
	resourceName := "aws_eip.test"

	address := os.Getenv("AWS_EC2_EIP_BYOIP_ADDRESS")
	poolName := os.Getenv("AWS_EC2_EIP_PUBLIC_IPV4_POOL")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_BYOIPAddress_custom_with_PublicIPv4Pool(address, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, false, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ip", address),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", poolName),
				),
			},
		},
	})
}

func testAccCheckEIPDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eip" {
			continue
		}

		if strings.Contains(rs.Primary.ID, "eipalloc") {
			req := &ec2.DescribeAddressesInput{
				AllocationIds: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				// Verify the error is what we want
				if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidAllocationID.NotFound" || ae.Code() == "InvalidAddress.NotFound" {
					continue
				}
				return err
			}

			if len(describe.Addresses) > 0 {
				return fmt.Errorf("still exists")
			}
		} else {
			req := &ec2.DescribeAddressesInput{
				PublicIps: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				// Verify the error is what we want
				if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidAllocationID.NotFound" || ae.Code() == "InvalidAddress.NotFound" {
					continue
				}
				return err
			}

			if len(describe.Addresses) > 0 {
				return fmt.Errorf("still exists")
			}
		}
	}

	return nil
}

func testAccCheckEIPAttributes(conf *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.PublicIp == "" {
			return fmt.Errorf("empty public_ip")
		}

		return nil
	}
}

func testAccCheckEIPAssociated(conf *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.AssociationId == nil || *conf.AssociationId == "" {
			return fmt.Errorf("empty association_id")
		}

		return nil
	}
}

func testAccCheckEIPExists(n string, ec2classic bool, res *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		if ec2classic {
			conn = acctest.ProviderEC2Classic.Meta().(*conns.AWSClient).EC2Conn
		}

		input := &ec2.DescribeAddressesInput{}

		if strings.Contains(rs.Primary.ID, "eipalloc") {
			input.AllocationIds = aws.StringSlice([]string{rs.Primary.ID})
		} else {
			input.PublicIps = aws.StringSlice([]string{rs.Primary.ID})
		}

		var output *ec2.DescribeAddressesOutput

		err := resource.Retry(15*time.Minute, func() *resource.RetryError {
			var err error

			output, err = conn.DescribeAddresses(input)

			if tfawserr.ErrCodeEquals(err, "InvalidAllocationID.NotFound") {
				return resource.RetryableError(err)
			}

			if tfawserr.ErrCodeEquals(err, "InvalidAddress.NotFound") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			output, err = conn.DescribeAddresses(input)
		}

		if err != nil {
			return fmt.Errorf("while describing addresses (%s): %w", rs.Primary.ID, err)
		}

		if len(output.Addresses) != 1 {
			return fmt.Errorf("wrong number of EIP found for (%s): %d", rs.Primary.ID, len(output.Addresses))
		}

		if aws.StringValue(output.Addresses[0].AllocationId) != rs.Primary.ID && aws.StringValue(output.Addresses[0].PublicIp) != rs.Primary.ID {
			return fmt.Errorf("EIP (%s) not found", rs.Primary.ID)
		}

		*res = *output.Addresses[0]

		return nil
	}
}

func testAccCheckEIPPrivateDNS(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		privateDNS := rs.Primary.Attributes["private_dns"]
		expectedPrivateDNS := fmt.Sprintf(
			"ip-%s.%s",
			tfec2.ConvertIPToDashIP(rs.Primary.Attributes["private_ip"]),
			tfec2.RegionalPrivateDNSSuffix(acctest.Region()),
		)

		if privateDNS != expectedPrivateDNS {
			return fmt.Errorf("expected private_dns value (%s), received: %s", expectedPrivateDNS, privateDNS)
		}

		return nil
	}
}

func testAccCheckEIPPublicDNS(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		publicDNS := rs.Primary.Attributes["public_dns"]
		expectedPublicDNS := fmt.Sprintf(
			"ec2-%s.%s.%s",
			tfec2.ConvertIPToDashIP(rs.Primary.Attributes["public_ip"]),
			tfec2.RegionalPublicDNSSuffix(acctest.Region()),
			acctest.PartitionDNSSuffix(),
		)

		if publicDNS != expectedPublicDNS {
			return fmt.Errorf("expected public_dns value (%s), received: %s", expectedPublicDNS, publicDNS)
		}

		return nil
	}
}

func testAccCheckEIPPublicDNSClassic(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		publicDNS := rs.Primary.Attributes["public_dns"]
		expectedPublicDNS := fmt.Sprintf(
			"ec2-%s.%s.%s",
			tfec2.ConvertIPToDashIP(rs.Primary.Attributes["public_ip"]),
			tfec2.RegionalPublicDNSSuffix(acctest.EC2ClassicRegion()),
			acctest.PartitionDNSSuffix(),
		)

		if publicDNS != expectedPublicDNS {
			return fmt.Errorf("expected public_dns value (%s), received: %s", expectedPublicDNS, publicDNS)
		}

		return nil
	}
}

const testAccEIPConfig = `
resource "aws_eip" "test" {
  vpc = true
}
`

func testAccEIPConfig_tagsVPC(rName, vpcConfig string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  %[1]s

  tags = {
    RandomName = %[2]q
    TestName   = %[2]q
  }
}
`, vpcConfig, rName)
}

func testAccEIPConfig_tagsClassic(rName, vpcConfig string) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
resource "aws_eip" "test" {
  %[1]s

  tags = {
    RandomName = %[2]q
    TestName   = %[2]q
  }
}
`, vpcConfig, rName))
}

const testAccEIPPublicIPv4PoolDefaultConfig = `
resource "aws_eip" "test" {
  vpc = true
}
`

func testAccEIPPublicIPv4PoolCustomConfig(poolName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc              = true
  public_ipv4_pool = %[1]q
}
`, poolName)
}

func testAccEIPConfig_instanceClassic() string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		testAccLatestAmazonLinuxPVEBSAMIConfig(),
		acctest.AvailableEC2InstanceTypeForRegion("t1.micro", "m3.medium", "m3.large", "c3.large", "r3.large"),
		`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-pv-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_eip" "test" {
  instance = aws_instance.test.id
}
`)
}

const testAccEIPConfig_BYOIPAddress_custom_default = `
resource "aws_eip" "test" {
  vpc = true
}
`

func testAccEIPConfig_BYOIPAddress_custom(address string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc     = true
  address = %[1]q
}
`, address)
}

func testAccEIPConfig_BYOIPAddress_custom_with_PublicIPv4Pool(address string, poolname string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc              = true
  address          = %[1]q
  public_ipv4_pool = %[2]q
}
`, address, poolname)
}

func testAccEIPInstanceConfig() string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		acctest.ConfigAvailableAZsNoOptIn(),
		`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id
}

resource "aws_eip" "test" {
  instance = aws_instance.test.id
  vpc      = true
}
`)
}

func testAccEIPInstanceAssociatedConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.default.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.default.id
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = [aws_internet_gateway.gw]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  private_ip = "10.0.0.12"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test2" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  private_ip = "10.0.0.19"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  vpc = true

  instance                  = aws_instance.test2.id
  associate_with_private_ip = "10.0.0.19"
}
`, rName))
}

func testAccEIPInstanceAssociatedSwitchConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.default.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.default.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = [aws_internet_gateway.gw]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"

  private_ip = "10.0.0.12"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test2" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

  instance_type = "t2.micro"

  private_ip = "10.0.0.19"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_eip" "test" {
  vpc = true

  instance                  = aws_instance.test.id
  associate_with_private_ip = "10.0.0.12"
}
`, rName))
}

func testAccEIPNetworkInterfaceConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/24"
  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["10.0.0.10"]
  security_groups = [aws_vpc.test.default_security_group_id]
}

resource "aws_eip" "test" {
  vpc               = "true"
  network_interface = aws_network_interface.test.id
  depends_on        = [aws_internet_gateway.test]
}
`, rName))
}

func testAccEIPMultiNetworkInterfaceConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/24"
  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["10.0.0.10", "10.0.0.11"]
  security_groups = [aws_vpc.test.default_security_group_id]
}

resource "aws_eip" "test" {
  vpc                       = "true"
  network_interface         = aws_network_interface.test.id
  associate_with_private_ip = "10.0.0.10"
  depends_on                = [aws_internet_gateway.test]
}

resource "aws_eip" "test2" {
  vpc                       = "true"
  network_interface         = aws_network_interface.test.id
  associate_with_private_ip = "10.0.0.11"
  depends_on                = [aws_internet_gateway.test]
}
`, rName))
}

func testAccEIPInstanceReassociateConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_eip" "test" {
  instance = aws_instance.test.id
  vpc      = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  lifecycle {
    create_before_destroy = true
  }
}

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

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

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
`, rName))
}

func testAccEIPInstanceAssociateNotAssociatedConfig() string {
	return acctest.ConfigCompose(
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id
}

resource "aws_eip" "test" {
}
`)
}

func testAccEIPInstanceAssociateAssociatedConfig() string {
	return acctest.ConfigCompose(
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id
}

resource "aws_eip" "test" {
  instance = aws_instance.test.id
  vpc      = true
}
`)
}

func testAccEIPCustomerOwnedIPv4PoolConfig() string {
	return `
data "aws_ec2_coip_pools" "test" {}

resource "aws_eip" "test" {
  customer_owned_ipv4_pool = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  vpc                      = true
}
`
}

const testAccEIPNetworkBorderGroupConfig = `
data "aws_region" current {}

resource "aws_eip" "test" {
  vpc                  = true
  network_border_group = data.aws_region.current.name
}
`

func testAccEIPCarrierIPConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccAvailableAZsWavelengthZonesDefaultExcludeConfig(),
		fmt.Sprintf(`
data "aws_availability_zone" "available" {
  name = data.aws_availability_zones.available.names[0]
}

resource "aws_eip" "test" {
  vpc                  = true
  network_border_group = data.aws_availability_zone.available.network_border_group

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
