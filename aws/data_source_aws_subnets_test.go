package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataSourceAwsSubnets_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSubnetsConfig(rName),
			},
			{
				Config: testAccDataSourceAwsSubnetsConfigWithDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnets.selected", "ids.#", "4"),
					resource.TestCheckResourceAttr("data.aws_subnets.private", "ids.#", "2"),
					testCheckResourceAttrGreaterThanValue("data.aws_subnets.all", "ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_subnets.none", "ids.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSubnets_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSubnets_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnets.test", "ids.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSubnetsConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_public_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-public-a"
    Tier = "Public"
  }
}

resource "aws_subnet" "test_public_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.124.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-public-b"
    Tier = "Public"
  }
}

resource "aws_subnet" "test_private_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.125.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-private-a"
    Tier = "Private"
  }
}

resource "aws_subnet" "test_private_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.126.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s-private-b"
    Tier = "Private"
  }
}
`, rName))
}

func testAccDataSourceAwsSubnetsConfigWithDataSource(rName string) string {
	return acctest.ConfigCompose(testAccDataSourceAwsSubnetsConfig(rName), `
data "aws_subnets" "selected" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }

  tags = {
    Tier = "Private"
  }
}

data "aws_subnets" "all" {}

data "aws_subnets" "none" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }

  filter {
    name   = "cidr-block"
    values = ["172.16.127.0/24"]
  }
}
`)
}

func testAccDataSourceAwsSubnets_filter(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_a_one" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-a1"
  }
}

resource "aws_subnet" "test_a_two" {
  vpc_id            = aws_subnet.test_b.vpc_id
  cidr_block        = "172.16.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-a2"
  }
}

resource "aws_subnet" "test_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.16.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s-b"
  }
}

data "aws_subnets" "test" {
  filter {
    name   = "availabilityZone"
    values = [aws_subnet.test_a_one.availability_zone]
  }

  filter {
    name   = "vpc-id"
    values = [aws_subnet.test_a_two.vpc_id]
  }
}
`, rName))
}
