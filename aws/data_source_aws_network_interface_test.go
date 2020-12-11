package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsNetworkInterface_basic(t *testing.T) {
	datasourceName := "data.aws_network_interface.test"
	resourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkInterfaceConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_ip", resourceName, "private_ip"),
					resource.TestCheckResourceAttrSet(datasourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrSet(datasourceName, "interface_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name", resourceName, "private_dns_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttr(datasourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttrSet(datasourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNetworkInterface_filters(t *testing.T) {
	datasourceName := "data.aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkInterfaceConfigFilters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "security_groups.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNetworkInterface_EIPAssociation(t *testing.T) {
	datasourceName := "data.aws_network_interface.test"
	resourceName := "aws_network_interface.test"
	eipResourceName := "aws_eip.test"
	eipAssociationResourceName := "aws_eip_association.test"
	securityGroupResourceName := "aws_security_group.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkInterfaceConfigEIPAssociation(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "association.#", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "association.0.allocation_id", eipResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "association.0.association_id", eipAssociationResourceName, "id"),
					resource.TestCheckResourceAttr(datasourceName, "association.0.carrier_ip", ""),
					testAccCheckResourceAttrAccountID(datasourceName, "association.0.ip_owner_id"),
					// EIP does not really get a public DNS name until it's associated with an instance.
					//resource.TestCheckResourceAttrPair(datasourceName, "association.0.public_dns_name", eipResourceName, "public_dns"),
					resource.TestCheckResourceAttr(datasourceName, "association.0.public_dns_name", ""),
					resource.TestCheckResourceAttrPair(datasourceName, "association.0.public_ip", eipResourceName, "public_ip"),
					resource.TestCheckResourceAttr(datasourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttrSet(datasourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttr(datasourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(datasourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(datasourceName, "mac_address"),
					resource.TestCheckResourceAttr(datasourceName, "outpost_arn", ""),
					testAccCheckResourceAttrAccountID(datasourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name", resourceName, "private_dns_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_ip", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_ips.#", resourceName, "private_ips.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_ips.0", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "security_groups.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNetworkInterfaceConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["10.0.0.50"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccDataSourceAwsNetworkInterfaceConfigBasic(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		testAccDataSourceAwsNetworkInterfaceConfigBase(rName),
		`
data "aws_network_interface" "test" {
  id = aws_network_interface.test.id
}
`)
}

func testAccDataSourceAwsNetworkInterfaceConfigEIPAssociation(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		testAccDataSourceAwsNetworkInterfaceConfigBase(rName),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip_association" "test" {
  allocation_id        = aws_eip.test.id
  network_interface_id = aws_network_interface.test.id
}

data "aws_network_interface" "test" {
  id = aws_eip_association.test.network_interface_id
}
`, rName))
}

func testAccDataSourceAwsNetworkInterfaceConfigFilters(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		testAccDataSourceAwsNetworkInterfaceConfigBase(rName),
		`
data "aws_network_interface" "test" {
  filter {
    name   = "network-interface-id"
    values = [aws_network_interface.test.id]
  }
}
`)
}
