package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCSubnetIDsDataSource_basic(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetIDsDataSourceConfig_basic(rName, rInt),
			},
			{
				Config: testAccVPCSubnetIDsDataSourceConfig_dataSource(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnet_ids.selected", "ids.#", "3"),
					resource.TestCheckResourceAttr("data.aws_subnet_ids.private", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccVPCSubnetIDsDataSource_filter(t *testing.T) {
	rInt := sdkacctest.RandIntRange(0, 256)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetIdsDataSourceConfig_filter(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnet_ids.test", "ids.#", "2"),
				),
			},
		},
	})
}

func testAccVPCSubnetIDsDataSourceConfig_dataSource(rName string, rInt int) string {
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
    Tier = "Public"
  }
}

resource "aws_subnet" "test_private_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.125.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
    Tier = "Private"
  }
}

resource "aws_subnet" "test_private_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.126.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
    Tier = "Private"
  }
}

data "aws_subnet_ids" "selected" {
  vpc_id = aws_vpc.test.id
}

data "aws_subnet_ids" "private" {
  vpc_id = aws_vpc.test.id

  tags = {
    Tier = "Private"
  }
}
`, rName, rInt))
}

func testAccVPCSubnetIDsDataSourceConfig_basic(rName string, rInt int) string {
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
    Tier = "Public"
  }
}

resource "aws_subnet" "test_private_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.125.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
    Tier = "Private"
  }
}

resource "aws_subnet" "test_private_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[2]d.126.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
    Tier = "Private"
  }
}
`, rName, rInt))
}

func testAccVPCSubnetIdsDataSourceConfig_filter(rName string, rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%[2]d.0.0/16"

  tags = {
    Name = %[1]q
  }
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
`, rName, rInt))
}
