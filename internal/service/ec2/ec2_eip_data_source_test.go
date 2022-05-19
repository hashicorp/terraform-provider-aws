package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2EIPDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPFilterDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_id(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPIDDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_PublicIP_ec2Classic(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPPublicIPClassicDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_PublicIP_vpc(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPPublicIPVPCDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain", resourceName, "domain"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_tags(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPTagsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_dns", resourceName, "public_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_networkInterface(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPNetworkInterfaceDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interface_id", resourceName, "network_interface"),
					resource.TestCheckResourceAttrPair(dataSourceName, "private_dns", resourceName, "private_dns"),
					resource.TestCheckResourceAttrPair(dataSourceName, "private_ip", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain", resourceName, "domain"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_instance(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPInstanceDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_id", resourceName, "instance"),
					resource.TestCheckResourceAttrPair(dataSourceName, "association_id", resourceName, "association_id"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_carrierIP(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWavelengthZoneAvailable(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPCarrierIPDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "carrier_ip", resourceName, "carrier_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccEC2EIPDataSource_customerOwnedIPv4Pool(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPCustomerOwnedIPv4PoolDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "customer_owned_ipv4_pool", dataSourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_owned_ip", dataSourceName, "customer_owned_ip"),
				),
			},
		},
	})
}

func testAccEIPCustomerOwnedIPv4PoolDataSourceConfig() string {
	return `
data "aws_ec2_coip_pools" "test" {}

resource "aws_eip" "test" {
  customer_owned_ipv4_pool = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
  vpc                      = true
}

data "aws_eip" "test" {
  id = aws_eip.test.id
}
`
}

func testAccEIPFilterDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %q
  }
}

data "aws_eip" "test" {
  filter {
    name   = "tag:Name"
    values = [aws_eip.test.tags.Name]
  }
}
`, rName)
}

const testAccEIPIDDataSourceConfig = `
resource "aws_eip" "test" {
  vpc = true
}

data "aws_eip" "test" {
  id = aws_eip.test.id
}
`

func testAccEIPPublicIPClassicDataSourceConfig() string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		`
resource "aws_eip" "test" {}

data "aws_eip" "test" {
  public_ip = aws_eip.test.public_ip
}
`)
}

const testAccEIPPublicIPVPCDataSourceConfig = `
resource "aws_eip" "test" {
  vpc = true
}

data "aws_eip" "test" {
  public_ip = aws_eip.test.public_ip
}
`

func testAccEIPTagsDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %q
  }
}

data "aws_eip" "test" {
  tags = {
    Name = aws_eip.test.tags["Name"]
  }
}
`, rName)
}

const testAccEIPNetworkInterfaceDataSourceConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_eip" "test" {
  vpc               = true
  network_interface = aws_network_interface.test.id
}

data "aws_eip" "test" {
  filter {
    name   = "network-interface-id"
    values = [aws_eip.test.network_interface]
  }
}
`

var testAccEIPInstanceDataSourceConfig = acctest.ConfigCompose(
	acctest.ConfigAvailableAZsNoOptInDefaultExclude(), `
resource "aws_vpc" "test" {
  cidr_block = "10.2.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.2.0.0/24"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

data "aws_ami" "test" {
  most_recent = true
  name_regex  = "^amzn-ami.*ecs-optimized$"

  owners = [
    "amazon",
  ]
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.test.id
  subnet_id     = aws_subnet.test.id
  instance_type = "t2.micro"
}

resource "aws_eip" "test" {
  vpc      = true
  instance = aws_instance.test.id
}

data "aws_eip" "test" {
  filter {
    name   = "instance-id"
    values = [aws_eip.test.instance]
  }
}
`)

func testAccEIPCarrierIPDataSourceConfig(rName string) string {
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

data "aws_eip" "test" {
  id = aws_eip.test.id
}
`, rName))
}
