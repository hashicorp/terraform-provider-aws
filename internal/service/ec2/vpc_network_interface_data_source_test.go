package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCNetworkInterfaceDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_network_interface.test"
	resourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceDataSourceConfig_basic(rName),
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
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_id", resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterfaceDataSource_filters(t *testing.T) {
	datasourceName := "data.aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceDataSourceConfig_filters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "security_groups.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCNetworkInterfaceDataSource_carrierIPAssociation(t *testing.T) {
	datasourceName := "data.aws_network_interface.test"
	resourceName := "aws_network_interface.test"
	eipResourceName := "aws_eip.test"
	eipAssociationResourceName := "aws_eip_association.test"
	securityGroupResourceName := "aws_security_group.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWavelengthZoneAvailable(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceDataSourceConfig_carrierIPAssociation(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "association.#", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "association.0.allocation_id", eipResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "association.0.association_id", eipAssociationResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "association.0.carrier_ip", eipResourceName, "carrier_ip"),
					resource.TestCheckResourceAttr(datasourceName, "association.0.customer_owned_ip", ""),
					acctest.CheckResourceAttrAccountID(datasourceName, "association.0.ip_owner_id"),
					resource.TestCheckResourceAttr(datasourceName, "association.0.public_dns_name", ""),
					resource.TestCheckResourceAttr(datasourceName, "association.0.public_ip", ""),
					resource.TestCheckResourceAttr(datasourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttrSet(datasourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttr(datasourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(datasourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(datasourceName, "mac_address"),
					resource.TestCheckResourceAttr(datasourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner_id"),
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

func TestAccVPCNetworkInterfaceDataSource_publicIPAssociation(t *testing.T) {
	datasourceName := "data.aws_network_interface.test"
	resourceName := "aws_network_interface.test"
	eipResourceName := "aws_eip.test"
	eipAssociationResourceName := "aws_eip_association.test"
	securityGroupResourceName := "aws_security_group.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceDataSourceConfig_publicIPAssociation(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "association.#", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "association.0.allocation_id", eipResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "association.0.association_id", eipAssociationResourceName, "id"),
					resource.TestCheckResourceAttr(datasourceName, "association.0.carrier_ip", ""),
					resource.TestCheckResourceAttr(datasourceName, "association.0.customer_owned_ip", ""),
					acctest.CheckResourceAttrAccountID(datasourceName, "association.0.ip_owner_id"),
					// Public DNS name is not set by the EC2 API.
					resource.TestCheckResourceAttr(datasourceName, "association.0.public_dns_name", ""),
					resource.TestCheckResourceAttrPair(datasourceName, "association.0.public_ip", eipResourceName, "public_ip"),
					resource.TestCheckResourceAttr(datasourceName, "attachment.#", "0"),
					resource.TestCheckResourceAttrSet(datasourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttr(datasourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(datasourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(datasourceName, "mac_address"),
					resource.TestCheckResourceAttr(datasourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner_id"),
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

func TestAccVPCNetworkInterfaceDataSource_attachment(t *testing.T) {
	datasourceName := "data.aws_network_interface.test"
	resourceName := "aws_network_interface.test"
	instanceResourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInterfaceDataSourceConfig_attachment(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "association.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "attachment.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "attachment.0.device_index", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "attachment.0.instance_id", instanceResourceName, "id"),
					acctest.CheckResourceAttrAccountID(datasourceName, "attachment.0.instance_owner_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttr(datasourceName, "interface_type", "interface"),
					resource.TestCheckResourceAttr(datasourceName, "ipv6_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(datasourceName, "mac_address"),
					resource.TestCheckResourceAttr(datasourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(datasourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_dns_name", resourceName, "private_dns_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_ip", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_ips.#", resourceName, "private_ips.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "private_ips.0", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccNetworkInterfaceBaseDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["10.0.0.50"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNetworkInterfaceDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccNetworkInterfaceBaseDataSourceConfig(rName),
		`
data "aws_network_interface" "test" {
  id = aws_network_interface.test.id
}
`)
}

func testAccVPCNetworkInterfaceDataSourceConfig_carrierIPAssociation(rName string) string {
	return acctest.ConfigCompose(
		testAccAvailableAZsWavelengthZonesDefaultExcludeConfig(),
		fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["10.0.0.50"]
  security_groups = [aws_security_group.test.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_carrier_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

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

resource "aws_eip_association" "test" {
  allocation_id        = aws_eip.test.id
  network_interface_id = aws_network_interface.test.id
}

data "aws_network_interface" "test" {
  id = aws_eip_association.test.network_interface_id
}
`, rName))
}

func testAccVPCNetworkInterfaceDataSourceConfig_publicIPAssociation(rName string) string {
	return acctest.ConfigCompose(
		testAccNetworkInterfaceBaseDataSourceConfig(rName),
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

func testAccVPCNetworkInterfaceDataSourceConfig_filters(rName string) string {
	return acctest.ConfigCompose(
		testAccNetworkInterfaceBaseDataSourceConfig(rName),
		`
data "aws_network_interface" "test" {
  filter {
    name   = "network-interface-id"
    values = [aws_network_interface.test.id]
  }
}
`)
}

func testAccVPCNetworkInterfaceDataSourceConfig_attachment(rName string) string {
	return acctest.ConfigCompose(
		testAccNetworkInterfaceBaseDataSourceConfig(rName),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface_attachment" "test" {
  device_index         = 1
  instance_id          = aws_instance.test.id
  network_interface_id = aws_network_interface.test.id
}

data "aws_network_interface" "test" {
  id = aws_network_interface_attachment.test.network_interface_id
}
`, rName))
}
