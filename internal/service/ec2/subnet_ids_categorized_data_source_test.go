package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2SubnetIDsCategorizedDataSource_basic(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetIDsCategorizedDataSourceConfig(rName, rInt),
			},
			{
				Config: testAccSubnetIDsWithDataSourceCategorizedDataSourceConfig(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "public_subnet_ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "private_subnet_ids.#", "3"),
				),
			},
		},
	})
}

func TestAccEC2SubnetIDsCategorizedDataSourceWithExplicitRoute_ids(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetIDsCategorizedDataSourceWithExplicitPublicRoute_ids(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "public_subnet_ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "private_subnet_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccEC2SubnetIDsCategorizedDataSourceWithImplicitRoute_ids(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetIDsCategorizedDataSourceWithImplicitPublicRoute_ids(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "public_subnet_ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "private_subnet_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccEC2SubnetIDsCategorizedDataSourceWithNoGateway_ids(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetIDsCategorizedDataSourceWithNoGateway_ids(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "public_subnet_ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "private_subnet_ids.#", "3"),
				),
			},
		},
	})
}

func testAccSubnetIDsWithDataSourceCategorizedDataSourceConfig(rName string, rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%[2]d.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_public_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_private_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.125.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_private_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.126.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

data "aws_subnet_ids_categorized" "categorized" {
  vpc_id = aws_vpc.test.id
}
`, rName, rInt))
}

func testAccSubnetIDsCategorizedDataSourceConfig(rName string, rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%[2]d.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_public_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_private_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.125.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_private_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.126.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}
`, rName, rInt))
}

// Here the route to the IGW is declared in a specific route table created by the configuration.
// Public subnets are determined as being directly associated with this route table
func testAccSubnetIDsCategorizedDataSourceWithExplicitPublicRoute_ids(rName string, rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%[2]d.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "test"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }
}

# Make subnet test_a_one public
resource "aws_route_table_association" "a_one" {
  subnet_id      = aws_subnet.test_a_one.id
  route_table_id = aws_route_table.public.id
}

resource "aws_subnet" "test_a_one" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_a_two" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

data "aws_subnet_ids_categorized" "categorized" {
  vpc_id = aws_vpc.test.id

  # Ensure data source not read till all referenced resources exist
  depends_on = [
    aws_route_table_association.a_one
  ]
}
`, rName, rInt))
}

// Here the route to the IGW is added to the main route table.
// Public subnets are therefore those _not_ associated with any other route table.
func testAccSubnetIDsCategorizedDataSourceWithImplicitPublicRoute_ids(rName string, rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%[2]d.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_route_table" "main" {
  vpc_id = aws_vpc.test.id

  filter {
    name = "association.main"
    values = [
      "true"
    ]
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "test"
  }
}

resource "aws_route" "public" {
  route_table_id = data.aws_route_table.main.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = aws_internet_gateway.gw.id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.test.id
}

# Make subnet test_a_one private
resource "aws_route_table_association" "a_one" {
  subnet_id      = aws_subnet.test_a_one.id
  route_table_id = aws_route_table.private.id
}

resource "aws_subnet" "test_a_one" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_a_two" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

data "aws_subnet_ids_categorized" "categorized" {
  vpc_id = aws_vpc.test.id

  # Ensure data source not read till all referenced resources exist
  depends_on = [
    aws_route_table_association.a_one
  ]
}

`, rName, rInt))
}

// Here these is no gateway, thus all subnets are private including those on the main route table
func testAccSubnetIDsCategorizedDataSourceWithNoGateway_ids(rName string, rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%[2]d.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.test.id
}

# Take subnet test_a_one off the main route table
resource "aws_route_table_association" "a_one" {
  subnet_id      = aws_subnet.test_a_one.id
  route_table_id = aws_route_table.private.id
}

resource "aws_subnet" "test_a_one" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_a_two" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

data "aws_subnet_ids_categorized" "categorized" {
  vpc_id = aws_vpc.test.id

  # Ensure data source not read till all referenced resources exist
  depends_on = [
    aws_subnet.test_a_one,
    aws_subnet.test_a_two,
    aws_subnet.test_b,
  ]
}

`, rName, rInt))
}
