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
				Config: testAccSubnetIDsCategorizedDataSourceWithExplicitRoute_ids(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					//resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "public_subnet_ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_subnet_ids_categorized.categorized", "private_subnet_ids.#", "2"),
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

func testAccSubnetIDsCategorizedDataSourceWithExplicitRoute_ids(rName string, rInt int) string {
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
  
resource "aws_route_table_association" "a" {
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

data "aws_subnet_ids" "test" {
  vpc_id = aws_subnet.test_a_two.vpc_id

  filter {
    name   = "availabilityZone"
    values = [aws_subnet.test_a_one.availability_zone]
  }
}

data "aws_subnet_ids_categorized" "categorized" {
    vpc_id = aws_vpc.test.id
  }
  
`, rName, rInt))
}
