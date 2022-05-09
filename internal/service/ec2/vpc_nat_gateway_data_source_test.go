package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCNATGatewayDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceNameById := "data.aws_nat_gateway.test_by_id"
	dataSourceNameBySubnetId := "data.aws_nat_gateway.test_by_subnet_id"
	dataSourceNameByTags := "data.aws_nat_gateway.test_by_tags"
	resourceName := "aws_nat_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNATGatewayDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceNameById, "connectivity_type", resourceName, "connectivity_type"),
					resource.TestCheckResourceAttrPair(dataSourceNameById, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetId, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrSet(dataSourceNameById, "state"),
					resource.TestCheckResourceAttrSet(dataSourceNameById, "allocation_id"),
					resource.TestCheckResourceAttrSet(dataSourceNameById, "network_interface_id"),
					resource.TestCheckResourceAttrSet(dataSourceNameById, "public_ip"),
					resource.TestCheckResourceAttrSet(dataSourceNameById, "private_ip"),
					resource.TestCheckNoResourceAttr(dataSourceNameById, "attached_vpc_id"),
					resource.TestCheckResourceAttrSet(dataSourceNameById, "tags.OtherTag"),
				),
			},
		},
	})
}

func testAccNATGatewayDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.5.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.5.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

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

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  subnet_id     = aws_subnet.test.id
  allocation_id = aws_eip.test.id

  tags = {
    Name     = %[1]q
    OtherTag = "some-value"
  }

  depends_on = [aws_internet_gateway.test]
}

data "aws_nat_gateway" "test_by_id" {
  id = aws_nat_gateway.test.id
}

data "aws_nat_gateway" "test_by_subnet_id" {
  subnet_id = aws_nat_gateway.test.subnet_id
}

data "aws_nat_gateway" "test_by_tags" {
  tags = {
    Name = aws_nat_gateway.test.tags["Name"]
  }
}
`, rName))
}
