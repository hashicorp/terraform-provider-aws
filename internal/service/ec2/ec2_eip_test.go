package ec2_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2EIP_basic(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "domain", "vpc"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					testAccCheckEIPPublicDNS(resourceName),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceEIP(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EIP_tags(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckVPCOnly(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
				Config: testAccEIPConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEIPConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2EIP_instance(t *testing.T) {
	var conf ec2.Address
	instanceResourceName := "aws_instance.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
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
	var conf ec2.Address
	instanceResourceName := "aws_instance.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_instanceReassociate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, "id"),
				),
			},
			{
				Config: testAccEIPConfig_instanceReassociate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
	var conf ec2.Address
	instance1ResourceName := "aws_instance.test.1"
	instance2ResourceName := "aws_instance.test.0"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_instanceAssociated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instance1ResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_with_private_ip"},
			},
			{
				Config: testAccEIPConfig_instanceAssociatedSwitch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instance2ResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIP_Instance_notAssociated(t *testing.T) {
	var conf ec2.Address
	instanceResourceName := "aws_instance.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_instanceAssociateNotAssociated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "association_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance", ""),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEIPConfig_instance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance", instanceResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIP_networkInterface(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_networkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_multiNetworkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &one),
					testAccCheckEIPAttributes(&one),
					testAccCheckEIPAssociated(&one),
					resource.TestCheckResourceAttr(resourceName, "domain", ec2.DomainTypeVpc),

					testAccCheckEIPExists(resourceName2, &two),
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

func TestAccEC2EIP_PublicIPv4Pool_default(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_publicIPv4PoolDefault,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_publicIPv4PoolCustom(poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_customerOwnedIPv4Pool(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_networkBorderGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckWavelengthZoneAvailable(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_carrierIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_byoipAddressCustomDefault,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_byoipAddressCustom(address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig_byoipAddressCustomPublicIPv4Pool(address, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(resourceName, &conf),
					testAccCheckEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ip", address),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", poolName),
				),
			},
		},
	})
}

func testAccCheckEIPExists(n string, v *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 EIP ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		var err error
		var output *ec2.Address

		if strings.Contains(rs.Primary.ID, "eipalloc") {
			output, err = tfec2.FindEIPByAllocationID(conn, rs.Primary.ID)
		} else {
			output, err = tfec2.FindEIPByPublicIP(conn, rs.Primary.ID)
		}

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEIPDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eip" {
			continue
		}

		var err error

		if strings.Contains(rs.Primary.ID, "eipalloc") {
			_, err = tfec2.FindEIPByAllocationID(conn, rs.Primary.ID)
		} else {
			_, err = tfec2.FindEIPByPublicIP(conn, rs.Primary.ID)
		}

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 EIP %s still exists", rs.Primary.ID)
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

const testAccEIPConfig_basic = `
resource "aws_eip" "test" {
  vpc = true
}
`

func testAccEIPConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccEIPConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

const testAccEIPConfig_publicIPv4PoolDefault = `
resource "aws_eip" "test" {
  vpc = true
}
`

func testAccEIPConfig_publicIPv4PoolCustom(poolName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc              = true
  public_ipv4_pool = %[1]q
}
`, poolName)
}

const testAccEIPConfig_byoipAddressCustomDefault = `
resource "aws_eip" "test" {
  vpc = true
}
`

func testAccEIPConfig_byoipAddressCustom(address string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc     = true
  address = %[1]q
}
`, address)
}

func testAccEIPConfig_byoipAddressCustomPublicIPv4Pool(address string, poolname string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc              = true
  address          = %[1]q
  public_ipv4_pool = %[2]q
}
`, address, poolname)
}

func testAccEIPConfig_baseInstance(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instance(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseInstance(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  instance = aws_instance.test.id
  vpc      = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instanceReassociate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
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
  subnet_id                   = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

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
  subnet_id      = aws_subnet.test[0].id
  route_table_id = aws_route_table.test.id
}
`, rName))
}

func testAccEIPConfig_baseInstanceAssociated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

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
  vpc_id                  = aws_vpc.test.id
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  count = 2

  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  private_ip = "10.0.0.1${count.index}"
  subnet_id  = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instanceAssociated(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseInstanceAssociated(rName),
		fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  instance                  = aws_instance.test[1].id
  associate_with_private_ip = aws_instance.test[1].private_ip

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instanceAssociatedSwitch(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseInstanceAssociated(rName),
		fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  instance                  = aws_instance.test[0].id
  associate_with_private_ip = aws_instance.test[0].private_ip

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_instanceAssociateNotAssociated(rName string) string {
	return acctest.ConfigCompose(testAccEIPConfig_baseInstance(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEIPConfig_networkInterface(rName string) string {
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

func testAccEIPConfig_multiNetworkInterface(rName string) string {
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

func testAccEIPConfig_customerOwnedIPv4Pool() string {
	return `
data "aws_ec2_coip_pools" "test" {}

resource "aws_eip" "test" {
  customer_owned_ipv4_pool = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  vpc                      = true
}
`
}

const testAccEIPConfig_networkBorderGroup = `
data "aws_region" current {}

resource "aws_eip" "test" {
  vpc                  = true
  network_border_group = data.aws_region.current.name
}
`

func testAccEIPConfig_carrierIP(rName string) string {
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
